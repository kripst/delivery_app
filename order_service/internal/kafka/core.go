package kafka

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/kripst/order_service/config"
	"go.uber.org/zap"
)

type Producer interface {
	PostEvent(topic string, eventData any) error

}

type KafkaProducer struct {
	Producer sarama.SyncProducer
	RetryCount int
}

func NewKafkaProducer(config *config.KafkaConfig) (*KafkaProducer, error) {
	saramaSyncProducer, err := sarama.NewSyncProducer(config.Brokers, config.SaramaConfig)
	if err != nil {
		return nil, err
	}
	
	zap.L().Info("Kafka succesfully connected", zap.Any("brokers", config.Brokers))	
	return &KafkaProducer{Producer: saramaSyncProducer, RetryCount: config.RetryCount}, nil
}

func (r *KafkaProducer) PostEvent(topic string, dataEvent any) error {
	jsonData, err := json.Marshal(dataEvent)
	if err != nil {
		return errInvalidInput
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonData),
	}

	var (
		partition int32
		offset  int64
		lastErr error
	)

	for i := 0; i < r.RetryCount; i++ {
		// Пытаемся отправить сообщение
		p, o, err := r.Producer.SendMessage(msg)
		lastErr = err

		if err == nil {
			partition = p
			offset = o
			zap.L().Info("Message successfully sent",
				zap.Int32("partition", partition),
				zap.Int64("offset", offset),
				zap.Int("attempt", i+1),
				zap.Any("json data",string(jsonData)))
			break
		}
		
		retryDecision := shouldRetry(lastErr)
		if retryDecision.Retry {
			zap.L().Error("error to send message after retry queue",
				zap.Error(lastErr),
			)
			return lastErr
		}
		
		zap.L().Warn("Failed to send message",
			zap.Error(lastErr),
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", r.RetryCount))

		time.Sleep(retryDecision.Delay)
	}
	
	if lastErr != nil {
		zap.L().Error("error to send message after retry queue",
				zap.Error(lastErr),
			)
		return lastErr
	}

	zap.L().Info("order is stored",
		zap.String("topic", topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))
	
	return nil
}

func (r *KafkaProducer) Close() {
	r.Producer.Close()
}

