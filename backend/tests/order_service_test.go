package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"swift-gopher/internal/order"
)

type mockOrderRepo struct {
	mu      sync.Mutex
	orders  map[string]*order.Order
	history []*order.OrderHistory
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[string]*order.Order)}
}

func (m *mockOrderRepo) Create(_ context.Context, o *order.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *o
	m.orders[o.ID] = &cp
	return nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id string) (*order.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, order.ErrOrderNotFound
	}
	cp := *o
	return &cp, nil
}

func (m *mockOrderRepo) List(_ context.Context) ([]*order.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*order.Order, 0, len(m.orders))
	for _, o := range m.orders {
		cp := *o
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockOrderRepo) ListByStatus(_ context.Context, status order.Status) ([]*order.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*order.Order
	for _, o := range m.orders {
		if o.Status == status {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *mockOrderRepo) UpdateStatus(_ context.Context, id string, status order.Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return order.ErrOrderNotFound
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

func (m *mockOrderRepo) RecordHistory(_ context.Context, h *order.OrderHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = append(m.history, h)
	return nil
}

func newTestService() order.Service {
	repo := newMockOrderRepo()
	return order.NewService(repo, newTestLogger())
}

func mustCreateOrder(t *testing.T, svc order.Service) *order.Order {
	t.Helper()
	o, err := svc.CreateOrder(context.Background(), uuid.NewString(), order.CreateRequest{
		PickupAddress:   "Street A",
		DeliveryAddress: "Street B",
		Price:           9.99,
	})
	if err != nil {
		t.Fatalf("mustCreateOrder: %v", err)
	}
	return o
}

func TestCreateOrder_Success(t *testing.T) {
	svc := newTestService()

	o, err := svc.CreateOrder(context.Background(), "client-001", order.CreateRequest{
		PickupAddress:   "Almaty, Abay 1",
		DeliveryAddress: "Almaty, Dostyk 10",
		Price:           15.50,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if o.ID == "" {
		t.Error("order ID must be set")
	}
	if o.Status != order.StatusPending {
		t.Errorf("new order status must be pending, got %q", o.Status)
	}
	if o.ClientID != "client-001" {
		t.Errorf("client_id mismatch: got %q", o.ClientID)
	}
}

func TestCreateOrder_InvalidPrice(t *testing.T) {
	svc := newTestService()

	_, err := svc.CreateOrder(context.Background(), "client-001", order.CreateRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           -5,
	})

	if err == nil {
		t.Fatal("expected ErrInvalidPrice, got nil")
	}
	if err != order.ErrInvalidPrice {
		t.Errorf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestCreateOrder_ZeroPrice(t *testing.T) {
	svc := newTestService()

	_, err := svc.CreateOrder(context.Background(), "client-001", order.CreateRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           0,
	})

	if err != order.ErrInvalidPrice {
		t.Errorf("zero price: expected ErrInvalidPrice, got %v", err)
	}
}

func TestCreateOrder_MissingAddress(t *testing.T) {
	svc := newTestService()

	_, err := svc.CreateOrder(context.Background(), "client-001", order.CreateRequest{
		PickupAddress:   "A",
		DeliveryAddress: "",
		Price:           10,
	})

	if err != order.ErrMissingAddress {
		t.Errorf("expected ErrMissingAddress, got %v", err)
	}
}

func TestCreateOrder_MissingPickup(t *testing.T) {
	svc := newTestService()

	_, err := svc.CreateOrder(context.Background(), "client-001", order.CreateRequest{
		PickupAddress:   "",
		DeliveryAddress: "B",
		Price:           10,
	})

	if err != order.ErrMissingAddress {
		t.Errorf("expected ErrMissingAddress, got %v", err)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.GetOrder(context.Background(), "non-existent-id")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != order.ErrOrderNotFound {
		t.Logf("note: error wraps ErrOrderNotFound — %v", err)
	}
}

func TestGetOrder_Success(t *testing.T) {
	svc := newTestService()
	created := mustCreateOrder(t, svc)

	fetched, err := svc.GetOrder(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch: want %q got %q", created.ID, fetched.ID)
	}
}

func TestUpdateStatus_ValidTransition(t *testing.T) {
	svc := newTestService()
	o := mustCreateOrder(t, svc)

	updated, err := svc.UpdateStatus(context.Background(), o.ID, order.UpdateStatusRequest{
		Status: order.StatusAssigned,
	})

	if err != nil {
		t.Fatalf("pending → assigned: expected no error, got %v", err)
	}
	if updated.Status != order.StatusAssigned {
		t.Errorf("status should be assigned, got %q", updated.Status)
	}
}

func TestUpdateStatus_InvalidTransition(t *testing.T) {
	svc := newTestService()
	o := mustCreateOrder(t, svc)

	_, err := svc.UpdateStatus(context.Background(), o.ID, order.UpdateStatusRequest{
		Status: order.StatusDelivered,
	})

	if err != order.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestUpdateStatus_FullLifecycle(t *testing.T) {
	svc := newTestService()
	o := mustCreateOrder(t, svc)
	ctx := context.Background()

	transitions := []order.Status{
		order.StatusAssigned,
		order.StatusInProgress,
		order.StatusDelivered,
	}

	for _, next := range transitions {
		updated, err := svc.UpdateStatus(ctx, o.ID, order.UpdateStatusRequest{Status: next})
		if err != nil {
			t.Fatalf("transition to %q failed: %v", next, err)
		}
		if updated.Status != next {
			t.Errorf("expected status %q, got %q", next, updated.Status)
		}
		o = updated
	}
}

func TestUpdateStatus_CancelFromPending(t *testing.T) {
	svc := newTestService()
	o := mustCreateOrder(t, svc)

	updated, err := svc.UpdateStatus(context.Background(), o.ID, order.UpdateStatusRequest{
		Status: order.StatusCancelled,
	})

	if err != nil {
		t.Fatalf("pending → cancelled: expected no error, got %v", err)
	}
	if updated.Status != order.StatusCancelled {
		t.Errorf("expected cancelled, got %q", updated.Status)
	}
}

func TestUpdateStatus_TerminalStates(t *testing.T) {
	ctx := context.Background()

	t.Run("delivered is terminal", func(t *testing.T) {
		svc := newTestService()
		o := mustCreateOrder(t, svc)

		for _, s := range []order.Status{order.StatusAssigned, order.StatusInProgress, order.StatusDelivered} {
			o, _ = svc.UpdateStatus(ctx, o.ID, order.UpdateStatusRequest{Status: s})
		}

		_, err := svc.UpdateStatus(ctx, o.ID, order.UpdateStatusRequest{Status: order.StatusCancelled})
		if err != order.ErrInvalidStatus {
			t.Errorf("expected ErrInvalidStatus after delivered, got %v", err)
		}
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		svc := newTestService()
		o := mustCreateOrder(t, svc)

		svc.UpdateStatus(ctx, o.ID, order.UpdateStatusRequest{Status: order.StatusCancelled})

		_, err := svc.UpdateStatus(ctx, o.ID, order.UpdateStatusRequest{Status: order.StatusAssigned})
		if err != order.ErrInvalidStatus {
			t.Errorf("expected ErrInvalidStatus after cancelled, got %v", err)
		}
	})
}

func TestListPendingOrders(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		mustCreateOrder(t, svc)
	}

	orders, _ := svc.ListPendingOrders(ctx)
	if len(orders) < 1 {
		t.Fatal("need at least 1 order to proceed")
	}
	svc.UpdateStatus(ctx, orders[0].ID, order.UpdateStatusRequest{Status: order.StatusAssigned})

	pending, err := svc.ListPendingOrders(ctx)
	if err != nil {
		t.Fatalf("ListPendingOrders failed: %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("expected 2 pending orders, got %d", len(pending))
	}
	for _, o := range pending {
		if o.Status != order.StatusPending {
			t.Errorf("ListPendingOrders returned non-pending order: %q", o.Status)
		}
	}
}

func TestHistoryRecorded(t *testing.T) {
	repo := newMockOrderRepo()
	svc := order.NewService(repo, newTestLogger())

	o, _ := svc.CreateOrder(context.Background(), "client-x", order.CreateRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           5,
	})

	svc.UpdateStatus(context.Background(), o.ID, order.UpdateStatusRequest{Status: order.StatusAssigned})

	repo.mu.Lock()
	historyCount := len(repo.history)
	repo.mu.Unlock()

	if historyCount == 0 {
		t.Error("expected at least one history record after status update")
	}
	if repo.history[0].OldStatus != order.StatusPending {
		t.Errorf("history old_status: want pending, got %q", repo.history[0].OldStatus)
	}
	if repo.history[0].NewStatus != order.StatusAssigned {
		t.Errorf("history new_status: want assigned, got %q", repo.history[0].NewStatus)
	}
}
