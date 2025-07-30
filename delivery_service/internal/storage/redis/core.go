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
	ScheduleOrders(ctx context.Context, orderID []string, deliveryWindow []string) error
	
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
// deliveryTime - строка в формате "YYYY:MM:DD:HH:MM"
func (t *TimeManagerImpl) ScheduleOrders(ctx context.Context, orderIDs []string, deliveryWindows []string) error {
	result := make([]redis.Z, len(orderIDs))
	for i := 0; i < len(orderIDs); i++ {
		parts := strings.Split(deliveryWindows[i], ":")
		if len(parts) != 5 {
			return fmt.Errorf("invalid time format, expected YYYY:MM:DD:HH:MM")
		}
		// НАДО ПОМЕНЯТЬ - пусть вставляет данные уже в формате "2006-01-02 15:04" потом добавлять + ":00"
		// Создаем строку в формате, который понимает time.Parse
		timeStr := fmt.Sprintf("%s-%s-%s %s:%s:00", parts[0], parts[1], parts[2], parts[3], parts[4])
		
		// Парсим с учетом временной зоны (можно указать time.UTC или Local)
		deliveryTimestamp, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			return fmt.Errorf("failed to parse time: %v", err)
		}

		result = append(result, redis.Z{
			Score: float64(deliveryTimestamp.Unix()),
			Member: orderIDs[i],
		})
	}
	deliveryOrdersKey := model.DeliveryOrderKey

	err := t.Client.ZAdd(ctx, deliveryOrdersKey, result...).Err()

	if err != nil {
		return fmt.Errorf("не удалось добавить заказы в Redis: %v", err)
	}

	for i := 0; i < len(orderIDs); i++ {
		deliveryTimestamp := result[i].Score
		// Преобразование обратно в time.Time
		ti := time.Unix(int64(deliveryTimestamp), 0) // наносекунды = 0

		// Форматирование в "2006-01-02 15:04:05"
		formattedTime := ti.Format("2006-01-02 15:04:05")

		t.Logger.Info("new redis order",
		zap.String("order ID", orderIDs[i]),
		zap.Any("order time ", formattedTime))
	}
	
	return nil
}

