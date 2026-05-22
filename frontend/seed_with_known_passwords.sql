-- SwiftGopher seed data with KNOWN passwords
-- admin@swiftgopher.io    → admin123
-- dispatcher@swiftgopher.io → disp123
-- courier1@swiftgopher.io → courier123
-- client1@swiftgopher.io  → client123

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
           '00000000-0000-0000-0000-000000000001',
           'admin@swiftgopher.io',
           '$2b$10$JiWxsqeVV2rTr4ceQt8WEe0IL9K.mrkHAO0/n1kaMOUDKoQZGEon2',
           'admin', NOW()
       ) ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
           '00000000-0000-0000-0000-000000000002',
           'dispatcher@swiftgopher.io',
           '$2b$10$Z8ff6x6jah8jwzCoDRmWS.F9axf7mRYcGQaKIPay/fN4Z3io9d.ye',
           'dispatcher', NOW()
       ) ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
           '00000000-0000-0000-0000-000000000003',
           'courier1@swiftgopher.io',
           '$2b$10$cs4p0CTvcpleGM/iTRZVluDuGz/LdM61jMIxZNSB8BRFn3Z2zH6t6',
           'courier', NOW()
       ) ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES (
           '00000000-0000-0000-0000-000000000004',
           'client1@swiftgopher.io',
           '$2b$10$im0fCj.WMl7jm.lvIBlrQON/rXuq8e9BeqeU2YsawiysuGZgZgEti',
           'client', NOW()
       ) ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

INSERT INTO couriers (id, user_id, transport_type, status, current_lat, current_lng)
VALUES (
           'c0000000-0000-0000-0000-000000000001',
           '00000000-0000-0000-0000-000000000003',
           'bike', 'free', 43.2220, 76.8512
       ) ON CONFLICT (id) DO NOTHING;

-- Multiple sample orders for demo
INSERT INTO orders (id, client_id, pickup_address, delivery_address, price, status, created_at, updated_at)
VALUES
    ('aaaaaaaa-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004','Almaty, Abay 1','Almaty, Dostyk 10',1500.00,'pending',NOW(),NOW()),
    ('aaaaaaaa-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000004','Almaty, Seifullin 5','Almaty, Al-Farabi 50',2000.00,'pending',NOW(),NOW()),
    ('aaaaaaaa-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000004','Almaty, Furmanov 100','Almaty, Raiymbek 200',800.00,'delivered',NOW()-INTERVAL '2 hours',NOW())
    ON CONFLICT (id) DO NOTHING;
