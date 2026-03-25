package modules

import "time"

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusAssigned   OrderStatus = "assigned"
	OrderStatusInProgress OrderStatus = "in_progress"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID              string      `json:"id"`
	ClientID        string      `json:"client_id"`
	PickupAddress   string      `json:"pickup_address"`
	DeliveryAddress string      `json:"delivery_address"`
	Status          OrderStatus `json:"status"`
	Price           float64     `json:"price"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderHistory struct {
	ID        string      `json:"id"`
	OrderID   string      `json:"order_id"`
	OldStatus OrderStatus `json:"old_status"`
	NewStatus OrderStatus `json:"new_status"`
	ChangedAt time.Time   `json:"changed_at"`
}

type CreateOrderRequest struct {
	PickupAddress   string  `json:"pickup_address"`
	DeliveryAddress string  `json:"delivery_address"`
	Price           float64 `json:"price"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
}