CREATE TYPE order_status AS ENUM (
    'CREATED',              -- Заказ создан
    'ASSEMBLY_PENDING',     -- Ожидает, пока сборщик возьмет заказ
    'ASSEMBLING',           -- В процессе сборки
    'ASSEMBLY_COMPLETED',   -- Собран, ожидает курьера
    'DELIVERY_PENDING',     -- Ожидает, пока курьер возьмет заказ
    'DELIVERING',           -- В процессе доставки
    'DELIVERED',            -- Доставлен
    'CANCELLED'             -- Отменен
);

