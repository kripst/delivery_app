package model

const DeliveryOrderKey = "delivery_orders"

type Delivery struct {
    OrderID          string         `json:"OrderID"`          // ID заказа
    UserID           string         `json:"UserID"`           // ID пользователя
    DarkstoreID      string         `json:"DarkstoreID"`      // ID темного магазина (склада)
    DeliveryWindow   string         `json:"DeliveryWindow"`   // Временное окно доставки
    Address          string         `json:"Address"`          // Адрес доставки
	CallBefore       bool           `json:"call_before"`
	UnderDoor        bool           `json:"under_door"`
    CommentToCourier string         `json:"CommentToCourier"` // Комментарий курьеру
    UserName         string         `json:"UserName"`         // Имя пользователя
    UserSurname      string         `json:"UserSurname"`     // Фамилия пользователя
    UserPhone        string         `json:"UserPhone"`       // Телефон
    DeliveryItems    []*DeliveryItem `json:"DeliveryItems"`    // Список товаров
}

type DeliveryInput struct {
    OrderID          string          `json:"OrderID"`
    UserID           string          `json:"UserID"`
    DarkstoreID      string          `json:"DarkstoreID"`
    Address          string          `json:"Address"`
    CallBefore       bool            `json:"call_before"`
    UnderDoor        bool            `json:"under_door"`
    CommentToCourier string          `json:"CommentToCourier"`
    UserName         string          `json:"UserName"`
    UserSurname      string          `json:"UserSurname"`
    UserPhone        string          `json:"UserPhone"`
    DeliveryItems    []*DeliveryItem `json:"DeliveryItems"`
}


type DeliveryItem struct {
    ID        string `json:"ID"`               // Уникальный ID позиции
    ProductID string `json:"ProductID"`        // ID товара в системе
    Quantity  int32  `json:"Quantity"`         // Количество
}