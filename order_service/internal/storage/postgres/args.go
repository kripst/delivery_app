package postgres


const (
    StatusCreated           = "CREATED"
    StatusAssemblyPending  = "ASSEMBLY_PENDING"
    StatusAssembling        = "ASSEMBLING"
    StatusAssemblyCompleted  = "ASSEMBLY_COMPLETED"
    StatusDeliveryPending   = "DELIVERY_PENDING"
    StatusDelivering        = "DELIVERING"
    StatusDelivered         = "DELIVERED"
    StatusCancelled         = "CANCELLED"
)

const (
	OutboxEventStatusPending  = "PENDING"
	OutboxEventStatusSent     = "SENT"
	OutboxEventStatusFailed   = "FAILED"
)


const (
    OrderOutboxEventsTable  = "outbox_events_table"
	FieldID                 = "id"
	FieldOrderOutboxID      = "order_id"
	FieldEventType          = "event_type"
	FieldPayload            = "payload"
	FieldCreatedAt          = "created_at"
	FieldProcessedAt        = "processed_at"
	FieldEventStatus        = "event_status"
)

const (
    OrderTable           = "orders_table"
    FieldOrderID         = "id"
    FieldUserID          = "user_id"
    FieldDarkstoreID     = "darkstore_id"
    FieldOrderStatus     = "order_status"
    FieldCustomerName    = "customer_name"
    FieldCustomerSurname = "customer_surname"
    FieldCustomerPhone   = "customer_phone"
    FieldDeliveryWindow  = "delivery_window"
    FieldAddress         = "delivery_address"
    FieldUnderDoor       = "under_door"
    FieldCallBefore      = "call_before"
    FieldCommentToCourier= "comment_to_courier"
    FieldTotalPrice      = "total_price"
)

const (
    OrderItemsTable           = "order_items_table"
    FieldOrderItemID          = "id"
    FieldOrderItemOrderID     = "order_id"
    FieldOrderItemProductID   = "product_id"
    FieldOrderItemProductName = "product_name"
    FieldOrderItemPrice       = "price"
    FieldOrderItemQuantity    = "quantity"
)