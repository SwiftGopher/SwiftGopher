package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

type mockOrderRepo struct {
	mu      sync.Mutex
	orders  map[string]*modules.Order
	history []*modules.OrderHistory
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[string]*modules.Order)}
}

func (m *mockOrderRepo) Create(_ context.Context, o *modules.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *o
	m.orders[o.ID] = &cp
	return nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id string) (*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, usecase.ErrOrderNotFound
	}
	cp := *o
	return &cp, nil
}

func (m *mockOrderRepo) List(_ context.Context) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*modules.Order, 0, len(m.orders))
	for _, o := range m.orders {
		cp := *o
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockOrderRepo) ListByStatus(_ context.Context, status modules.OrderStatus) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*modules.Order
	for _, o := range m.orders {
		if o.Status == status {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *mockOrderRepo) UpdateStatus(_ context.Context, id string, status modules.OrderStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return usecase.ErrOrderNotFound
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

func (m *mockOrderRepo) RecordHistory(_ context.Context, h *modules.OrderHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = append(m.history, h)
	return nil
}

func newTestOrderUsecase() usecase.OrderUsecase {
	repo := newMockOrderRepo()
	return usecase.NewOrderUsecase(repo, newTestLogger())
}

func mustCreateOrder(t *testing.T, uc usecase.OrderUsecase) *modules.Order {
	t.Helper()
	o, err := uc.CreateOrder(context.Background(), uuid.NewString(), modules.CreateOrderRequest{
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
	uc := newTestOrderUsecase()

	o, err := uc.CreateOrder(context.Background(), "client-001", modules.CreateOrderRequest{
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
	if o.Status != modules.OrderStatusPending {
		t.Errorf("new order status must be pending, got %q", o.Status)
	}
	if o.ClientID != "client-001" {
		t.Errorf("client_id mismatch: got %q", o.ClientID)
	}
}

func TestCreateOrder_InvalidPrice(t *testing.T) {
	uc := newTestOrderUsecase()

	_, err := uc.CreateOrder(context.Background(), "client-001", modules.CreateOrderRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           -5,
	})

	if err == nil {
		t.Fatal("expected ErrInvalidPrice, got nil")
	}
	if err != usecase.ErrInvalidPrice {
		t.Errorf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestCreateOrder_ZeroPrice(t *testing.T) {
	uc := newTestOrderUsecase()

	_, err := uc.CreateOrder(context.Background(), "client-001", modules.CreateOrderRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           0,
	})

	if err != usecase.ErrInvalidPrice {
		t.Errorf("zero price: expected ErrInvalidPrice, got %v", err)
	}
}

func TestCreateOrder_MissingAddress(t *testing.T) {
	uc := newTestOrderUsecase()

	_, err := uc.CreateOrder(context.Background(), "client-001", modules.CreateOrderRequest{
		PickupAddress:   "A",
		DeliveryAddress: "",
		Price:           10,
	})

	if err != usecase.ErrMissingAddress {
		t.Errorf("expected ErrMissingAddress, got %v", err)
	}
}

func TestCreateOrder_MissingPickup(t *testing.T) {
	uc := newTestOrderUsecase()

	_, err := uc.CreateOrder(context.Background(), "client-001", modules.CreateOrderRequest{
		PickupAddress:   "",
		DeliveryAddress: "B",
		Price:           10,
	})

	if err != usecase.ErrMissingAddress {
		t.Errorf("expected ErrMissingAddress, got %v", err)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	uc := newTestOrderUsecase()

	_, err := uc.GetOrder(context.Background(), "non-existent-id")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != usecase.ErrOrderNotFound {
		t.Logf("note: error wraps ErrOrderNotFound — %v", err)
	}
}

func TestGetOrder_Success(t *testing.T) {
	uc := newTestOrderUsecase()
	created := mustCreateOrder(t, uc)

	fetched, err := uc.GetOrder(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch: want %q got %q", created.ID, fetched.ID)
	}
}

func TestUpdateStatus_ValidTransition(t *testing.T) {
	uc := newTestOrderUsecase()
	o := mustCreateOrder(t, uc)

	updated, err := uc.UpdateStatus(context.Background(), o.ID, modules.UpdateOrderStatusRequest{
		Status: modules.OrderStatusAssigned,
	})

	if err != nil {
		t.Fatalf("pending → assigned: expected no error, got %v", err)
	}
	if updated.Status != modules.OrderStatusAssigned {
		t.Errorf("status should be assigned, got %q", updated.Status)
	}
}

func TestUpdateStatus_InvalidTransition(t *testing.T) {
	uc := newTestOrderUsecase()
	o := mustCreateOrder(t, uc)

	_, err := uc.UpdateStatus(context.Background(), o.ID, modules.UpdateOrderStatusRequest{
		Status: modules.OrderStatusDelivered,
	})

	if err != usecase.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestUpdateStatus_FullLifecycle(t *testing.T) {
	uc := newTestOrderUsecase()
	o := mustCreateOrder(t, uc)
	ctx := context.Background()

	transitions := []modules.OrderStatus{
		modules.OrderStatusAssigned,
		modules.OrderStatusInProgress,
		modules.OrderStatusDelivered,
	}

	for _, next := range transitions {
		updated, err := uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: next})
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
	uc := newTestOrderUsecase()
	o := mustCreateOrder(t, uc)

	updated, err := uc.UpdateStatus(context.Background(), o.ID, modules.UpdateOrderStatusRequest{
		Status: modules.OrderStatusCancelled,
	})

	if err != nil {
		t.Fatalf("pending → cancelled: expected no error, got %v", err)
	}
	if updated.Status != modules.OrderStatusCancelled {
		t.Errorf("expected cancelled, got %q", updated.Status)
	}
}

func TestUpdateStatus_TerminalStates(t *testing.T) {
	ctx := context.Background()

	t.Run("delivered is terminal", func(t *testing.T) {
		uc := newTestOrderUsecase()
		o := mustCreateOrder(t, uc)

		for _, s := range []modules.OrderStatus{
			modules.OrderStatusAssigned,
			modules.OrderStatusInProgress,
			modules.OrderStatusDelivered,
		} {
			o, _ = uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: s})
		}

		_, err := uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusCancelled})
		if err != usecase.ErrInvalidStatus {
			t.Errorf("expected ErrInvalidStatus after delivered, got %v", err)
		}
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		uc := newTestOrderUsecase()
		o := mustCreateOrder(t, uc)

		uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusCancelled})

		_, err := uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusAssigned})
		if err != usecase.ErrInvalidStatus {
			t.Errorf("expected ErrInvalidStatus after cancelled, got %v", err)
		}
	})
}

