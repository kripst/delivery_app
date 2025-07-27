package model

import "errors"

var ErrOrderIsNil = errors.New("Order is nil")
var ErrCreateOrderRequestIsNil = errors.New("CreateOrderRequest is nil")