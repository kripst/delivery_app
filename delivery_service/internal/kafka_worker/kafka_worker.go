package kafka_worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kripst/delivery_service/internal/storage/postgres"
	"github.com/kripst/delivery_service/model"
	"go.uber.org/zap"
)

type Kafka_worker struct {
	producerOutbox ProducerOutbox
	outboxService  postgres.OutboxService
	logger        *zap.Logger
	
}

func NewKafkaWorker(producerOutbox ProducerOutbox, outboxService postgres.OutboxService, logger *zap.Logger) *Kafka_worker{
	return &Kafka_worker{
		producerOutbox: producerOutbox,
		outboxService: outboxService,
		logger: logger,
	}
}

func (w *Kafka_worker) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Kafka worker was cancelled")
			return
		case <-ticker.C:
			w.produceIvents(ctx)
		}
	}
}

func (w *Kafka_worker) produceIvents(ctx context.Context) {
	datas , err := w.outboxService.GetData(ctx)
	var orderIDs []string
	if len(datas) == 0{
		w.logger.Info("Kafka_worker: no new orders found")
		return
	}

	if err != nil {
		w.logger.Error("get data error", zap.Error(err))
		return
	}

	var jsonDatas [][]byte
	for _, data := range datas {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			w.logger.Error("invalid data input", zap.Error(err))
			return
		}
		orderIDs  = append(orderIDs, data.OrderID)
		jsonDatas = append(jsonDatas, jsonData)
	}

	
	if err := w.producerOutbox.SendToTopic(model.Pending_topic, jsonDatas); err != nil {
		w.logger.Error("invalid send to topic", zap.Error(err))
		return
	}

	// обновить статус заказ с пендинг на sent
	if err := w.outboxService.UpdateDelivery(ctx, orderIDs, model.DeliveryStatusSENT); err != nil {
		w.logger.Error("could not update delivery status", zap.Error(err))
	}
}