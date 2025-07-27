package main

import (
	pb "client_service/proto"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("localhost:8080", opts...)
	if err != nil {
		log.Fatalf("Failed to conn: %v", err)
	}

	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	log.Println("Creating a new order...")
	createReq := &pb.CreateOrderRequest{
		RequestId:      "987654321-0009",
		UserId:        "987654321",
		DarkstoreId:   "store789",
		
		// Информация о клиенте
		CustomerName:    "Иван",
		CustomerSurname: "Иванов",
		CustomerPhone:   "+79161234567",
		
		// Доставка
		DeliveryWindow:    "04:33-05:33",
		Address:          "ул. Пушкина, д. Колотушкина, кв. 123",
		UnderDoor:        true,
		CallBefore:       true,
		CommentToCourier: "Позвонить за 10 минут",
		
		// Заказ
		TotalPrice: 150, // Общая стоимость в рублях (150 рублей)
		Items: []*pb.OrderItem{
			{
				ProductId: "prod1",
				Name:      "Хлеб",
				Price:     50,  // 50 рублей
				Quantity:  1,
			},
			{
				ProductId: "prod2",
				Name:      "Молоко",
				Price:     50,  // 50 рублей
				Quantity:  2,
			},
		},
	}

	createRes, err := client.CreateOrder(ctx, createReq)
	if err != nil {
		log.Fatalf("Error during Create: %v", err)
	}

	log.Printf("order created: %+v\n", createRes)

	// cancelReq := &pb.CancelOrderRequest{OrderId: "3644c850-399c-477f-b1eb-71026b702995"}
	// cancelRes, err := client.CancelOrder(ctx, cancelReq)
	// if err != nil {
	// 	log.Fatalf("Error during cancel: %v", err)
	// }

	// log.Printf("order cancelled: %+v\n", cancelRes)

}
