package model

import (

	"github.com/google/uuid"
	pb "github.com/kripst/order_service/proto"
)

type Order struct {
	ID               string       `json:"id"`                 // column: id (primary key)
	UserID           string       `json:"user_id"`            // column: user_id
	DarkstoreID      string       `json:"darkstore_id"`       // column: darkstore_id
	CustomerName     string       `json:"customer_name"`      // column: customer_name
	CustomerSurname  string       `json:"customer_surname"`   // column: customer_surname
	CustomerPhone    string       `json:"customer_phone"`     // column: customer_phone
	DeliveryWindow   string       `json:"delivery_window"`    // column: delivery_window
	Address          string       `json:"address"`            // column: address
	UnderDoor        bool         `json:"under_door"`         // column: under_door
	CallBefore       bool         `json:"call_before"`        // column: call_before
	CommentToCourier string       `json:"comment_to_courier"` // column: comment_to_courier
	TotalPrice       int32        `json:"-"`                  // column: total_price (не экспортируется в JSON)
	Items            []*OrderItem `json:"items"`              // Связь по order_id
}

type OrderItem struct {
	ID          string `json:"id"`          // column: id (primary key)
	OrderID     string `json:"order_id"`    // column: order_id (foreign key)
	ProductID   string `json:"product_id"`  // column: product_id
	ProductName string `json:"product_name"` // column: product_name
	Quantity    int32  `json:"quantity"`    // column: quantity
	Price       int32  `json:"price"`      // column: price
}


type OrderStatus int32

const (
	StatusNew        OrderStatus = 0
	StatusProcessing OrderStatus = 1
	StatusReady      OrderStatus = 2
	StatusDelivering OrderStatus = 3
	StatusDelivered  OrderStatus = 4
	StatusCancelled  OrderStatus = 5
)

func (o *Order) FromGrpc(in *pb.CreateOrderRequest) (*Delivery, error) {
	if o == nil {
		return nil, ErrOrderIsNil  // или panic(), если это контракт метода
	}
	if in == nil {
		return nil ,ErrCreateOrderRequestIsNil
	}
	delivery := &Delivery{}
	delivery.DeliveryItems = make([]*DeliveryItem, len(in.GetItems()))

	// Основные поля
	o.ID = in.GetRequestId()
	delivery.OrderID = o.ID
	o.UserID = in.GetUserId()
	delivery.UserID = o.UserID
	o.DarkstoreID = in.GetDarkstoreId()
	delivery.DarkstoreID = o.DarkstoreID

	// Информация о клиенте
	delivery.UserName = in.GetCustomerName()
	delivery.UserSurname = in.GetCustomerSurname()
	delivery.UserPhone = in.GetCustomerPhone()

	// Доставка
	delivery.DeliveryWindow = in.GetDeliveryWindow()
	delivery.Address = in.GetAddress()
	delivery.UnderDoor = in.GetUnderDoor()
	delivery.CallBefore = in.GetCallBefore()
	delivery.CommentToCourier = in.GetCommentToCourier()

	// Заказ
	o.TotalPrice = in.GetTotalPrice()
	o.Items = make([]*OrderItem, len(in.GetItems()))
	for i, item := range in.GetItems() {
		o.Items[i] = &OrderItem{
			ID:        uuid.NewString(), // Генерируем уникальный ID для позиции НАДО ПОМЕНЯТЬ 
			ProductID: item.GetProductId(),
			ProductName:      item.GetName(),
			Quantity:  item.GetQuantity(),
			Price:     item.GetPrice(),
		}
		delivery.DeliveryItems[i] = &DeliveryItem{
			ID:        uuid.NewString(),
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
		}
	}

	return delivery, nil
}