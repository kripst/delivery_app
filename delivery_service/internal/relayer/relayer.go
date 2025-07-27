package relayer

import (
	"context"
	"time"

	"github.com/kripst/delivery_service/internal/storage/postgres"
	"github.com/kripst/delivery_service/model"
	"go.uber.org/zap"
)

type Relayer struct {
	deliveryService deliveryService
	outboxServiceImpl postgres.OutboxService
	interval     time.Duration
	logger       *zap.Logger
}

func NewRelayer(service deliveryService, interval time.Duration, logger *zap.Logger, outboxServiceImpl postgres.OutboxService) (*Relayer) {
	return &Relayer{
		deliveryService: service,
		interval: interval,
		logger: logger,
		outboxServiceImpl: outboxServiceImpl, 
	}
}

func (r *Relayer) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)

	defer ticker.Stop()


	for {
		select {
		case <-ctx.Done():
			r.logger.Debug("Relayer stopped by context cancellation.")
			return
		case <-ticker.C:
			r.logger.Debug("Relayer: Checking for ready orders...")
			r.updateOrdersStatus(ctx)
		}
	}
}

func (r *Relayer) updateOrdersStatus(ctx context.Context) {
	orderIDs, err := r.deliveryService.GetReadyOrders(ctx)
	if err != nil {
		r.logger.Error("Relayer: Error getting ready orders",
			zap.Error(err))
		return
	}

	if len(orderIDs) == 0 {
		r.logger.Info("Relayer: No ready orders found.")
		return
	}

	err = r.outboxServiceImpl.UpdateDelivery(ctx, orderIDs, model.DeliveryStatusPending)
	if err != nil {
		r.logger.Error("Relayer: Error update ready deliveries",
			zap.Error(err))
		return
	}

	

}