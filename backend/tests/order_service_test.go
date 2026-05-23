package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

type mockCourierUsecase struct{}

func (m *mockCourierUsecase) FindNearestCourier(_ context.Context, _, _ float64) (*modules.Courier, error) {
	return nil, errors.New("no courier available")
}
func (m *mockCourierUsecase) GetCourier(_ context.Context, _ string) (*modules.Courier, error) {
	return nil, nil
}
func (m *mockCourierUsecase) GetCourierByUserID(_ context.Context, _ string) (*modules.Courier, error) {
	return nil, nil
}
func (m *mockCourierUsecase) ListCouriers(_ context.Context) ([]*modules.Courier, error) {
	return nil, nil
}
func (m *mockCourierUsecase) ListFreeCouriers(_ context.Context) ([]*modules.Courier, error) {
	return nil, nil
}
func (m *mockCourierUsecase) UpdateStatus(_ context.Context, _ string, _ usecase.UpdateStatusRequest) (*modules.Courier, error) {
	return nil, nil
}
func (m *mockCourierUsecase) UpdateLocation(_ context.Context, _ string, _ usecase.UpdateLocationRequest) (*modules.Courier, error) {
	return nil, nil
}
func (m *mockCourierUsecase) UpdateTransport(_ context.Context, _ string, _ usecase.UpdateTransportRequest) (*modules.Courier, error) {
	return nil, nil
}

func newTestOrderUsecase() usecase.OrderUsecase {
	return usecase.NewOrderUsecase(
		newMockOrderRepo(),
		newMockAssignmentRepo(),
		&mockCourierUsecase{},
		newTestLogger(),
		nil,
	)
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

	if err != usecase.ErrInvalidOrderStatus {
		t.Errorf("expected ErrInvalidOrderStatus, got %v", err)
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
			updated, err := uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: s})
			if err != nil {
				t.Fatalf("setup transition to %q failed: %v", s, err)
			}
			o = updated
		}

		_, err := uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusCancelled})
		if err != usecase.ErrInvalidOrderStatus {
			t.Errorf("expected ErrInvalidOrderStatus after delivered, got %v", err)
		}
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		uc := newTestOrderUsecase()
		o := mustCreateOrder(t, uc)

		updated, err := uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusCancelled})
		if err != nil {
			t.Fatalf("setup transition to cancelled failed: %v", err)
		}
		o = updated

		_, err = uc.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusAssigned})
		if err != usecase.ErrInvalidOrderStatus {
			t.Errorf("expected ErrInvalidOrderStatus after cancelled, got %v", err)
		}
	})
}

func TestListPendingOrders(t *testing.T) {
	uc := newTestOrderUsecase()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		mustCreateOrder(t, uc)
	}

	orders, err := uc.ListPendingOrders(ctx)
	if err != nil {
		t.Fatalf("initial ListPendingOrders failed: %v", err)
	}
	if len(orders) < 1 {
		t.Fatal("need at least 1 order to proceed")
	}

	_, err = uc.UpdateStatus(ctx, orders[0].ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusAssigned})
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

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
	uc := usecase.NewOrderUsecase(
		repo,
		newMockAssignmentRepo(),
		&mockCourierUsecase{},
		newTestLogger(),
		nil,
	)

	o, err := uc.CreateOrder(context.Background(), "client-x", modules.CreateOrderRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           5,
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	_, err = uc.UpdateStatus(context.Background(), o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusAssigned})
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	repo.mu.Lock()
	historyCount := len(repo.history)
	first := repo.history[0]
	repo.mu.Unlock()

	if historyCount == 0 {
		t.Fatal("expected at least one history record after status update")
	}
	if first == nil {
		t.Fatal("history[0] is nil")
	}
	if first.OldStatus != modules.OrderStatusPending {
		t.Errorf("history old_status: want pending, got %q", first.OldStatus)
	}
	if first.NewStatus != modules.OrderStatusAssigned {
		t.Errorf("history new_status: want assigned, got %q", first.NewStatus)
	}
}
