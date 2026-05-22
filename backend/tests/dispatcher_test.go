package tests

import (
	"context"
	"io"
	"log/slog"
	"sort"
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

func (f *fakeOrderUC) GetCourierOrders(ctx context.Context, courierID string) ([]*modules.Order, error) {
	return nil, nil
}

type fakeCourierUC struct {
	mu       sync.Mutex
	couriers []*modules.Courier
	assigned map[string]string
}

func newFakeCourierUC(couriers []*modules.Courier) *fakeCourierUC {
	return &fakeCourierUC{
		couriers: couriers,
		assigned: make(map[string]string),
	}
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

func (f *fakeCourierUC) GetCourierByUserID(ctx context.Context, userID string) (*modules.Courier, error) {
	return nil, nil
}

func (f *fakeCourierUC) ListCouriers(ctx context.Context) ([]*modules.Courier, error) {
	return nil, nil
}

func (f *fakeCourierUC) FindNearestCourier(ctx context.Context, lat, lng float64) (*modules.Courier, error) {
	return nil, nil
}

type fakeRepo struct {
	mu          sync.Mutex
	calls       int
	assignments []*modules.Assignment
}

func (f *fakeRepo) Create(ctx context.Context, a *modules.Assignment) error {
	f.mu.Lock()
	f.calls++
	f.assignments = append(f.assignments, a)
	f.mu.Unlock()
	return nil
}

func (f *fakeRepo) GetByOrderID(ctx context.Context, orderID string) (*modules.Assignment, error) {
	return nil, nil
}

func (f *fakeRepo) Complete(ctx context.Context, orderID string, t time.Time) error {
	return nil
}

func (f *fakeRepo) GetByCourierID(ctx context.Context, courierID string) ([]*modules.Assignment, error) {
	return nil, nil
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

	courierUC := newFakeCourierUC([]*modules.Courier{
		{ID: "c1"},
		{ID: "c2"},
	})

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

func TestDispatcher_NearestCourier(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	repo := &fakeRepo{}

	orderA := &modules.Order{ID: "oA", PickupLat: 0, PickupLng: 0}
	orderB := &modules.Order{ID: "oB", PickupLat: 10, PickupLng: 10}

	c1 := &modules.Courier{ID: "c1", CurrentLat: 0, CurrentLng: 0.1}
	c2 := &modules.Courier{ID: "c2", CurrentLat: 10, CurrentLng: 10.1}

	orderUC := &fakeOrderUC{orders: []*modules.Order{orderA, orderB}}
	courierUC := newFakeCourierUC([]*modules.Courier{c2, c1})

	d := worker.NewDispatcher(orderUC, courierUC, repo, 1*time.Second, log)
	d.Dispatch(context.Background())

	repo.mu.Lock()
	got := make([]*modules.Assignment, len(repo.assignments))
	copy(got, repo.assignments)
	repo.mu.Unlock()

	if len(got) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(got))
	}

	result := make(map[string]string, 2)
	for _, a := range got {
		result[a.OrderID] = a.CourierID
	}

	keys := []string{"oA", "oB"}
	sort.Strings(keys)

	if result["oA"] != "c1" {
		t.Errorf("order oA: want courier c1, got %s", result["oA"])
	}
	if result["oB"] != "c2" {
		t.Errorf("order oB: want courier c2, got %s", result["oB"])
	}
}
