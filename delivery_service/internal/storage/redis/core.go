package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kripst/delivery_service/config"
	"github.com/kripst/delivery_service/model"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)


type TimeManager interface {
	ScheduleOrder(ctx context.Context, orderID string, deliveryWindow string) error
	
}

type TimeManagerImpl struct {
	Client   *redis.Client
	Logger *zap.Logger
}

func NewTimeManagerImpl(config *config.RedisConfig) (*TimeManagerImpl, error) {
	client := redis.NewClient(config.Opts)

	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logCfg.OutputPaths = []string{
		// включить чтобы в папку записывались "logs/order_service.logs",
		// os.Getenv("LOGSPATH"),
		"stdout",
	}

	logger, err  := logCfg.Build()
	if err != nil {
		log.Fatalf("Не удалось проинициализировать логгер зап: %v", err)
	}

	return &TimeManagerImpl{Client: client, Logger: logger}, nil

}

// scheduleOrder добавляет заказ в очередь на доставку.
// deliveryTime - строка в формате "13:00-14:00"
func (t *TimeManagerImpl) ScheduleOrder(ctx context.Context, orderID string, deliveryWindow string) error {
	deliveryOrdersKey := model.DeliveryOrderKey

	startTimeStr := strings.Split(deliveryWindow, "-")[0]
	now := time.Now()
	parsedTime, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		return fmt.Errorf("invalid %s :%v", deliveryWindow, err)
	}

	deliveryTimestamp := time.Date(
		now.Year(), now.Month(), now.Day(),
		parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location(),
	)

	err = t.Client.ZAdd(ctx, deliveryOrdersKey, redis.Z{
		Score: float64(deliveryTimestamp.Unix()),
		Member: orderID,
	}).Err()

	if err != nil {
		return fmt.Errorf("не удалось добавить заказ в Redis: %v", err)
	}

	t.Logger.Info("new redis order",
		zap.String("order ID", orderID),
		zap.Any("order time", deliveryTimestamp))
	
	return nil
}

