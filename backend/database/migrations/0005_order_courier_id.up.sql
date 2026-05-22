ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS courier_id UUID REFERENCES couriers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_orders_courier_id ON orders (courier_id);