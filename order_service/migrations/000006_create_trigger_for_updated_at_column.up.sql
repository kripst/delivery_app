-- 1. Создаем функцию, которую будет вызывать триггер
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW(); -- Перезаписываем поле перед обновлением
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. Создаем триггер для таблицы `orders`
CREATE TRIGGER update_orders_table_timestamp
BEFORE UPDATE ON orders_table
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();