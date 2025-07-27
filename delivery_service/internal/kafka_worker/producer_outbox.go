package kafka_worker

import (

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type ProducerOutbox interface {
	SendToTopic(topic string, jsonDatas [][]byte) error
}

type ProducerOutboxImpl struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
}

func NewOutboxProducer(producer sarama.SyncProducer, logger *zap.Logger) *ProducerOutboxImpl {
	return &ProducerOutboxImpl{
		producer: producer,
		logger:   logger,
	}
}

func (p *ProducerOutboxImpl) SendToTopic(topic string, jsonDatas [][]byte) error {
	var msgs []*sarama.ProducerMessage
	
	for _, jsonData := range jsonDatas {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(jsonData),
		}
		msgs = append(msgs, msg)
	}

	err := p.producer.SendMessages(msgs)
	if err != nil {
		p.logger.Error("could not sent msgs ",
			zap.String("topic", topic),
			zap.Int("len msgs", len(msgs)),
			zap.Any("data", msgs))
		return err
	}

	if len(msgs) > 0 {
		p.logger.Info("new delivery send to assembling",
			zap.String("topic", topic),
			zap.Int("len msgs", len(msgs)))
		return nil
	}
	p.logger.Info("new delivery send to assembling")

	return nil
}
