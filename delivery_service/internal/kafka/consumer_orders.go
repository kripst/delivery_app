package kafka

import (
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/kripst/delivery_service/config"
	"go.uber.org/zap"
)

type Consumer interface {
	ListenTopic(topic string, partition int32) ([]byte, error)
}

type OrdersConsumer struct {
	Consumer sarama.Consumer
	Log *zap.Logger
	lastOffset int64
    offsetMu   sync.Mutex
}


func NewOrdersConsumer(config *config.KafkaConsumerConfig) (*OrdersConsumer, error) {
	consumer , err := sarama.NewConsumer(config.Brokers, config.SaramaConfig)
	if err != nil {
		return  nil, err
	}

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



	return &OrdersConsumer{Consumer: consumer, Log: logger, lastOffset: sarama.OffsetOldest}, nil
}

func (o *OrdersConsumer) ListenTopic(topic string, partition int32) ([]byte, error) {
	o.offsetMu.Lock()
    offset := o.lastOffset + 1
    o.offsetMu.Unlock()

	worker , err := o.Consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return  nil, err
	}

	defer worker.Close()

	select {
	case err := <-worker.Errors():
		return  nil, err
	case msg := <-worker.Messages():
		o.Log.Info("new consumer message",
			zap.String("topic", topic),
			zap.Int32("partition",partition),
			zap.Int64("offset",offset),
			zap.String("Value", string(msg.Value)))

		o.offsetMu.Lock()
		o.lastOffset = msg.Offset
		o.offsetMu.Unlock()
		return msg.Value, nil
	}
	
}

