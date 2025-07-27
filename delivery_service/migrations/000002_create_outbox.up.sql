CREATE TYPE delivery_status AS ENUM (
    'WAITING',
    'PENDING',
    'SENT'
);

CREATE TABLE IF NOT EXISTS delivery_outbox (
    delivery_id     TEXT            PRIMARY KEY UNIQUE            NOT NULL,
    delivery_window TEXT                                          NOT NULL,
    delivery_status delivery_status DEFAULT 'WAITING'             NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE     DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE     DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW(); -- Перезаписываем поле перед обновлением
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. Создаем триггер для таблицы `orders`
CREATE TRIGGER update_delivery_outbox_timestamp
BEFORE UPDATE ON delivery_outbox
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();