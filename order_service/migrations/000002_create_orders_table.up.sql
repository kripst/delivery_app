CREATE TABLE IF NOT EXISTS orders_table
(
    id                   TEXT                      NOT NULL,
    user_id              TEXT                      NOT NULL,
    darkstore_id         TEXT                      NOT NULL,
    order_status         order_status              NOT NULL DEFAULT 'CREATED',
    total_price          INTEGER                   NOT NULL,
    created_at           TIMESTAMP WITH TIME ZONE  NOT NULL DEFAULT NOW(),
    
    CONSTRAINT pk_orders_table PRIMARY KEY (id)
);