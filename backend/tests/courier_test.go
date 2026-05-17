package tests

import (
	"context"
	"testing"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

func newTestCourierUsecase() usecase.CourierUsecase {
	return usecase.NewCourierUsecase(newMockCourierRepo())
}

func TestGetCourier_Success(t *testing.T) {
	repo := newMockCourierRepo()
	c := &modules.Courier{ID: "c1", UserID: "u1", Status: modules.StatusFree}
	repo.couriers[c.ID] = c

	uc := usecase.NewCourierUsecase(repo)

	got, err := uc.GetCourier(context.Background(), "c1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.ID != "c1" {
		t.Errorf("expected ID c1, got %s", got.ID)
	}
}

func TestGetCourier_NotFound(t *testing.T) {
	uc := newTestCourierUsecase()

	_, err := uc.GetCourier(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListCouriers(t *testing.T) {
	repo := newMockCourierRepo()
	c1 := &modules.Courier{ID: "c1", Status: modules.StatusFree}
	c2 := &modules.Courier{ID: "c2", Status: modules.StatusBusy}
	repo.couriers[c1.ID] = c1
	repo.couriers[c2.ID] = c2

	uc := usecase.NewCourierUsecase(repo)

	list, err := uc.ListCouriers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 couriers, got %d", len(list))
	}
}

func TestListFreeCouriers(t *testing.T) {
	repo := newMockCourierRepo()
	c1 := &modules.Courier{ID: "c1", Status: modules.StatusFree}
	c2 := &modules.Courier{ID: "c2", Status: modules.StatusBusy}
	repo.couriers[c1.ID] = c1
	repo.couriers[c2.ID] = c2

	uc := usecase.NewCourierUsecase(repo)

	free, err := uc.ListFreeCouriers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(free) != 1 {
		t.Errorf("expected 1 free courier, got %d", len(free))
	}
	if free[0].ID != "c1" {
		t.Errorf("expected free courier c1, got %s", free[0].ID)
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	repo := newMockCourierRepo()
	c := &modules.Courier{ID: "c1", Status: modules.StatusFree}
	repo.couriers[c.ID] = c

	uc := usecase.NewCourierUsecase(repo)

	updated, err := uc.UpdateStatus(context.Background(), "c1", usecase.UpdateStatusRequest{Status: modules.StatusBusy})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != modules.StatusBusy {
		t.Errorf("expected status busy, got %s", updated.Status)
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	uc := newTestCourierUsecase()

	_, err := uc.UpdateStatus(context.Background(), "nonexistent", usecase.UpdateStatusRequest{Status: modules.StatusBusy})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
