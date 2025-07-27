package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/kripst/delivery_service/internal/kafka"
	"github.com/kripst/delivery_service/internal/storage/postgres"
	"github.com/kripst/delivery_service/internal/storage/redis"
	"github.com/kripst/delivery_service/model"
	"go.uber.org/zap"
)



type DeliveryServer struct {
	Consumer kafka.Consumer
	Db       postgres.DeliveryRepository
	Logger   *zap.Logger
	Tm       redis.TimeManager
}

func (d *DeliveryServer) HandleNewOrders() {
	var partition int32 = 0

	for {
		data, err := d.Consumer.ListenTopic(model.Orders_topic, partition)

		if err != nil {
			d.Logger.Error("consumer error",
				zap.Error(err),
				zap.String("topic", model.Orders_topic),
				zap.Int32("partition", partition))
				continue
		}

		deliveryData := &model.Delivery{}
		if err := json.Unmarshal(data, deliveryData); err != nil {
			// dead letter queue сделать для ошибок непонятных НАДО ПОМЕНЯТЬ 
			d.Logger.Error("json Unmarshal error",
				zap.Error(err),
				zap.String("topic", model.Orders_topic),
				zap.Int32("partition", partition),
				zap.Any("data", data))
			continue
		}

		d.Logger.Debug("parsing json data to struct", zap.Any("data", deliveryData))
		
		if err := d.Db.CreateDelivery(context.Background(), deliveryData); err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}

			d.Logger.Error("could not create delivery",
				zap.Error(err),
				zap.String("topic", model.Orders_topic),
				zap.Int32("partition", partition),
				zap.Any("data", data))
				continue
		}

		if err := d.Tm.ScheduleOrder(context.Background(), deliveryData.OrderID, deliveryData.DeliveryWindow); err != nil {
			d.Logger.Error("could not create delivery in tm",
				zap.Error(err),
				zap.String("topic", model.Orders_topic),
				zap.Int32("partition", partition),
				zap.Any("order ID", deliveryData.OrderID),
				zap.String("delivery window", deliveryData.DeliveryWindow))
				continue
		}
	}
}
