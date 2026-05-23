# SwiftGopher — Delivery Management REST API

SwiftGopher is a production-grade Go backend service for managing the full lifecycle of delivery orders — from creation and courier assignment through real-time status tracking to final delivery. The system supports four distinct roles (admin, dispatcher, courier, client) and automatically assigns the nearest available courier to every new order.

---

## Table of Contents

1. [Team](#team)
2. [Project Overview](#project-overview)
3. [Rubric Coverage](#rubric-coverage)
4. [Tech Stack](#tech-stack)
5. [Architecture](#architecture)
6. [Database Schema](#database-schema)
7. [API Endpoints](#api-endpoints)
8. [Getting Started](#getting-started)
9. [Running Tests](#running-tests)
10. [Environment Variables](#environment-variables)
11. [Monitoring](#monitoring)

---

## Team

| Name | Role | Name |
|------|------|------|
| Member 1 | Team Lead | Marzhan Omarbekova |
| Member 2 | Scrum Master | Zhangarys Ryskali |
| Member 3 | Core Backend Developer | Ayala Nurakyn |
| Member 4 | QA Engineer | Yerdaulet Tileukhanov |

---

## Project Overview

SwiftGopher solves a real-world last-mile delivery scenario. Clients create orders with pickup and delivery addresses; the system automatically finds the geographically nearest free courier using the Haversine formula and assigns the order. Dispatchers and admins monitor all orders and couriers in real time. The backend exposes a clean REST API consumed by an Angular frontend.

**Domain entities:** Users, Couriers, Orders, Assignments, OrderHistory

---

## Rubric Coverage

### Authentication & Authorization

- JWT-based registration and login (`POST /auth/register`, `POST /auth/login`)
- Secure token generation (HS256) and validation via `golang-jwt/jwt`
- Refresh token flow (`POST /auth/refresh`)
- Logout with token blacklisting stored in Redis — revoked tokens are rejected on subsequent requests
- Password hashing with `bcrypt` (cost factor 10)
- Role-based access control across four roles: `admin`, `dispatcher`, `courier`, `client`
- Every protected route is guarded by JWT middleware and a `RequireRole` middleware that enforces per-endpoint role requirements

### CRUD Operations

- Three major entities with meaningful relationships:
    - `users` — `one-to-many` with `orders` (a client has many orders), `one-to-one` with `couriers`
    - `couriers` — `one-to-many` with `assignments`
    - `orders` — `one-to-many` with `order_history`; `assignments` is a join/linking table between orders and couriers (many-to-one from both sides)
- Full CRUD for orders, courier status/transport/location, and order status transitions
- Repository/service (usecase) layered architecture with interfaces at every boundary
- Idiomatic Go error handling — domain errors (`ErrOrderNotFound`, `ErrEmailTaken`, etc.) are defined in the usecase layer and mapped to HTTP status codes in handlers
- Request validation (Gin binding tags + custom logic in usecases)

### Database & Migrations

- PostgreSQL with `pgx/v5` connection pool
- Schema managed by `golang-migrate/migrate` — five sequential versioned migrations (`0001` through `0005`)
- Foreign keys with `ON DELETE CASCADE` between all related tables
- Indexes on high-cardinality query columns: `idx_users_email`, `idx_orders_client_id`, `idx_orders_status`, `idx_orders_courier_id`, `idx_couriers_status`, `idx_assignments_courier_id`, `idx_order_history_order_id`
- Seed migration with known-password test accounts for every role
- Auto-migration runs on application startup

### Concurrency & Context

- Background dispatcher worker runs on a configurable interval (`WORKER_INTERVAL`, default 30 s)
- Worker pool with a fixed pool size of 5 goroutines and a buffered job queue of 50; implements safe shutdown with `WaitGroup`
- Dispatcher uses goroutines and channels internally (`pool.go`)
- Context propagation from HTTP request through repository layer; `context.WithTimeout` on DB fetch operations inside the dispatcher
- Graceful shutdown: SIGINT/SIGTERM triggers context cancellation, dispatcher drains its pool, HTTP server calls `Shutdown` with a 15-second timeout
- Retry with exponential backoff (`pkg/retry`) used by the dispatcher when assignment fails transiently; supports non-retryable error wrapping

### API Documentation

- Full endpoint documentation in `backend/ENDPOINTS.md` — every route listed with method, path, auth requirements, request/response schema, and all status codes
- This README provides a summary table

### Testing

- Unit tests using the standard `testing` package — no external test framework required
- `testify/assert` used in HTTP handler integration tests
- Test coverage areas:
    - Auth usecase: register, duplicate email, invalid role, login, wrong password, token refresh, token validation
    - Order usecase: create, invalid price, zero price, missing address, get, update status (valid/invalid transitions), full lifecycle, cancel from pending, terminal state enforcement, pagination, history recording
    - Courier usecase: get, not found, list, list free, update status
    - Dispatcher: assignment success, nearest-courier selection
    - Worker pool: job execution, queue-full rejection, shutdown drain, panic recovery, post-shutdown rejection
    - Retry package: first-attempt success, second-attempt success, max-attempts exhaustion, non-retryable abort
    - Middleware: CORS allow-all, CORS preflight, specific origin allow/block, rate limit pass, rate limit block, independent IPs, idempotency pass-through variants
    - Token blacklist: not revoked by default, revoked after logout, TTL expiry, independent tokens
    - Cache: set/get, missing key, expiry, delete

### Code Organization & Best Practices

- Standard Go project layout:
  ```
  backend/
    cmd/app/          — entrypoint
    internal/
      app/            — wiring & lifecycle
      handler/        — HTTP handlers (auth, order, courier, user, logout)
      middleware/      — JWT, CORS, rate limiting, recovery, idempotency, logger
      repository/      — interfaces + postgres & redis implementations
      usecase/         — business logic (auth, order, courier, user)
      worker/          — dispatcher + pool
    pkg/
      cache/           — Redis client wrapper
      metrics/         — Prometheus registry + middleware
      modules/         — shared models and config structs
      redisclient/     — typed Redis wrapper
      retry/           — exponential backoff with non-retryable errors
  database/migrations/ — SQL migration files
  tests/               — unit & integration tests
  ```
- Dependency injection via constructor functions; all inter-layer dependencies are injected as interfaces
- Structured logging with Go 1.21 `log/slog` — JSON in production, text in development
- Containerised with Docker; `compose.yaml` orchestrates backend, PostgreSQL, Redis, Prometheus, and Grafana

### Advanced / Bonus Features Implemented

| Feature | Implementation |
|---------|---------------|
| Rate limiting (in-memory) | Token-bucket per IP, `middleware/ratelimit.go` |
| CORS middleware | Configurable allowed origins, preflight support, `middleware/cors.go` |
| Request logging middleware | Structured per-request logs (method, path, status, duration, IP) |
| Panic recovery middleware | Catches panics, logs stack trace, returns 500 |
| Worker pool for background tasks | Fixed-size goroutine pool with job channel and graceful drain |
| Prometheus metrics | Custom registry with total requests, OK/error counts by path, total orders counter; exposed at `GET /metrics` |
| Grafana dashboard | Pre-configured datasource pointing at Prometheus |
| Retry with exponential backoff | `pkg/retry` with non-retryable error wrapping |
| Idempotency middleware | `X-Idempotency-Key` header deduplicates POST requests via Redis |
| Token blacklist (logout) | Redis-backed revocation list checked on every authenticated request |
| Nearest-courier algorithm | Haversine great-circle distance; implemented independently in both the usecase layer and the dispatcher |
| Angular frontend | Full SPA with role-based navigation, order creation with geocoding, courier management, i18n (EN/RU/KZ) |

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| HTTP framework | Gin v1.10 |
| Database | PostgreSQL 17 |
| DB driver | `pgx/v5` with connection pool |
| Migrations | `golang-migrate/migrate` |
| Auth | `golang-jwt/jwt` v5, `bcrypt` |
| Cache / Blacklist | Redis 8 (`go-redis/v9`) |
| Metrics | Prometheus client Go |
| Dashboards | Grafana |
| Containerisation | Docker + Docker Compose |
| Testing | `testing`, `testify` |
| Frontend | Angular 17, TypeScript, ngx-translate |

---

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                        HTTP Layer                          │
│  Gin Router ─ Middleware (JWT, CORS, RateLimit, Logger)    │
│              └─ Handlers (auth, order, courier, user)      │
└──────────────────────────┬─────────────────────────────────┘
                           │ calls
┌──────────────────────────▼─────────────────────────────────┐
│                      Usecase Layer                         │
│  AuthUsecase │ OrderUsecase │ CourierUsecase │ UserUsecase  │
└──────────────────────────┬─────────────────────────────────┘
                           │ calls (via interfaces)
┌──────────────────────────▼─────────────────────────────────┐
│                    Repository Layer                        │
│  Postgres (pgx) │ Redis (token blacklist, idempotency)     │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│              Background Worker (goroutine)                 │
│  Dispatcher ─ Worker Pool (5 goroutines) ─ Retry logic     │
│  Polls pending orders every 30 s; assigns nearest courier  │
└────────────────────────────────────────────────────────────┘
```

---

## Database Schema

```
users
  id UUID PK | email UNIQUE | password_hash | role | created_at

couriers
  id UUID PK | user_id FK(users) UNIQUE | transport_type | status | current_lat | current_lng

orders
  id UUID PK | client_id FK(users) | courier_id FK(couriers) NULL
  pickup_address | pickup_lat | pickup_lng | delivery_address
  price | status | created_at | updated_at

assignments
  id UUID PK | order_id FK(orders) UNIQUE | courier_id FK(couriers)
  assigned_at | completed_at NULL

order_history
  id UUID PK | order_id FK(orders) | old_status | new_status | changed_at
```

**Relationships:**
- `users` 1:1 `couriers` (a courier user has exactly one courier profile)
- `users` 1:N `orders` (a client has many orders)
- `couriers` 1:N `assignments`
- `orders` 1:1 `assignments` (unique constraint on `order_id`)
- `orders` 1:N `order_history`

---

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/auth/register` | Register user |
| POST | `/auth/login` | Login, receive JWT pair |
| POST | `/auth/refresh` | Refresh access token |

### Protected (Bearer token required)

| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST | `/auth/logout` | any | Blacklist current token |
| GET | `/profile` | any | Get own profile |
| POST | `/orders` | client, admin | Create order |
| GET | `/orders` | admin, dispatcher, courier | List orders (filterable) |
| GET | `/orders/my` | client | Get own orders |
| GET | `/orders/:id` | any | Get order by ID |
| GET | `/orders/:id/history` | any | Get order status history |
| PATCH | `/orders/:id/status` | admin, dispatcher, courier | Update order status |
| GET | `/couriers` | admin, dispatcher | List all couriers |
| GET | `/couriers/free` | admin, dispatcher | List free couriers |
| GET | `/couriers/me` | courier | Get own courier profile |
| GET | `/couriers/me/orders` | courier | Get assigned orders |
| PATCH | `/couriers/me/status` | courier | Update own status |
| PATCH | `/couriers/me/transport` | courier | Update own transport type |
| PATCH | `/couriers/me/location` | courier | Update own location |
| PATCH | `/couriers/:id/status` | admin, dispatcher | Update courier status |
| PATCH | `/couriers/:id/transport` | admin, dispatcher | Update courier transport |
| PATCH | `/couriers/:id/location` | admin, dispatcher | Update courier location |
| GET | `/metrics` | — | Prometheus metrics |

**Order status transitions:**

```
pending ──► assigned ──► in_progress ──► delivered
   └────────────────────────────────────► cancelled
```

---

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Run with Docker Compose

```bash
git clone <repo-url>
cd swift-gopher

# Copy environment file
cp .env.example .env

# Start all services (backend, postgres, redis, prometheus, grafana)
docker compose up --build
```

Services:

| Service | URL |
|---------|-----|
| Backend API | http://localhost:8081 |
| Grafana | http://localhost:3001 |
| Prometheus | http://localhost:9091 |
| PostgreSQL | localhost:5433 |
| Redis | localhost:6380 |

### Seed Demo Data

Run the migration seed (applied automatically on startup) or manually apply `frontend/seed_with_known_passwords.sql`.

Demo accounts:

| Role | Email | Password |
|------|-------|----------|
| Admin | admin@swiftgopher.io | admin123 |
| Dispatcher | dispatcher@swiftgopher.io | disp123 |
| Courier | courier1@swiftgopher.io | courier123 |
| Client | client1@swiftgopher.io | client123 |

### Run Backend Locally (without Docker)

```bash
cd backend

# Set environment variables (see .env.example)
export DB_HOST=localhost
export DB_PORT=5433
export DB_USERNAME=testuser
export DB_PASSWORD=1234
export DB_NAME=testdb
export JWT_SECRET=change-me-in-production-min-32-chars

go run ./cmd/app/main.go
```

### Run Frontend

```bash
cd frontend
npm install
npm start
# Available at http://localhost:4200
# Proxies /api → http://localhost:8081
```

---

## Running Tests

```bash
cd backend

# Run all tests
go test ./tests/... -v

# Run with coverage
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Test files and what they cover:

| File | Coverage |
|------|----------|
| `auth_usecase_test.go` | Register, login, refresh, token validation |
| `order_service_test.go` | Full order lifecycle, status transitions, history |
| `courier_test.go` | Courier CRUD, status updates |
| `dispatcher_test.go` | Assignment logic, nearest-courier selection |
| `pool_test.go` | Worker pool behaviour |
| `retry_test.go` | Retry with backoff, non-retryable errors |
| `middleware_test.go` | CORS, rate limiting, idempotency |
| `blacklist_test.go` | Token blacklist TTL and revocation |
| `cache_test.go` | In-memory cache with expiry |
| `assignment_test.go` | Assignment repository integration |
| `handler_test.go` | HTTP handler integration (health, register, login, auth guard) |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENV` | `development` | `development` or `production` |
| `HTTP_PORT` | `8080` | Server listen port |
| `JWT_SECRET` | — | HS256 signing key (min 32 chars) |
| `ACCESS_TOKEN_TTL` | `120m` | Access token lifetime |
| `REFRESH_TOKEN_TTL` | `168h` | Refresh token lifetime |
| `DB_HOST` | `db` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USERNAME` | — | PostgreSQL user |
| `DB_PASSWORD` | — | PostgreSQL password |
| `DB_NAME` | — | PostgreSQL database name |
| `WORKER_INTERVAL` | `30s` | Dispatcher polling interval |
| `GF_SECURITY_ADMIN_USER` | `admin` | Grafana admin user |
| `GF_SECURITY_ADMIN_PASSWORD` | — | Grafana admin password |

---

## Monitoring

Prometheus scrapes `GET /metrics` every 10 seconds. Custom metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `api_total_http_requests` | Counter | All HTTP requests |
| `api_total_ok_requests` | Counter (path, status) | Successful requests |
| `api_total_error_requests` | Counter (path, status) | Error requests (4xx/5xx) |
| `api_total_orders` | Counter | Orders created (201 on /orders) |

Grafana is pre-configured with a Prometheus datasource at `http://prom:9090`. Log in at http://localhost:3001 with `admin` / `12345678`.
