package postgres


const (
    // Table names
    DeliveriesTable    = "deliveries"
    DeliveryItemsTable = "delivery_items"
    DeliveryOutboxTable = "delivery_outbox"

    // Common fields
    FieldID         = "id"
    FieldCreatedAt  = "created_at"
    FieldUpdatedAt  = "updated_at"

    // Deliveries table fields
    FieldOrderID          = "order_id"
    FieldUserID           = "user_id"
    FieldDarkstoreID      = "darkstore_id"
    FieldDeliveryWindow   = "delivery_window"
    FieldDeliveryAddress  = "delivery_address"
    FieldCommentToCourier = "comment_to_courier"
    FieldUnderDoor        = "under_door"
    FieldCallCourier      = "call_before"
    FieldUserName         = "user_name"
    FieldUserSurname      = "user_surname"
    FieldUserPhone        = "user_phone"

    // DeliveryItems table fields
    FieldDeliveryID = "delivery_id"
    FieldProductID  = "product_id"
    FieldQuantity   = "quantity"
    
    // DeliveryOutboxTable table fields

)