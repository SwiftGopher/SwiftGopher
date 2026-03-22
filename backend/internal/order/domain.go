package order

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusAssigned   Status = "assigned"
	StatusInProgress Status = "in_progress"
	StatusDelivered  Status = "delivered"
	StatusCancelled  Status = "cancelled"
)

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrInvalidStatus  = errors.New("invalid status transition")
	ErrInvalidPrice   = errors.New("price must be greater than 0")
	ErrMissingAddress = errors.New("pickup and delivery addresses are required")
)

type Order struct {
	ID              string    `json:"id"`
	ClientID        string    `json:"client_id"`
	PickupAddress   string    `json:"pickup_address"`
	DeliveryAddress string    `json:"delivery_address"`
	Status          Status    `json:"status"`
	Price           float64   `json:"price"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OrderHistory struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	OldStatus Status    `json:"old_status"`
	NewStatus Status    `json:"new_status"`
	ChangedAt time.Time `json:"changed_at"`
}

type CreateRequest struct {
	PickupAddress   string  `json:"pickup_address"`
	DeliveryAddress string  `json:"delivery_address"`
	Price           float64 `json:"price"`
}

type UpdateStatusRequest struct {
	Status Status `json:"status"`
}

type Repository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	List(ctx context.Context) ([]*Order, error)
	ListByStatus(ctx context.Context, status Status) ([]*Order, error)
	UpdateStatus(ctx context.Context, id string, status Status) error
	RecordHistory(ctx context.Context, history *OrderHistory) error
}

type Service interface {
	CreateOrder(ctx context.Context, clientID string, req CreateRequest) (*Order, error)
	GetOrder(ctx context.Context, id string) (*Order, error)
	ListOrders(ctx context.Context) ([]*Order, error)
	ListPendingOrders(ctx context.Context) ([]*Order, error)
	UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*Order, error)
}
