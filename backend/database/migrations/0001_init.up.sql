CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    role          VARCHAR(50)  NOT NULL,
    CHECK (role IN ('admin','dispatcher','courier','client')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
    );
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);


CREATE TABLE IF NOT EXISTS couriers (
    id             UUID PRIMARY KEY,
    user_id        UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    transport_type VARCHAR(50)  NOT NULL CHECK (transport_type IN ('bike','car','foot','scooter')),
    status         VARCHAR(50)  NOT NULL DEFAULT 'offline',
    CHECK (status IN ('free','busy','offline')),
    current_lat    DOUBLE PRECISION NOT NULL DEFAULT 0,
    current_lng    DOUBLE PRECISION NOT NULL DEFAULT 0
    );
CREATE INDEX IF NOT EXISTS idx_couriers_status ON couriers (status);


CREATE TABLE IF NOT EXISTS orders (
    id               UUID PRIMARY KEY,
    client_id        UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pickup_address   TEXT         NOT NULL,
    delivery_address TEXT         NOT NULL,
    price            NUMERIC(10,2) NOT NULL CHECK (price > 0),
    status           VARCHAR(50)  NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','assigned','in_progress','delivered','cancelled')),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_client_id ON orders (client_id);
CREATE INDEX IF NOT EXISTS idx_orders_status    ON orders (status);


CREATE TABLE IF NOT EXISTS assignments (
    id           UUID PRIMARY KEY,
    order_id     UUID        NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    courier_id   UUID        NOT NULL REFERENCES couriers(id) ON DELETE CASCADE,
    assigned_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
    );
CREATE INDEX IF NOT EXISTS idx_assignments_courier_id ON assignments (courier_id);