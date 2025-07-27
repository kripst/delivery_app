package kafka

import (
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Producer interface {
	SendToTopic(topic string, partition int32, jsonData []byte) error
}

type ProducerOutboxImpl struct {
	producer sarama.SyncProducer
	logger *zap.Logger
}

func NewProducer(producer *sarama.SyncProducer, logger *zap.Logger) *ProducerOutboxImpl{
	return &ProducerOutboxImpl{
		producer: *producer,
		logger: logger,
	}
}

func (p *ProducerOutboxImpl) SendToTopic(topic string, partition int32, jsonData []byte) error {
	
}