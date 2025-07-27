package model

type Delivery struct {
    OrderID          string         `json:"OrderID"`          // ID заказа
    UserID           string         `json:"UserID"`           // ID пользователя
    DarkstoreID      string         `json:"DarkstoreID"`      // ID темного магазина (склада)
    DeliveryWindow   string         `json:"DeliveryWindow"`   // Временное окно доставки
    Address          string         `json:"Address"`          // Адрес доставки
	CallBefore       bool           
	UnderDoor        bool
    CommentToCourier string         `json:"CommentToCourier"` // Комментарий курьеру
    UserName         string         `json:"UserName"`         // Имя пользователя
    UserSurname      string         `json:"UserSurname"`     // Фамилия пользователя
    UserPhone        string         `json:"UserPhone"`       // Телефон
    DeliveryItems    []*DeliveryItem `json:"DeliveryItems"`    // Список товаров
}

type DeliveryItem struct {
    ID        string `json:"ID"`               // Уникальный ID позиции
    ProductID string `json:"ProductID"`        // ID товара в системе
    Quantity  int32  `json:"Quantity"`         // Количество
}