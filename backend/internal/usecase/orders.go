package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"swift-gopher/internal/repository"
	"swift-gopher/pkg/modules"

	"github.com/google/uuid"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderStatus = errors.New("invalid status transition")
	ErrInvalidPrice       = errors.New("price must be greater than 0")
	ErrMissingAddress     = errors.New("pickup and delivery addresses are required")
)

type orderUsecase struct {
	repo repository.OrderRepository
	log  *slog.Logger
}

func NewOrderUsecase(repo repository.OrderRepository, log *slog.Logger) OrderUsecase {
	return &orderUsecase{repo: repo, log: log}
}

func isValidTransition(from, to modules.OrderStatus) bool {
	allowed := map[modules.OrderStatus][]modules.OrderStatus{
		modules.OrderStatusPending:    {modules.OrderStatusAssigned, modules.OrderStatusCancelled},
		modules.OrderStatusAssigned:   {modules.OrderStatusInProgress, modules.OrderStatusCancelled},
		modules.OrderStatusInProgress: {modules.OrderStatusDelivered, modules.OrderStatusCancelled},
		modules.OrderStatusDelivered:  {},
		modules.OrderStatusCancelled:  {},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}

func (u *orderUsecase) CreateOrder(ctx context.Context, clientID string, req modules.CreateOrderRequest) (*modules.Order, error) {
	if req.PickupAddress == "" || req.DeliveryAddress == "" {
		return nil, ErrMissingAddress
	}
	if req.Price <= 0 {
		return nil, ErrInvalidPrice
	}

	now := time.Now().UTC()
	order := &modules.Order{
		ID:              uuid.NewString(),
		ClientID:        clientID,
		PickupAddress:   req.PickupAddress,
		DeliveryAddress: req.DeliveryAddress,
		Status:          modules.OrderStatusPending,
		Price:           req.Price,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := u.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("orderUsecase.CreateOrder: %w", err)
	}

	u.log.Info("order created", "order_id", order.ID, "client_id", clientID)
	return order, nil
}

func (u *orderUsecase) GetOrder(ctx context.Context, id string) (*modules.Order, error) {
	order, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("orderUsecase.GetOrder: %w", err)
	}
	return order, nil
}

func (u *orderUsecase) ListOrders(ctx context.Context) ([]*modules.Order, error) {
	orders, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("orderUsecase.ListOrders: %w", err)
	}
	return orders, nil
}

func (u *orderUsecase) ListPendingOrders(ctx context.Context) ([]*modules.Order, error) {
	orders, err := u.repo.ListByStatus(ctx, modules.OrderStatusPending)
	if err != nil {
		return nil, fmt.Errorf("orderUsecase.ListPendingOrders: %w", err)
	}
	return orders, nil
}

func (u *orderUsecase) UpdateStatus(ctx context.Context, id string, req modules.UpdateOrderStatusRequest) (*modules.Order, error) {
	order, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("orderUsecase.UpdateStatus: %w", err)
	}

	if !isValidTransition(order.Status, req.Status) {
		return nil, ErrInvalidOrderStatus
	}

	oldStatus := order.Status

	if err := u.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		return nil, fmt.Errorf("orderUsecase.UpdateStatus: %w", err)
	}

	history := &modules.OrderHistory{
		ID:        uuid.NewString(),
		OrderID:   id,
		OldStatus: oldStatus,
		NewStatus: req.Status,
		ChangedAt: time.Now().UTC(),
	}
	if err := u.repo.RecordHistory(ctx, history); err != nil {
		u.log.Warn("failed to record order history", "order_id", id, "error", err)
	}

	order.Status = req.Status
	order.UpdatedAt = history.ChangedAt

	u.log.Info("order status updated", "order_id", id, "old", oldStatus, "new", req.Status)
	return order, nil
}

func (u *orderUsecase) ListOrdersFiltered(ctx context.Context, filter modules.OrderFilter) ([]*modules.Order, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortDir == "" {
		filter.SortDir = "desc"
	}

	orders, err := u.repo.ListWithFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("orderUsecase.ListOrdersFiltered: %w", err)
	}
	return orders, nil
}

func (u *orderUsecase) GetOrderHistory(ctx context.Context, orderID string) ([]*modules.OrderHistory, error) {
	_, err := u.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("orderUsecase.GetOrderHistory: %w", err)
	}

	history, err := u.repo.GetHistory(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("orderUsecase.GetOrderHistory: %w", err)
	}
	return history, nil
}
