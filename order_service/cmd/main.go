package main

import (
	"context"
	"net"

	"github.com/kripst/order_service/config"
	"github.com/kripst/order_service/internal/kafka"
	"github.com/kripst/order_service/internal/service/orders"
	"github.com/kripst/order_service/internal/storage/postgres"
	pb "github.com/kripst/order_service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// логгер 
	config.InitLoggerZap()
	defer zap.L().Sync()
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("Failed to load config",
			zap.Error(err),
		)
	}
	
	// gRPC коннект
	lis, err := net.Listen(cfg.GrpcNet, cfg.GrpcAddress)
	if err != nil {
		zap.L().Fatal("не удалось коннект с интернетом %v",
			zap.Error(err), 
			zap.String("net",cfg.GrpcNet), 
			zap.String("address",cfg.GrpcAddress))
	}
	s := grpc.NewServer()

	// kafka connect
	kafkaProducer, err := kafka.NewKafkaProducer(cfg.KafkaConfig)
	if err != nil {
		zap.L().Fatal("не удалось коннект с кафкой %v", 
			zap.Error(err),
			zap.Any("kafka config",cfg.KafkaConfig))
	}

	// postgres connect
	PostgresRepository, err := postgres.NewPostgresRepository(context.Background(), cfg.PostgresPgxConfig)
	if err != nil {
		zap.L().Fatal("не удалось коннект с постгресом %v", 
			zap.Error(err),
			zap.Any("kafka config",cfg.KafkaConfig))
	}

	orderServer := &orders.OrderServer{
		Database: PostgresRepository,
		Producer: kafkaProducer,
	}
	pb.RegisterOrderServiceServer(s, orderServer)
	zap.L().Info("gRPC server listening",
		zap.String("grpcAddress", lis.Addr().String()),
	)
	if err := s.Serve(lis); err != nil {
		zap.L().Fatal("failed to serve: %v", zap.Error(err))
	}

}