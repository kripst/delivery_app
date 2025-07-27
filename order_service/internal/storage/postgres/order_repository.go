package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kripst/order_service/model"
	"go.uber.org/zap"
)

type OrderRepositoty interface {
	CreateOrder(ctx context.Context, order *model.Order) (string, error)
	CancelOrder(ctx context.Context, orderId string) (error)
}
 
type PostgresRepository struct {
	Db *pgxpool.Pool
}

func NewPostgresRepository(ctx context.Context, config *pgxpool.Config) (*PostgresRepository, error) {
	// Создаем пул
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping pgx pool: %w", err)
	}

	zap.L().Info("Postgres successfully connected", 
		zap.String("connection string", config.ConnString()),
		zap.Int("max_connections", int(config.MaxConns)),
	)
	return &PostgresRepository{Db: pool}, nil
}

// Закрытие соединения при завершении
func (r *PostgresRepository) Close() {
	r.Db.Close()
}

func (r *PostgresRepository) CancelOrder(ctx context.Context, orderID string) error {
    query := fmt.Sprintf(
        "UPDATE %s SET %s = $1 WHERE %s = $2", 
        OrderTable, 
        FieldOrderStatus, 
        FieldOrderID,
    )

    result, err := r.Db.Exec(ctx, query, StatusCancelled, orderID)
    if err != nil {
        zap.L().Error("could not UPDATE order status",
            zap.String("query", query),
            zap.String("order_id", orderID),
            zap.Error(err))
        return fmt.Errorf("failed to cancel order: %w", err)
    }

    // Проверяем, была ли обновлена хотя бы одна запись
    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        zap.L().Warn("no order found to cancel",
            zap.String("order_id", orderID))
        return fmt.Errorf("order not found")
    }

    zap.L().Info("order successfully cancelled",
        zap.String("order_id", orderID))
    return nil
}

func (r *PostgresRepository) CreateOrder(ctx context.Context, order *model.Order) (string, error) {
	orderID := order.ID
	// Начинаем транзакцию
	tx, err := r.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		zap.L().Error("Failed to begin PGX transaction", zap.Error(err))
		return "", fmt.Errorf("transaction begin failed: %w", err)
	}

	// Откатываем транзакцию при ошибке (defer сработает при выходе из функции)
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				zap.L().Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()

	// 1. Создаем заказ
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s) VALUES ($1, $2, $3, $4)",
		OrderTable,
		FieldOrderID,
		FieldUserID,
		FieldDarkstoreID,
		FieldTotalPrice,
	)

	_, err = tx.Exec(
		ctx,
		query,
		orderID,
		order.UserID,
		order.DarkstoreID,
		order.TotalPrice,
	)

	if err != nil {
		zap.L().Error("Could not create order in PGX transaction", zap.Error(err))
		return "", fmt.Errorf("order insert failed: %w", err)
	}

	// 2. Создаем элементы заказа (пакетная вставка)
	batch := &pgx.Batch{}

	for _, item := range order.Items {
		itemQuery := fmt.Sprintf(
			"INSERT INTO %s (%s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6)",
			OrderItemsTable,
			FieldOrderItemID,
			FieldOrderItemOrderID,
			FieldOrderItemProductID,
			FieldOrderItemProductName,
			FieldOrderItemQuantity,
			FieldOrderItemPrice,
		)

		batch.Queue(
			itemQuery,
			uuid.New().String(),
			orderID,
			item.ProductID,
			item.ProductName,
			item.Quantity,
			item.Price,
		)
	}

	// Отправляем все запросы разом
	br := tx.SendBatch(ctx, batch)

	// Проверяем ошибки для каждой операции в batch
	for i := 0; i < batch.Len(); i++ {
		_, err = br.Exec()
		if err != nil {
			zap.L().Error("Failed to create order item in batch", 
				zap.Error(err), 
				zap.String("order_id", orderID),
				zap.Int("item_index", i),
			)
			return "", fmt.Errorf("batch insert failed at item %d: %w", i, err)
		}
	}
	if err := br.Close(); err != nil {
		zap.L().Error("Failed to close batch", 
				zap.Error(err), 
				zap.String("order_id", orderID),
			)
			return "", err
	}

	// Фиксируем транзакцию
	if err = tx.Commit(ctx); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return "", fmt.Errorf("transaction commit failed: %w", err)
	}

	zap.L().Info("Successfully created new order", zap.String("order_id", orderID))
	return orderID, nil
}
