package worker

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"

	"swift-gopher/internal/repository"
	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

type Dispatcher struct {
	orderUsecase   usecase.OrderUsecase
	courierUsecase usecase.CourierUsecase
	assignmentRepo repository.AssignmentRepository
	interval       time.Duration
	log            *slog.Logger
}

func NewDispatcher(
	orderUsecase usecase.OrderUsecase,
	courierUsecase usecase.CourierUsecase,
	assignmentRepo repository.AssignmentRepository,
	interval time.Duration,
	log *slog.Logger,
) *Dispatcher {
	return &Dispatcher{
		orderUsecase:   orderUsecase,
		courierUsecase: courierUsecase,
		assignmentRepo: assignmentRepo,
		interval:       interval,
		log:            log,
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	d.log.Info("dispatcher started", "interval", d.interval)
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.log.Info("dispatcher shutting down")
			return
		case <-ticker.C:
			d.dispatch(ctx)
		}
	}
}

func (d *Dispatcher) dispatch(parentCtx context.Context) {
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	pendingOrders, err := d.orderUsecase.ListPendingOrders(ctx)
	if err != nil {
		d.log.Error("dispatcher: list pending orders", "error", err)
		return
	}
	if len(pendingOrders) == 0 {
		d.log.Debug("dispatcher: no pending orders")
		return
	}

	freeCouriers, err := d.courierUsecase.ListFreeCouriers(ctx)
	if err != nil {
		d.log.Error("dispatcher: list free couriers", "error", err)
		return
	}
	if len(freeCouriers) == 0 {
		d.log.Debug("dispatcher: no free couriers")
		return
	}

	for _, o := range pendingOrders {
		if ctx.Err() != nil {
			d.log.Warn("dispatcher: context expired mid-loop")
			return
		}

		nearest := findNearestCourier(freeCouriers, 0, 0)
		if nearest == nil {
			break
		}

		if err := d.assign(ctx, o, nearest); err != nil {
			d.log.Error("dispatcher: assign", "order_id", o.ID, "courier_id", nearest.ID, "error", err)
			continue
		}

		freeCouriers = removeCourier(freeCouriers, nearest.ID)
		d.log.Info("dispatcher: assigned", "order_id", o.ID, "courier_id", nearest.ID)
	}
}

func (d *Dispatcher) assign(ctx context.Context, o *modules.Order, c *modules.Courier) error {
	a := &modules.Assignment{
		ID:         uuid.NewString(),
		OrderID:    o.ID,
		CourierID:  c.ID,
		AssignedAt: time.Now().UTC(),
	}

	if err := d.assignmentRepo.Create(ctx, a); err != nil {
		return fmt.Errorf("create assignment: %w", err)
	}

	if _, err := d.orderUsecase.UpdateStatus(ctx, o.ID, modules.UpdateOrderStatusRequest{Status: modules.OrderStatusAssigned}); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	if _, err := d.courierUsecase.UpdateStatus(ctx, c.ID, usecase.UpdateStatusRequest{Status: modules.StatusBusy}); err != nil {
		return fmt.Errorf("update courier status: %w", err)
	}

	return nil
}

func findNearestCourier(couriers []*modules.Courier, lat, lng float64) *modules.Courier {
	if len(couriers) == 0 {
		return nil
	}
	nearest := couriers[0]
	minDist := haversine(lat, lng, nearest.CurrentLat, nearest.CurrentLng)
	for _, c := range couriers[1:] {
		if d := haversine(lat, lng, c.CurrentLat, c.CurrentLng); d < minDist {
			minDist = d
			nearest = c
		}
	}
	return nearest
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func removeCourier(couriers []*modules.Courier, id string) []*modules.Courier {
	out := make([]*modules.Courier, 0, len(couriers))
	for _, c := range couriers {
		if c.ID != id {
			out = append(out, c)
		}
	}
	return out
}
