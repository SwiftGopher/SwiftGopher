package modules

import (
	"time"
)

type Assignment struct {
	ID          string     `json:"id"`
	OrderID     string     `json:"order_id"`
	CourierID   string     `json:"courier_id"`
	AssignedAt  time.Time  `json:"assigned_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
