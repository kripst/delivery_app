package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kripst/delivery_service/model"
	"go.uber.org/zap"
	"log"
)

type DeliveryRepository interface {
	BatchInsert(ctx context.Context, delivery []*model.Delivery) error
	// CancelDelivery(orderID string) error
}

type PostgresRepository struct {
	Db  *pgxpool.Pool
	Log *zap.Logger
}

func NewPostgresRepository(ctx context.Context, config *pgxpool.Config) (*PostgresRepository, error) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logCfg.OutputPaths = []string{
		// включить чтобы в папку записывались "logs/order_service.logs",
		// os.Getenv("LOGSPATH"),
		"stdout",
	}

	logger, err := logCfg.Build()
	if err != nil {
		log.Fatalf("Не удалось проинициализировать логгер зап: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping pgx pool: %w", err)
	}

	logger.Info("Postgres successfully connected",
		zap.String("connection string", config.ConnString()),
		zap.Int("max_connections", int(config.MaxConns)),
	)

	return &PostgresRepository{Db: pool, Log: logger}, nil
}

func (r *PostgresRepository) BatchInsert(ctx context.Context, deliveries []*model.Delivery) error {
	orderIDs := make([]string, 0, len(deliveries))

	tx, err := r.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		r.Log.Error("Failed to begin PGX transaction", zap.Error(err))
		return fmt.Errorf("transaction begin failed: %w", err)
	}

	// Откатываем транзакцию при ошибке (defer сработает при выходе из функции)
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				r.Log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()

	batch := &pgx.Batch{}
	
	for i := 0; i < len(deliveries); i++ {
		insertBatch(batch, deliveries[i])
		orderIDs = append(orderIDs, deliveries[i].OrderID)
	}
	
	br := tx.SendBatch(ctx, batch)

	for i := 0; i < len(deliveries); i++ {
		if _, err := br.Exec(); err != nil {
			r.Log.Error("could not create order", 
				zap.String("order_id", deliveries[i].OrderID),
				zap.Any("order data", deliveries[i]),
				zap.Error(err))
		}
	}

	if err := br.Close(); err != nil {
		r.Log.Error("could not close batch", 
				zap.Error(err))
	}

	// Фиксируем транзакцию
	if err = tx.Commit(ctx); err != nil {
		r.Log.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	r.Log.Info("Successfully created new orders", zap.Any("orderIDs", orderIDs))

	return nil
}

func insertBatch(batch *pgx.Batch, delivery *model.Delivery) {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s,%s, %s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		DeliveriesTable,
		FieldOrderID,
		FieldUserID,
		FieldDarkstoreID,
		FieldDeliveryWindow,
		FieldDeliveryAddress,
		FieldCommentToCourier,
		FieldUnderDoor,
		FieldCallCourier,
		FieldUserName,
		FieldUserSurname,
		FieldUserPhone,
	)
		
	batch.Queue(
		query,
		delivery.OrderID,
		delivery.UserID,
		delivery.DarkstoreID,
		delivery.DeliveryWindow,
		delivery.Address,
		delivery.CommentToCourier,
		delivery.UnderDoor,
		delivery.CallBefore,
		delivery.UserName,
		delivery.UserSurname,
		delivery.UserPhone,
	)

	for _, item := range delivery.DeliveryItems {
		itemQuery := fmt.Sprintf(
			"INSERT INTO %s (%s, %s, %s, %s) VALUES ($1, $2, $3, $4)",
			DeliveryItemsTable,
			FieldID,
			FieldDeliveryID,
			FieldProductID,
			FieldQuantity,
		)

		batch.Queue(
			itemQuery,
			uuid.New().String(),
			delivery.OrderID,
			item.ProductID,
			item.Quantity,
		)
	}

	// outbox
	outBoxQueue := fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES ($1, $2)",
		DeliveryOutboxTable,
		FieldDeliveryID,
		FieldDeliveryWindow)

	batch.Queue(outBoxQueue, delivery.OrderID, delivery.DeliveryWindow)
}