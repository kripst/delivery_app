package kafka_consumer

import (
	"context"
	"log"

	"github.com/IBM/sarama"
	"github.com/kripst/delivery_service/config"
	"go.uber.org/zap"
)

type Consumer interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
	GetMessages() <- chan *sarama.ConsumerMessage
}

type ConsumerImpl struct {
	ready      chan bool
	messagePool chan *sarama.ConsumerMessage
	logger     *zap.Logger
}

func NewOrdersConsumer(ctx context.Context, config *config.KafkaConsumerConfig) (*ConsumerImpl, error) {
	zapCfg := zap.NewDevelopmentConfig()

	zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapCfg.OutputPaths = []string{
		// включить чтобы в папку записывались "logs/order_service.logs",
		// os.Getenv("LOGSPATH"),
		"stdout",
	}

	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("Не удалось проинициализировать логгер зап: %v", err)
	}

	consumer := &ConsumerImpl{
		ready:      make(chan bool),
		messagePool: make(chan *sarama.ConsumerMessage, config.MaxWorkerPool),
		logger:     logger,
	}

	client, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID[0], config.SaramaConfig)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			if err := client.Consume(ctx, []string{config.Topic}, consumer); err != nil {
				consumer.logger.Fatal("Ошибка от консьемера", zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	consumer.logger.Info("Sarama consumer запущен и работает")

	return consumer, nil
}

// Setup вызывается при начале новой сессии консьюмера.
func (c *ConsumerImpl) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup вызывается при завершении сессии консьюмера.
func (c *ConsumerImpl) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim читает сообщения из партиции и отправляет их в пул воркеров.
func (c *ConsumerImpl) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		c.logger.Info("Получено сообщение", 
			zap.String("topic",message.Topic), 
			zap.Int32("partition", message.Partition), 
			zap.Int64("offset", message.Offset))
		// Отправляем сообщение в пул воркеров
		c.messagePool <- message
		session.MarkMessage(message, "")

	}

	return nil
}

func (c *ConsumerImpl) GetMessages() <-chan *sarama.ConsumerMessage {
    return c.messagePool
}