package relayer

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kripst/delivery_service/model"
	"go.uber.org/zap"
)

type outboxService interface {
	updateDelivery(ctx context.Context, orderIDs []string) (error)
}

type outboxServiceImpl struct {
	db *pgxpool.Pool
	logger *zap.Logger
}

func NewoutboxServiceImpl(db *pgxpool.Pool, logger *zap.Logger) *outboxServiceImpl {
	return &outboxServiceImpl{
		db: db,
		logger: logger,
	} 
	
}

func (s *outboxServiceImpl) updateDelivery(ctx context.Context, orderIDs []string) (error) {
	batch := &pgx.Batch{}

	for _, orderID := range orderIDs {
		query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2",
			model.DeliveryOutboxTable, model.FieldDeliveryStatus, model.FieldDeliveryID)

		batch.Queue(query,model.DeliveryStatusPending, orderID)
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

	s.logger.Info("successfully updated delivery status",
		zap.Any("orderIDs", orderIDs))	
	return nil
}