package usecase

import (
	"context"
	"errors"
	"fmt"

	"swift-gopher/pkg/modules"
)

var (
	ErrCourierNotFound  = errors.New("courier not found")
	ErrInvalidStatus    = errors.New("invalid courier status")
	ErrInvalidTransport = errors.New("invalid courier transport type")
	ErrInvalidLocation  = errors.New("invalid courier location")
)

type UpdateStatusRequest struct {
	Status modules.CourierStatus
}

type courierUsecase struct {
	repo CourierRepository
}
type UpdateTransportRequest struct {
	TransportType modules.TransportType
}

type UpdateLocationRequest struct {
	Lat float64
	Lng float64
}

type CourierRepository interface {
	GetByID(ctx context.Context, id string) (*modules.Courier, error)
	List(ctx context.Context) ([]*modules.Courier, error)
	ListFree(ctx context.Context) ([]*modules.Courier, error)

	UpdateStatus(ctx context.Context, id string, status modules.CourierStatus) error
	UpdateTransport(ctx context.Context, id string, transport modules.TransportType) error
	UpdateLocation(ctx context.Context, id string, lat, lng float64) error
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

func (uc *courierUsecase) UpdateTransport(ctx context.Context, id string, req UpdateTransportRequest) (*modules.Courier, error) {
	if !isValidTransport(req.TransportType) {
		return nil, ErrInvalidTransport
	}

	c, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCourierNotFound
	}

	if err := uc.repo.UpdateTransport(ctx, id, req.TransportType); err != nil {
		return nil, fmt.Errorf("updating courier transport: %w", err)
	}

	c.TransportType = req.TransportType
	return c, nil
}

func isValidTransport(t modules.TransportType) bool {
	switch t {
	case modules.TransportBike,
		modules.TransportCar,
		modules.TransportFoot:
		return true
	}
	return false
}

func (uc *courierUsecase) UpdateLocation(ctx context.Context, id string, req UpdateLocationRequest) (*modules.Courier, error) {
	if req.Lat < -90 || req.Lat > 90 {
		return nil, ErrInvalidLocation
	}
	if req.Lng < -180 || req.Lng > 180 {
		return nil, ErrInvalidLocation
	}

	c, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCourierNotFound
	}

	if err := uc.repo.UpdateLocation(ctx, id, req.Lat, req.Lng); err != nil {
		return nil, fmt.Errorf("updating courier location: %w", err)
	}

	c.CurrentLat = req.Lat
	c.CurrentLng = req.Lng

	return c, nil
}
