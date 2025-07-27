package config

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type KafkaConfig struct {
	SaramaConfig *sarama.Config
	Brokers      []string
}

type RedisConfig struct {
	Opts *redis.Options
}

type config struct {
	PostgresUrl       string
	PostgresPgxConfig *pgxpool.Config
	Dsn               string
	KafkaConfig       *KafkaConfig
	RedisConfig       *RedisConfig
}

func Load() (config, error) {
	// пока что тут переменные, в будущем в .env НАДО ПОМЕНЯТЬ на PostgresConfig Grpc Config ...
	postgresHost := "localhost"
	postgresUser := "user"
	postgresPass := "password"
	postgresDbName := "dbname"
	postgresPort := "5430"
	postgresSslMode := "disable"
	postgresTimeZone := "Europe/Moscow"
	kafkaBrokers  := []string{"localhost:9092"}
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Errors = true

	kafkaConfig := &KafkaConfig{
		SaramaConfig: saramaConfig,
		Brokers:      kafkaBrokers,
	}

	redisOpts := &redis.Options{
		Addr: "localhost:6379",
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		postgresHost, postgresUser, postgresPass, postgresDbName,
		postgresPort, postgresSslMode, postgresTimeZone,
	)

	postgresPgxConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		
	}

	// Опциональные настройки пула
	postgresPgxConfig.MaxConns = 100
	postgresPgxConfig.MinConns = 5                 // Максимальное число соединений                
	postgresPgxConfig.MaxConnLifetime = 30 * time.Second    // Максимальное время жизни соединения
	postgresPgxConfig.HealthCheckPeriod = 10 * time.Second // Как часто проверять соединения

	return config{
		PostgresPgxConfig:  postgresPgxConfig,
		Dsn:                dsn,
		KafkaConfig: kafkaConfig,
		RedisConfig: &RedisConfig{
			Opts: redisOpts,
		},
	}, nil

}