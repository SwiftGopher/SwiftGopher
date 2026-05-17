# SwiftGopher Backend API Endpoints

This document describes all HTTP endpoints registered by the backend application in `internal/handler/handler.go`.

## Authentication

### `GET /health`
- Description: Health check endpoint
- Auth: none
- Response:
  - `200 OK` `{ "status": "ok" }`

### `POST /auth/register`
- Description: Register a new user
- Auth: none
- Request JSON:
  - `email` (string, required, email format)
  - `password` (string, required, min 6 chars)
  - `role` (string, optional)
    - Allowed values: `admin`, `dispatcher`, `courier`, `client`
    - Default: `client` when omitted
- Responses:
  - `201 Created` - returns created `User`
  - `400 Bad Request` - invalid payload or invalid role
  - `409 Conflict` - email already taken

### `POST /auth/login`
- Description: Log in and obtain JWT access/refresh tokens
- Auth: none
- Request JSON:
  - `email` (string, required)
  - `password` (string, required)
- Responses:
  - `200 OK` - returns `TokenPair`
  - `400 Bad Request` - invalid payload
  - `401 Unauthorized` - invalid credentials

### `POST /auth/refresh`
- Description: Refresh access token with a refresh token
- Auth: none
- Request JSON:
  - `refresh_token` (string, required)
- Responses:
  - `200 OK` - returns `TokenPair`
  - `400 Bad Request` - invalid payload
  - `401 Unauthorized` - invalid refresh token

## Protected routes

All protected routes require the `Authorization` header:

```
Authorization: Bearer <access_token>
```

The backend uses JWT validation via `internal/middleware/auth.go`.

### `GET /profile`
- Description: Get the authenticated user's own profile
- Auth: required
- Roles: any authenticated user
- Response:
  - `200 OK` - returns user profile object

## Orders

### `POST /orders`
- Description: Create a new order
- Auth: required
- Roles: `client`, `admin`
- Request JSON:
  - `pickup_address` (string)
  - `delivery_address` (string)
  - `price` (number)
- Responses:
  - `201 Created` - returns created `Order`
  - `400 Bad Request` - invalid payload, missing address, or invalid price

### `GET /orders`
- Description: List orders with optional filtering and pagination
- Auth: required
- Roles: `admin`, `dispatcher`, `courier`
- Query parameters:
  - `status` - order status filter (`pending`, `assigned`, `in_progress`, `delivered`, `cancelled`)
  - `limit` - integer pagination limit
  - `offset` - integer pagination offset
  - `sort_by` - sort field
  - `sort_dir` - sort direction
- Response:
  - `200 OK` - JSON object containing `data`, `limit`, and `offset`

### `GET /orders/my`
- Description: Get orders for the authenticated client user
- Auth: required
- Roles: `client`
- Response:
  - `200 OK` - returns an array of orders for the current user
  - `404 Not Found` - when no orders exist for the user

### `GET /orders/:id`
- Description: Retrieve a single order by ID
- Auth: required
- Roles: any authenticated user
- Path parameters:
  - `id` - order ID
- Responses:
  - `200 OK` - returns the `Order`
  - `400 Bad Request` - missing order id
  - `404 Not Found` - order not found

### `GET /orders/:id/history`
- Description: Retrieve status history for an order
- Auth: required
- Roles: any authenticated user
- Path parameters:
  - `id` - order ID
- Responses:
  - `200 OK` - returns array of `OrderHistory`
  - `400 Bad Request` - missing order id
  - `404 Not Found` - order not found

### `PATCH /orders/:id/status`
- Description: Update an order's status
- Auth: required
- Roles: `admin`, `dispatcher`, `courier`
- Path parameters:
  - `id` - order ID
- Request JSON:
  - `status` (string)
    - Allowed values: `pending`, `assigned`, `in_progress`, `delivered`, `cancelled`
- Responses:
  - `200 OK` - returns updated `Order`
  - `400 Bad Request` - invalid payload or missing id
  - `404 Not Found` - order not found
  - `422 Unprocessable Entity` - invalid order status

## Couriers

### `GET /couriers`
- Description: List all couriers
- Auth: required
- Roles: `admin`, `dispatcher`
- Response:
  - `200 OK` - returns an array of `Courier`

### `GET /couriers/free`
- Description: List all free couriers
- Auth: required
- Roles: `admin`, `dispatcher`
- Response:
  - `200 OK` - returns an array of `Courier`

### `PATCH /couriers/:id/status`
- Description: Update courier status
- Auth: required
- Roles: `admin`, `dispatcher`, `courier`
- Path parameters:
  - `id` - courier ID
- Request JSON:
  - `status` (string)
    - Allowed values: `free`, `busy`, `offline`
- Responses:
  - `200 OK` - returns updated `Courier`
  - `400 Bad Request` - invalid payload
  - `404 Not Found` - courier not found

### `PATCH /couriers/:id/transport`
- Description: Update courier transport type
- Auth: required
- Roles: `admin`, `dispatcher`, `courier`
- Path parameters:
  - `id` - courier ID
- Request JSON:
  - `transport_type` (string)
    - Allowed values: `bike`, `car`, `foot`, `scooter`
- Responses:
  - `200 OK` - returns updated `Courier`
  - `400 Bad Request` - invalid payload
  - `404 Not Found` - courier not found

### `PATCH /couriers/:id/location`
- Description: Update courier location coordinates
- Auth: required
- Roles: `admin`, `dispatcher`, `courier`
- Path parameters:
  - `id` - courier ID
- Request JSON:
  - `lat` (number)
  - `lng` (number)
- Responses:
  - `200 OK` - returns updated `Courier`
  - `400 Bad Request` - invalid payload
  - `404 Not Found` - courier not found

## Data models

### `User`
- `id` (string)
- `email` (string)
- `role` (string)
- `created_at` (timestamp)

### `TokenPair`
- `access_token` (string)
- `refresh_token` (string)

### `Order`
- `id` (string)
- `client_id` (string)
- `pickup_address` (string)
- `delivery_address` (string)
- `status` (string)
- `price` (number)
- `created_at` (timestamp)
- `updated_at` (timestamp)

### `OrderHistory`
- `id` (string)
- `order_id` (string)
- `old_status` (string)
- `new_status` (string)
- `changed_at` (timestamp)

### `Courier`
- `id` (string)
- `user_id` (string)
- `transport_type` (string)
- `status` (string)
- `current_lat` (number)
- `current_lng` (number)
