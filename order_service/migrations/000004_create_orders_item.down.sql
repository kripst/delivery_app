ALTER TABLE order_items_table
DROP CONSTRAINT IF EXISTS fk_order_items_order;
DROP TABLE IF EXISTS order_items_table;