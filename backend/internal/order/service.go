package order

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type service struct {
	repo Repository
	log  *slog.Logger
}

func NewService(repo Repository, log *slog.Logger) Service {
	return &service{repo: repo, log: log}
}

func isValidTransition(from, to Status) bool {
	allowed := map[Status][]Status{
		StatusPending:    {StatusAssigned, StatusCancelled},
		StatusAssigned:   {StatusInProgress, StatusCancelled},
		StatusInProgress: {StatusDelivered, StatusCancelled},
		StatusDelivered:  {},
		StatusCancelled:  {},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}

func (s *service) CreateOrder(ctx context.Context, clientID string, req CreateRequest) (*Order, error) {
	if req.PickupAddress == "" || req.DeliveryAddress == "" {
		return nil, ErrMissingAddress
	}
	if req.Price <= 0 {
		return nil, ErrInvalidPrice
	}

	now := time.Now()
	order := &Order{
		ID:              uuid.NewString(),
		ClientID:        clientID,
		PickupAddress:   req.PickupAddress,
		DeliveryAddress: req.DeliveryAddress,
		Status:          StatusPending,
		Price:           req.Price,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("order.service.CreateOrder: %w", err)
	}

	s.log.Info("order created", "order_id", order.ID, "client_id", clientID)
	return order, nil
}

func (s *service) GetOrder(ctx context.Context, id string) (*Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order.service.GetOrder: %w", err)
	}
	return order, nil
}

func (s *service) ListOrders(ctx context.Context) ([]*Order, error) {
	orders, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("order.service.ListOrders: %w", err)
	}
	return orders, nil
}

func (s *service) ListPendingOrders(ctx context.Context) ([]*Order, error) {
	orders, err := s.repo.ListByStatus(ctx, StatusPending)
	if err != nil {
		return nil, fmt.Errorf("order.service.ListPendingOrders: %w", err)
	}
	return orders, nil
}

func (s *service) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order.service.UpdateStatus: %w", err)
	}

	if !isValidTransition(order.Status, req.Status) {
		return nil, ErrInvalidStatus
	}

	oldStatus := order.Status

	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		return nil, fmt.Errorf("order.service.UpdateStatus: %w", err)
	}

	history := &OrderHistory{
		ID:        uuid.NewString(),
		OrderID:   id,
		OldStatus: oldStatus,
		NewStatus: req.Status,
		ChangedAt: time.Now(),
	}
	if err := s.repo.RecordHistory(ctx, history); err != nil {
		s.log.Warn("failed to record order history", "order_id", id, "error", err)
	}

	order.Status = req.Status
	order.UpdatedAt = history.ChangedAt

	s.log.Info("order status updated",
		"order_id", id,
		"old_status", oldStatus,
		"new_status", req.Status,
	)
	return order, nil
}
