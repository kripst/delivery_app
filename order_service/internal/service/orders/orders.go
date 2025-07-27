package orders

import (
	"context"

	"github.com/kripst/order_service/internal/kafka"
	"github.com/kripst/order_service/internal/storage/postgres"
	"github.com/kripst/order_service/model"
	pb "github.com/kripst/order_service/proto"
	"go.uber.org/zap"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	Database postgres.OrderRepositoty
	Producer kafka.Producer
}

func (s *OrderServer) CreateOrder(ctx context.Context, in *pb.CreateOrderRequest) (*pb.CreateOrderResponce, error) {
	order := &model.Order{}

	requestID := in.GetRequestId()
	zap.L().Info("get new order request",
		zap.String("request_id", requestID),
	)

	deliveryOrder, err := order.FromGrpc(in)
	if err != nil {
		zap.L().Error("empty mapping from grpc",
			zap.Error(err))
		return nil, err
	}

	zap.L().Debug("successfully mapping order, deliveryOrder from grpc",
		zap.Any("order struct", order),
		zap.Any("deliveryOrder struct", deliveryOrder))

	orderID, err := s.Database.CreateOrder(ctx, order)

	
	if err != nil {
		zap.L().Error("could not create the order",
			zap.Error(err),
			zap.String("order_id", requestID))
		return nil, err
	}

	if err := s.Producer.PostEvent(create_orders, deliveryOrder); err != nil {
			zap.L().Error("kafka send error",
				zap.Error(err),
				zap.String("topic", create_orders),
				zap.Any("send data", order))
			return nil, err
		}	

	return &pb.CreateOrderResponce{
		OrderId: orderID,
	}, nil
}

func (s *OrderServer) CancelOrder(ctx context.Context, in *pb.CancelOrderRequest) (*pb.SuccessResponse, error) {
	orderID := in.GetOrderId()
	zap.L().Info("get new cancel order request",
		zap.String("order_id", orderID),
	)

	if err := s.Database.CancelOrder(ctx, orderID); err != nil {
		return nil, err
	}

	if err := s.Producer.PostEvent(cancel_orders, orderID); err != nil {
		zap.L().Error("kafka send error",
			zap.Error(err),
			zap.String("topic", cancel_orders),
			zap.Any("send data", orderID))
		return nil, err
	}

	zap.L().Info("successfully cancel order in database",
		zap.String("order_id", orderID),
	)

	return &pb.SuccessResponse{}, nil
}
