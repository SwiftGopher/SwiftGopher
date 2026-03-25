INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@swiftgopher.io',
    '$2a$10$rRQu5V8p7wlWzPTVBJkxUOv3sVbKpJqVpH7A3EO/BjMvJjMMMzUwW',
    'admin',
    NOW()
) ON CONFLICT DO NOTHING;

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'dispatcher@swiftgopher.io',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lh/K',
    'dispatcher',
    NOW()
) ON CONFLICT DO NOTHING;

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'courier1@swiftgopher.io',
    '$2a$10$9Ux0jC3wQuXBDXIGKq2AZuOMiXEJW9oiyxzqpJGJkJKkZGRLuiI8K',
    'courier',
    NOW()
) ON CONFLICT DO NOTHING;

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000004',
    'client1@swiftgopher.io',
    '$2a$10$7HmVMvXzFo1OdOxV5vFcEuD3HtBz8nN5N2XmQMmqFqFjkJkZGRLuiI',
    'client',
    NOW()
) ON CONFLICT DO NOTHING;

INSERT INTO couriers (id, user_id, transport_type, status, current_lat, current_lng)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000003',
    'bike', 'free', 51.5074, -0.1278
) ON CONFLICT DO NOTHING;

INSERT INTO orders (id, client_id, pickup_address, delivery_address, price, status, created_at, updated_at)
VALUES (
    'aaaaaaaa-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000004',
    '10 Downing Street, London',
    'Buckingham Palace, London',
    12.50, 'pending', NOW(), NOW()
) ON CONFLICT DO NOTHING;