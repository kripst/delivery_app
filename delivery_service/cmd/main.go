package main

import (
	l "log"
	"time"

	"context"

	"github.com/kripst/delivery_service/config"
	"github.com/kripst/delivery_service/internal/kafka"
	"github.com/kripst/delivery_service/internal/logger"
	s "github.com/kripst/delivery_service/internal/service"
	"github.com/kripst/delivery_service/internal/relayer"
	"github.com/kripst/delivery_service/internal/storage/postgres"
	"github.com/kripst/delivery_service/internal/storage/redis"
	// "github.com/kripst/delivery_service/internal/storage/redis"
)

func main() {
	// logger
	log, err := logger.NewLogger()
	if err != nil {
		l.Fatal("Could not initialize logger")
	}
	

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config")
	}
	

	// postgres
	postgresRepository, err := postgres.NewPostgresRepository(context.Background(), cfg.PostgresPgxConfig)
	if err != nil {
		log.Fatal("Failed to load PostgresRepository")
	}


	// kafka
	consumer, err := kafka.NewOrdersConsumer(cfg.KafkaConfig) 
	if err != nil {
		log.Fatal("Failed to load kafka consumer")
	}

	log.Info("kafka successfully connected")

	// Redis 
	timeManager, err := redis.NewTimeManagerImpl(cfg.RedisConfig)
	if err != nil {
		log.Fatal("Failed to load redis timeManager")
	}
	service, err := relayer.NewdeliveryServiceImpl(cfg.RedisConfig)
	if err != nil {
		log.Fatal("Failed to load redis timeManager")
	}
	outbox := relayer.NewoutboxServiceImpl(postgresRepository.Db, log)
	interval := time.Second * 60
	relayer := relayer.NewRelayer(service, interval, log, outbox)
	go func() {
		relayer.Start(context.Background())
	}()

	deliveryServer := &s.DeliveryServer{
		Consumer: consumer,
		Db:       postgresRepository,
		Logger: log,
		Tm : timeManager,
	}

	deliveryServer.HandleNewOrders()

}