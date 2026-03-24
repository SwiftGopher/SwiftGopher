package usecase

import (
	"context"
	"errors"
	"fmt"

	"swift-gopher/pkg/modules"
)

var (
	ErrCourierNotFound = errors.New("courier not found")
	ErrInvalidStatus   = errors.New("invalid courier status")
)

type UpdateStatusRequest struct {
	Status modules.CourierStatus
}

type CourierUsecase interface {
	GetCourier(ctx context.Context, id string) (*modules.Courier, error)
	ListCouriers(ctx context.Context) ([]*modules.Courier, error)
	UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*modules.Courier, error)
	ListFreeCouriers(ctx context.Context) ([]*modules.Courier, error)
}

type courierUsecase struct {
	repo CourierRepository
}

type CourierRepository interface {
	GetByID(ctx context.Context, id string) (*modules.Courier, error)
	List(ctx context.Context) ([]*modules.Courier, error)
	ListFree(ctx context.Context) ([]*modules.Courier, error)
	UpdateStatus(ctx context.Context, id string, status modules.CourierStatus) error
}

func NewCourierUsecase(repo CourierRepository) CourierUsecase {
	return &courierUsecase{repo: repo}
}

func (uc *courierUsecase) GetCourier(ctx context.Context, id string) (*modules.Courier, error) {
	c, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCourierNotFound
	}
	return c, nil
}

func (uc *courierUsecase) ListCouriers(ctx context.Context) ([]*modules.Courier, error) {
	return uc.repo.List(ctx)
}

func (uc *courierUsecase) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*modules.Courier, error) {
	if !isValidStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	c, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCourierNotFound
	}

	if err := uc.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		return nil, fmt.Errorf("updating courier status: %w", err)
	}

	c.Status = req.Status
	return c, nil
}

func (uc *courierUsecase) ListFreeCouriers(ctx context.Context) ([]*modules.Courier, error) {
	return uc.repo.ListFree(ctx)
}

func isValidStatus(s modules.CourierStatus) bool {
	switch s {
	case modules.StatusFree, modules.StatusBusy, modules.StatusOffline:
		return true
	}
	return false
}
