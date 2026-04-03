CREATE TABLE IF NOT EXISTS order_history (
    id         UUID PRIMARY KEY,
    order_id   UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    old_status VARCHAR(50) NOT NULL,
    new_status VARCHAR(50) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_history_order_id ON order_history (order_id);