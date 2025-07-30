package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/kripst/delivery_service/internal/kafka_consumer"
	"github.com/kripst/delivery_service/internal/storage/postgres"
	"github.com/kripst/delivery_service/internal/storage/redis"
	"github.com/kripst/delivery_service/model"
	"go.uber.org/zap"
)



type DeliveryServer struct {
	Consumer kafka_consumer.Consumer
	Db       postgres.DeliveryRepository
	Logger   *zap.Logger
	Tm       redis.TimeManager
	Ctx      context.Context
}

func NewDeliveryServer(consumer kafka_consumer.Consumer, db postgres.DeliveryRepository, logger *zap.Logger, tm redis.TimeManager, ctx context.Context) *DeliveryServer {
	return &DeliveryServer{
		Consumer: consumer,
		Db: db,
		Logger: logger,
		Tm: tm,
		Ctx: ctx,
	}
}

func (d *DeliveryServer) processOrders(batch [][]byte) {
	deliveries := make([]*model.Delivery, 0 , len(batch))
	deliveriesWindows   := make([]string, 0 , len(batch))
	deliveriesOrderIDs := make([]string, 0 , len(batch))

	for _, data := range batch {
		delivery := &model.Delivery{}

		if err := json.Unmarshal(data, delivery); err != nil {
			d.Logger.Fatal("invalid input", zap.Error(err), zap.String("data", string(data)))
		}
		deliveries = append(deliveries, delivery)
		deliveriesWindows   = append(deliveriesWindows, delivery.DeliveryWindow)
		deliveriesOrderIDs = append(deliveriesOrderIDs, delivery.OrderID)
	}

	if err := d.Db.BatchInsert(d.Ctx, deliveries); err != nil {
		d.Logger.Fatal("invalid input", zap.Error(err), zap.Any("data", deliveries))
	}

	if err := d.Tm.ScheduleOrders(d.Ctx, deliveriesOrderIDs, deliveriesWindows); err != nil {
		d.Logger.Fatal("invalid input", zap.Error(err), zap.Any("deliveriesOrderIDs", deliveriesOrderIDs),
			zap.Any("deliveriesWindows", deliveriesWindows))
	}
}

func (d *DeliveryServer) HandleNewOrders() {
    const (
        maxWorkerPool = 6
        // Размер буфера канала. Подбирается экспериментально.
        // Должен быть больше, чем batchSize * maxWorkerPool
        jobsChannelBuffer = 100
    )

    // Канал для передачи сообщений от "читателя" к "воркерам"
    jobs := make(chan []byte, jobsChannelBuffer)
    var wg sync.WaitGroup

    // --- Запускаем воркеров (Consumers) ---
    // Они будут жить постоянно и ждать заданий в канале jobs
    for i := 0; i < maxWorkerPool; i++ {
        wg.Add(1)
        go d.worker(&wg, jobs)
    }

    // --- Запускаем читателя из Kafka (Producer) ---
    // Он тоже работает постоянно, читает из Kafka и кидает в канал jobs
    go d.kafkaReader(jobs)

    // --- Ожидание сигнала завершения ---
    <-d.Ctx.Done()
    close(jobs) // Закрываем канал, чтобы воркеры поняли, что заданий больше не будет
    wg.Wait()   // Ждем, пока все воркеры завершат обработку оставшихся сообщений
}


// worker - это один из ваших 6 обработчиков.
// Он сам накапливает пакеты.
func (d *DeliveryServer) worker(wg *sync.WaitGroup, jobs <-chan []byte) {
    defer wg.Done()

    const (
        batchSize = 10
        // Таймаут, чтобы не ждать вечно, если сообщения идут редко
        batchTimeout = 2 * time.Second
    )

    batch := make([][]byte, 0, batchSize)
    ticker := time.NewTicker(batchTimeout)
    defer ticker.Stop()

    for {
        select {
        case msg, ok := <-jobs:
            if !ok {
                // Канал jobs закрыт. Отправляем остатки и выходим.
                if len(batch) > 0 {
                    d.processOrders(batch)
                }
                return
            }

            batch = append(batch, msg)
            if len(batch) >= batchSize {
                // Пакет набрался по размеру, отправляем
                d.processOrders(batch)
                batch = batch[:0] // Очищаем
            }

        case <-ticker.C:
            // Сработал таймаут. Отправляем то, что успели накопить
            if len(batch) > 0 {
                d.processOrders(batch)
                batch = batch[:0] // Очищаем
            }
        }
    }
}

// kafkaReader просто читает из Kafka и перекладывает в канал
func (d *DeliveryServer) kafkaReader(jobs chan<- []byte) {
    // Предполагается, что GetMessages возвращает канал для чтения
    messagesChannel := d.Consumer.GetMessages()
    for msg := range messagesChannel {
        // Здесь можно добавить обработку d.Ctx.Done(), чтобы прекратить чтение
        jobs <- msg.Value
    }
}