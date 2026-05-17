package tests

import (
	"context"
	"testing"
	"time"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

func TestDispatcherAssignsCourier(t *testing.T) {
	ctx := context.Background()

	courierRepo := newMockCourierRepo()
	assignRepo := newMockAssignmentRepo()

	c1 := &modules.Courier{ID: "c1", Status: modules.StatusFree}
	c2 := &modules.Courier{ID: "c2", Status: modules.StatusBusy}
	courierRepo.couriers[c1.ID] = c1
	courierRepo.couriers[c2.ID] = c2

	courierUC := usecase.NewCourierUsecase(courierRepo, nil)

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
