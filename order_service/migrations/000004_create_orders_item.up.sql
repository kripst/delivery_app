CREATE TABLE IF NOT EXISTS order_items_table
(
    id                   TEXT                      NOT NULL,
    order_id             TEXT                      NOT NULL,
    product_id           TEXT                      NOT NULL,
    product_name         TEXT                      NOT NULL,
    quantity             INTEGER                   NOT NULL,
    price                INTEGER                   NOT NULL,
    
    CONSTRAINT pk_order_items PRIMARY KEY (id),
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders_table(id)
);