package tests

import (
	"context"
	"io"
	"log/slog"
	"swift-gopher/internal/worker"
	"sync"
	"testing"
	"time"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

type fakeOrderUC struct {
	orders []*modules.Order
}

func (f *fakeOrderUC) ListPendingOrders(ctx context.Context) ([]*modules.Order, error) {
	return f.orders, nil
}

func (f *fakeOrderUC) UpdateStatus(ctx context.Context, id string, req modules.UpdateOrderStatusRequest) (*modules.Order, error) {
	return &modules.Order{ID: id, Status: req.Status}, nil
}

func (f *fakeOrderUC) CreateOrder(ctx context.Context, clientID string, req modules.CreateOrderRequest) (*modules.Order, error) {
	return nil, nil
}

func (f *fakeOrderUC) GetOrder(ctx context.Context, id string) (*modules.Order, error) {
	return nil, nil
}

func (f *fakeOrderUC) ListOrders(ctx context.Context) ([]*modules.Order, error) {
	return nil, nil
}

func (f *fakeOrderUC) ListOrdersFiltered(ctx context.Context, filter modules.OrderFilter) ([]*modules.Order, error) {
	return nil, nil
}

func (f *fakeOrderUC) GetOrderHistory(ctx context.Context, orderID string) ([]*modules.OrderHistory, error) {
	return nil, nil
}

func (f *fakeOrderUC) GetMyOrders(ctx context.Context, userId string) ([]*modules.Order, error) {
	return nil, nil
}

type fakeCourierUC struct {
	couriers []*modules.Courier
}

func (f *fakeCourierUC) ListFreeCouriers(ctx context.Context) ([]*modules.Courier, error) {
	return f.couriers, nil
}

func (f *fakeCourierUC) UpdateStatus(ctx context.Context, id string, req usecase.UpdateStatusRequest) (*modules.Courier, error) {
	return &modules.Courier{ID: id, Status: req.Status}, nil
}

func (f *fakeCourierUC) UpdateTransport(ctx context.Context, id string, req usecase.UpdateTransportRequest) (*modules.Courier, error) {
	return nil, nil
}

func (f *fakeCourierUC) UpdateLocation(ctx context.Context, id string, req usecase.UpdateLocationRequest) (*modules.Courier, error) {
	return nil, nil
}

func (f *fakeCourierUC) GetCourier(ctx context.Context, id string) (*modules.Courier, error) {
	return nil, nil
}

func (f *fakeCourierUC) ListCouriers(ctx context.Context) ([]*modules.Courier, error) {
	return nil, nil
}

type fakeRepo struct {
	mu    sync.Mutex
	calls int
}

func (f *fakeRepo) Create(ctx context.Context, a *modules.Assignment) error {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	return nil
}

func (f *fakeRepo) GetByOrderID(ctx context.Context, orderID string) (*modules.Assignment, error) {
	return nil, nil
}

func (f *fakeRepo) Complete(ctx context.Context, orderID string, t time.Time) error {
	return nil
}

func TestDispatcher_AssignSuccess(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	repo := &fakeRepo{}

	orderUC := &fakeOrderUC{
		orders: []*modules.Order{
			{ID: "o1"},
			{ID: "o2"},
		},
	}

	courierUC := &fakeCourierUC{
		couriers: []*modules.Courier{
			{ID: "c1"},
			{ID: "c2"},
		},
	}

	d := worker.NewDispatcher(orderUC, courierUC, repo, 1*time.Second, log)

	ctx := context.Background()

	d.Dispatch(ctx)

	repo.mu.Lock()
	c := repo.calls
	repo.mu.Unlock()

	if c != 2 {
		t.Fatalf("expected 2 assignments, got %d", c)
	}
}
