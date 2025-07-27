DROP TRIGGER IF EXISTS update_delivery_outbox_timestamp ON delivery_outbox;
DROP FUNCTION IF EXISTS update_modified_column();
DROP TABLE IF EXISTS delivery_outbox;
DROP TYPE IF EXISTS  delivery_status;
