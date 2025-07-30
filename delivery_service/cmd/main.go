package main

import (
	l "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/IBM/sarama"
	"github.com/kripst/delivery_service/config"
	kafka "github.com/kripst/delivery_service/internal/kafka_consumer"
	"github.com/kripst/delivery_service/internal/kafka_worker"
	"github.com/kripst/delivery_service/internal/logger"
	"github.com/kripst/delivery_service/internal/relayer"
	s "github.com/kripst/delivery_service/internal/service"
	"github.com/kripst/delivery_service/internal/storage/postgres"
	"github.com/kripst/delivery_service/internal/storage/redis"
	"go.uber.org/zap"
	// "github.com/kripst/delivery_service/internal/storage/redis"
)

func main() {
	// logger
	log, err := logger.NewLogger()
	if err != nil {
		l.Fatal("Could not initialize logger")
	}
	
	// graceful shd ctx
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config",
			zap.Error(err))
	}
	

	// postgres
	postgresRepository, err := postgres.NewPostgresRepository(context.Background(), cfg.PostgresPgxConfig)
	if err != nil {
		log.Fatal("Failed to load PostgresRepository",
			zap.Error(err))
	}

	interval := time.Second * 20

	// kafka
	
	defer cancel()
	consumer, err := kafka.NewOrdersConsumer(ctx, cfg.KafkaConsumerConfig) 
	if err != nil {
		log.Fatal("Failed to load kafka consumer",
			zap.Error(err))
	}

	log.Info("kafka successfully connected")

	// Redis 
	timeManager, err := redis.NewTimeManagerImpl(cfg.RedisConfig)
	if err != nil {
		log.Fatal("Failed to load redis timeManager",
			zap.Error(err))
	}
	service, err := relayer.NewdeliveryServiceImpl(cfg.RedisConfig)
	if err != nil {
		log.Fatal("Failed to load redis timeManager",
			zap.Error(err))
	}
	outbox := postgres.NewoutboxServiceImpl(postgresRepository.Db, log)
	
	relayer := relayer.NewRelayer(service, interval, log, outbox)
	go func() {
		relayer.Start(context.Background())
	}()
	
	// kafka_worker
	producer, err := sarama.NewSyncProducer(cfg.KafkaProducerConfig.Brokers, cfg.KafkaProducerConfig.SaramaConfig)
	if err != nil {
		log.Fatal("Failed to load kafka producer",
			zap.Error(err))
	}
	outboxProducer := kafka_worker.NewOutboxProducer(producer, log)
	outboxService  := postgres.NewoutboxServiceImpl(postgresRepository.Db, log)
	kafka_worker := kafka_worker.NewKafkaWorker(outboxProducer, outboxService, log)

	go func() {
		kafka_worker.Start(context.Background(), interval)
	}()

	deliveryServer := &s.DeliveryServer{
		Consumer: consumer,
		Db:       postgresRepository,
		Logger: log,
		Tm : timeManager,
		Ctx: ctx,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		deliveryServer.HandleNewOrders()
	}()

	<-sigChan
	log.Info("TODO graceful shd active")
	//TODO добавить для каждого сервиса поле read
	// y chan bool или Active короче чтобы когда система падала , то сменялось на тру ,
	//  только если закончил все свои задачи сервис

}