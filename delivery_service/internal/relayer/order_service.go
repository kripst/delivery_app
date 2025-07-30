package relayer

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/kripst/delivery_service/config"
	"github.com/kripst/delivery_service/model"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type deliveryService interface {
	GetReadyOrders(ctx context.Context) ([]string, error)
}

type deliveryServiceImpl struct {
	Client *redis.Client
	Logger *zap.Logger
}

func NewdeliveryServiceImpl(config *config.RedisConfig) (*deliveryServiceImpl, error) {
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

	return &deliveryServiceImpl{Client: client, Logger: logger}, nil

}

// возвращает [] заказов , чье время настало
func (t *deliveryServiceImpl) GetReadyOrders(ctx context.Context) ([]string, error) {
	deliveryOrdersKey := model.DeliveryOrderKey

	nowUnix := time.Now().Unix()

	opt := &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(nowUnix, 10),
			// Добавляем эти параметры, чтобы получить ВСЕ записи
		Offset: 0,
		Count:  -1,
	}

	readyOrders, err := t.Client.ZRangeByScore(ctx, deliveryOrdersKey, opt).Result()
	if err != nil {
		t.Logger.Error("could not get orders",
		zap.Error(err))
	}

	var result []string

	for _, orderID := range readyOrders {
		result = append(result, orderID)
		// НАДО ПОМЕНЯТЬ чтобы удалялись данные , только когда успешно получены в другом сервисе
		removedCnt , err := t.Client.ZRem(ctx, deliveryOrdersKey, orderID).Result()
		if err != nil {
			t.Logger.Error("could not delete orders",
			zap.Error(err))
		} else if removedCnt > 0 {
			t.Logger.Debug("successfully deleted order from redis queue", zap.String("order ID", orderID))
		}
	}

	if len(result) > 0 {
		t.Logger.Info("new orders in redis", zap.Any("orderIDs", result))
	}
	
	return result, nil

}

