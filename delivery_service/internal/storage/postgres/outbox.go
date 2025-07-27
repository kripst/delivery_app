package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kripst/delivery_service/model"
	"go.uber.org/zap"
)

type OutboxService interface {
	UpdateDelivery(ctx context.Context, orderIDs []string, newUpdate string) error
	GetData(ctx context.Context) ([]*model.DeliveryInput, error)
}

type OutboxServiceImpl struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewoutboxServiceImpl(db *pgxpool.Pool, logger *zap.Logger) *OutboxServiceImpl {
	return &OutboxServiceImpl{
		db:     db,
		logger: logger,
	}

}

func (s *OutboxServiceImpl) UpdateDelivery(ctx context.Context, orderIDs []string, newUpdate string) error {
	batch := &pgx.Batch{}
	for _, orderID := range orderIDs {
		query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2",
			model.DeliveryOutboxTable, model.FieldDeliveryStatus, model.FieldDeliveryID)

		batch.Queue(query, newUpdate, orderID)
	}

	// Отправляем все запросы разом
	br := s.db.SendBatch(ctx, batch)
	// Проверяем ошибки для каждой операции в batch
	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			s.logger.Error("Failed to create order item in batch",
				zap.Error(err),
				zap.String("order_id", orderIDs[i]),
				zap.Int("item_index", i),
			)
			return fmt.Errorf("batch insert failed at item %d: %w", i, err)
		}
	}

	
	if err := br.Close(); err != nil {
		s.logger.Error("Failed to close batch",
			zap.Error(err),
			zap.Any("orderIDs", orderIDs),
		)
		return err
	}

	if len(orderIDs) > 0 {
		s.logger.Info("successfully updated delivery status",
			zap.Any("orderIDs", orderIDs))
	}
	
	return nil
}

func (s *OutboxServiceImpl) GetData(ctx context.Context) ([]*model.DeliveryInput, error) {
	var result []*model.DeliveryInput

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", model.FieldDeliveryID,
		model.DeliveryOutboxTable,
		model.FieldDeliveryStatus)

	rowsOrderID, err := s.db.Query(ctx, query, model.DeliveryStatusPending)
	if err != nil {
		return result, fmt.Errorf("get order failed: %w", err)
	}
	defer rowsOrderID.Close()
	
	// ищем по orderID данные у которых статус PENDING
	for rowsOrderID.Next() {
		var orderID string
		if err := rowsOrderID.Scan(&orderID); err != nil {
			return nil, err
		}

		query := fmt.Sprintf(`
			SELECT 
				d.%s, d.%s, d.%s, d.%s, d.%s, d.%s, d.%s, d.%s, d.%s, d.%s,
				i.%s, i.%s, i.%s
			FROM 
				%s d
			JOIN 
				%s i ON d.%s = i.%s
			WHERE 
				d.%s = $1`,  // или другое условие
			// Поля из deliveries
			FieldOrderID,
			FieldUserID,
			FieldDarkstoreID,
			FieldDeliveryAddress,
			FieldCommentToCourier,
			FieldUnderDoor,
			FieldCallCourier,
			FieldUserName,
			FieldUserSurname,
			FieldUserPhone,
			
			// Поля из delivery_items
			FieldID,
			FieldProductID,
			FieldQuantity,
			
			// Имена таблиц
			DeliveriesTable,
			DeliveryItemsTable,
			
			// Условия JOIN
			FieldOrderID,  // из deliveries
			FieldDeliveryID,  // из delivery_items
			
			// Условие WHERE
			FieldOrderID,
		)

		rowsDelivery, err := s.db.Query(ctx, query, orderID)
		if err != nil {
			return result, fmt.Errorf("get order failed: %w", err)
		}
		defer rowsDelivery.Close()
		delivery := &model.DeliveryInput{}
		var items []*model.DeliveryItem
		firstRow := true
		// добавляем 1 заказ по ордер id и все его item
		for rowsDelivery.Next() {
			item := &model.DeliveryItem{}
			if firstRow {
				err := rowsDelivery.Scan(
				&delivery.OrderID,
				&delivery.UserID,
				&delivery.DarkstoreID,
				&delivery.Address,
				&delivery.CommentToCourier,
				&delivery.UnderDoor,
				&delivery.CallBefore,
				&delivery.UserName,
				&delivery.UserSurname,
				&delivery.UserPhone,
				&item.ID,
				&item.ProductID,
				&item.Quantity,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan delivery: %w", err)
			}
			firstRow = false	 
			} else {
				err := rowsDelivery.Scan(
				nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
				&item.ID,
				&item.ProductID,
				&item.Quantity,)
				if err != nil {
					return nil, fmt.Errorf("failed to scan item: %w", err)
				}
			}

    		items = append(items, item)
		}

		delivery.DeliveryItems = items
		result = append(result, delivery)
	}


	if len(result) == 0 {
		return nil, nil
	}
	
	s.logger.Info("Successfully get data", zap.Int("len data", len(result)))

	return result, nil
}
