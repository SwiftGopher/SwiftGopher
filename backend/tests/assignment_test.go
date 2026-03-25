package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

type mockCourierRepo struct {
	couriers map[string]*modules.Courier
}

func newMockCourierRepo() *mockCourierRepo {
	return &mockCourierRepo{couriers: make(map[string]*modules.Courier)}
}

func (m *mockCourierRepo) GetByID(_ context.Context, id string) (*modules.Courier, error) {
	c, ok := m.couriers[id]
	if !ok {
		return nil, errors.New("courier not found")
	}
	return c, nil
}

func (m *mockCourierRepo) List(_ context.Context) ([]*modules.Courier, error) {
	list := make([]*modules.Courier, 0, len(m.couriers))
	for _, c := range m.couriers {
		list = append(list, c)
	}
	return list, nil
}

func (m *mockCourierRepo) ListFree(_ context.Context) ([]*modules.Courier, error) {
	var free []*modules.Courier
	for _, c := range m.couriers {
		if c.Status == modules.StatusFree {
			free = append(free, c)
		}
	}
	return free, nil
}

func (m *mockCourierRepo) UpdateStatus(_ context.Context, id string, status modules.CourierStatus) error {
	c, ok := m.couriers[id]
	if !ok {
		return errors.New("courier not found")
	}
	c.Status = status
	return nil
}

type mockAssignmentRepo struct {
	assignments map[string]*modules.Assignment
}

func newMockAssignmentRepo() *mockAssignmentRepo {
	return &mockAssignmentRepo{assignments: make(map[string]*modules.Assignment)}
}

func (m *mockAssignmentRepo) Create(_ context.Context, a *modules.Assignment) error {
	m.assignments[a.OrderID] = a
	return nil
}

func (m *mockAssignmentRepo) GetByOrderID(_ context.Context, orderID string) (*modules.Assignment, error) {
	a, ok := m.assignments[orderID]
	if !ok {
		return nil, errors.New("assignment not found")
	}
	return a, nil
}

func (m *mockAssignmentRepo) Complete(_ context.Context, orderID string, completedAt time.Time) error {
	a, ok := m.assignments[orderID]
	if !ok {
		return errors.New("assignment not found")
	}
	a.CompletedAt = &completedAt
	return nil
}

func TestDispatcherAssignsCourier(t *testing.T) {
	ctx := context.Background()

	courierRepo := newMockCourierRepo()
	assignRepo := newMockAssignmentRepo()

	c1 := &modules.Courier{ID: "c1", Status: modules.StatusFree}
	c2 := &modules.Courier{ID: "c2", Status: modules.StatusBusy}
	courierRepo.couriers[c1.ID] = c1
	courierRepo.couriers[c2.ID] = c2

	courierUC := usecase.NewCourierUsecase(courierRepo)

	free, err := courierUC.ListFreeCouriers(ctx)
	if err != nil {
		t.Fatalf("failed to list free couriers: %v", err)
	}
	if len(free) == 0 {
		t.Fatal("no free couriers available")
	}
	selected := free[0]

	assign := &modules.Assignment{
		ID:         "a1",
		OrderID:    "order123",
		CourierID:  selected.ID,
		AssignedAt: time.Now(),
	}
	if err := assignRepo.Create(ctx, assign); err != nil {
		t.Fatalf("failed to create assignment: %v", err)
	}

	if _, err := courierUC.UpdateStatus(ctx, selected.ID, usecase.UpdateStatusRequest{Status: modules.StatusBusy}); err != nil {
		t.Fatalf("failed to update courier status: %v", err)
	}

	a, err := assignRepo.GetByOrderID(ctx, "order123")
	if err != nil {
		t.Fatalf("failed to get assignment: %v", err)
	}
	if a.CourierID != selected.ID {
		t.Errorf("expected courier %s, got %s", selected.ID, a.CourierID)
	}
	if c1.Status != modules.StatusBusy {
		t.Errorf("expected courier status busy, got %s", c1.Status)
	}
}