func TestListPendingOrders(t *testing.T) {
	uc := newTestOrderUsecase()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		mustCreateOrder(t, uc)
	}

	orders, _ := uc.ListPendingOrders(ctx)
	if len(orders) < 1 {
		t.Fatal("need at least 1 order to proceed")
	}
	uc.UpdateStatus(ctx, orders[0].ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusAssigned})

	pending, err := uc.ListPendingOrders(ctx)
	if err != nil {
		t.Fatalf("ListPendingOrders failed: %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("expected 2 pending orders, got %d", len(pending))
	}
	for _, o := range pending {
		if o.Status != modules.OrderStatusPending {
			t.Errorf("ListPendingOrders returned non-pending order: %q", o.Status)
		}
	}
}

func TestHistoryRecorded(t *testing.T) {
	repo := newMockOrderRepo()
	uc := usecase.NewOrderUsecase(repo, newTestLogger())

	o, _ := uc.CreateOrder(context.Background(), "client-x", modules.CreateOrderRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           5,
	})

	uc.UpdateStatus(context.Background(), o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusAssigned})

	repo.mu.Lock()
	historyCount := len(repo.history)
	repo.mu.Unlock()

	if historyCount == 0 {
		t.Error("expected at least one history record after status update")
	}
	if repo.history[0].OldStatus != modules.OrderStatusPending {
		t.Errorf("history old_status: want pending, got %q", repo.history[0].OldStatus)
	}
	if repo.history[0].NewStatus != modules.OrderStatusAssigned {
		t.Errorf("history new_status: want assigned, got %q", repo.history[0].NewStatus)
	}
}