-- Таблица для доставок (Delivery)
CREATE TABLE IF NOT EXISTS deliveries (
    order_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    darkstore_id TEXT NOT NULL,
    delivery_window TEXT NOT NULL,
    delivery_address TEXT NOT NULL,
    comment_to_courier TEXT,
    under_door     BOOLEAN DEFAULT FALSE NOT NULL,
    call_before   BOOLEAN DEFAULT FALSE NOT NULL,
    user_name TEXT NOT NULL,
    user_surname TEXT NOT NULL,
    user_phone TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для товаров в доставке (DeliveryItem)
CREATE TABLE IF NOT EXISTS delivery_items (
    id TEXT PRIMARY KEY,
    delivery_id TEXT REFERENCES deliveries(order_id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создаем индекс для ускорения поиска товаров по доставке
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_delivery_items_delivery_id' 
        AND tablename = 'delivery_items'
    ) THEN
        CREATE INDEX idx_delivery_items_delivery_id ON delivery_items(delivery_id);
        RAISE NOTICE 'Index idx_delivery_items_delivery_id created';
    ELSE
        RAISE NOTICE 'Index idx_delivery_items_delivery_id already exists';
    END IF;
END
$$;