package kafka

import (
	"errors"
	"net"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)


// shouldRetry определяет, нужно ли повторять операцию для ошибки Sarama
func shouldRetry(err error) RetryDecision {
	if err == nil {
		return RetryDecision{false, 0}
	}

	// 1. Критические ошибки (не ретраить)
		switch {
	case errors.Is(err, sarama.ErrUnknownTopicOrPartition),
		errors.Is(err, sarama.ErrInvalidConfig),
		errors.Is(err, sarama.ErrMessageSizeTooLarge),
		errors.Is(err, sarama.ErrSASLAuthenticationFailed),
		errors.Is(err, sarama.ErrTopicAuthorizationFailed),
		errors.Is(err, sarama.ErrClusterAuthorizationFailed):
		zap.L().Error("Critical Kafka error, no retry", zap.Error(err))
		return RetryDecision{false, 0}
	}

	// 2. Сетевые ошибки (ретраить)
	var netErr net.Error
	if errors.As(err, &netErr) {
		zap.L().Warn("Network-related error, will retry", zap.Error(err))
		return RetryDecision{true, 1 * time.Second}
	}

	// 3. Временные ошибки Kafka (ретраить)
	var kafkaErr sarama.KError
	if errors.As(err, &kafkaErr) {
		switch kafkaErr {
		case sarama.ErrNotEnoughReplicas,
			sarama.ErrNotEnoughReplicasAfterAppend,
			sarama.ErrRequestTimedOut,
			sarama.ErrBrokerNotAvailable,
			sarama.ErrLeaderNotAvailable,
			sarama.ErrNetworkException:
			delay := calculateKafkaBackoff(kafkaErr)
			zap.L().Warn("Temporary Kafka error, will retry",
				zap.Error(err),
				zap.Duration("delay", delay))
			return RetryDecision{true, delay}
		}
	}

	// 4. Неизвестные ошибки (по умолчанию не ретраить)
	zap.L().Error("Unknown error type, no retry", zap.Error(err))
	return RetryDecision{false, 0}
}

// calculateKafkaBackoff вычисляет задержку для ошибок Sarama
func calculateKafkaBackoff(err sarama.KError) time.Duration {
	switch err {
	case sarama.ErrRequestTimedOut:
		return 2 * time.Second
	case sarama.ErrNotEnoughReplicas:
		return 3 * time.Second
	case sarama.ErrLeaderNotAvailable:
		return 5 * time.Second
	default:
		return 1 * time.Second
	}
}

type RetryDecision struct {
	Retry bool
	Delay time.Duration
}