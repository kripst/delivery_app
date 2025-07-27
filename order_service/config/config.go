package config

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KafkaConfig struct {
	SaramaConfig       *sarama.Config
	Brokers            []string
	RetryCount         int
}


type config struct {
	GrpcAddress        string
	GrpcNet            string
	PostgresUrl        string
	PostgresPgxConfig  *pgxpool.Config
	Dsn                string
	KafkaConfig        *KafkaConfig
}


func Load() (config, error) {
	// пока что тут переменные, в будущем в .env НАДО ПОМЕНЯТЬ на PostgresConfig Grpc Config ...
	grpcAddress := ":8080"
	grpcNet := "tcp"
	postgresHost := "localhost"
	postgresUser := "user"
	postgresPass := "password"
	postgresDbName := "dbname"
	postgresPort := "5431"
	postgresSslMode := "disable"
	postgresTimeZone := "Europe/Moscow"
	kafkaBrokers  := []string{"localhost:9092"}
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll

	KafkaConfig := &KafkaConfig{
		SaramaConfig: saramaConfig,
		Brokers:      kafkaBrokers,
		RetryCount:   5,
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
		GrpcAddress:        grpcAddress,
		GrpcNet:            grpcNet,
		PostgresPgxConfig:  postgresPgxConfig,
		Dsn:                dsn,
		KafkaConfig: KafkaConfig,
	}, nil

}

