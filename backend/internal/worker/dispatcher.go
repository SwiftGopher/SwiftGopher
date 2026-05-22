package worker

import (
	"context"
	"log/slog"
	"math"
	"swift-gopher/pkg/modules"
	"swift-gopher/pkg/retry"
	"time"

	"swift-gopher/internal/repository"
	"swift-gopher/internal/usecase"

	"github.com/google/uuid"
)

const (
	poolSize       = 5
	queueSize      = 50
	maxAttempts    = 3
	retryBaseDelay = 500 * time.Millisecond
)

type Dispatcher struct {
	orderUsecase   usecase.OrderUsecase
	courierUsecase usecase.CourierUsecase
	assignmentRepo repository.AssignmentRepository
	interval       time.Duration
	log            *slog.Logger
	pool           *Pool
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
		pool:           NewPool(poolSize, queueSize, log),
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	d.log.Info("dispatcher started",
		"interval", d.interval,
		"workers", poolSize,
	)

	d.pool.Start(ctx)
	defer d.pool.Shutdown(ctx)

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	d.dispatch(ctx)

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

func (d *Dispatcher) Dispatch(ctx context.Context) {
	d.pool = NewPool(poolSize, queueSize, d.log)
	d.pool.Start(ctx)
	d.dispatch(ctx)
	d.pool.Wait()
}

func (d *Dispatcher) dispatch(parentCtx context.Context) {
	fetchCtx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	orders, err := d.orderUsecase.ListPendingOrders(fetchCtx)
	if err != nil {
		d.log.Error("failed to list orders", "error", err)
		return
	}

	couriers, err := d.courierUsecase.ListFreeCouriers(fetchCtx)
	if err != nil {
		d.log.Error("failed to list couriers", "error", err)
		return
	}

	if len(orders) == 0 || len(couriers) == 0 {
		return
	}

	assigned := make(map[string]bool, len(couriers))

	for _, order := range orders {
		courier := nearestCourier(order, couriers, assigned)
		if courier == nil {
			break
		}
		assigned[courier.ID] = true

		o := order
		c := courier

		err := d.pool.Submit(func() {
			d.assignWithRetry(parentCtx, o, c)
		})

		if err != nil {
			d.log.Warn("submit failed",
				"order_id", order.ID,
				"error", err,
			)
		}
	}
}

func nearestCourier(order *modules.Order, couriers []*modules.Courier, assigned map[string]bool) *modules.Courier {
	var best *modules.Courier
	bestDist := math.MaxFloat64

	for _, c := range couriers {
		if assigned[c.ID] {
			continue
		}
		dist := haversine(order.PickupLat, order.PickupLng, c.CurrentLat, c.CurrentLng)
		if dist < bestDist {
			bestDist = dist
			best = c
		}
	}

	return best
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func (d *Dispatcher) assignWithRetry(
	ctx context.Context,
	order *modules.Order,
	courier *modules.Courier,
) {
	err := retry.Do(ctx, maxAttempts, retryBaseDelay, func() error {
		return d.assign(ctx, order, courier)
	})

	if err != nil {
		d.log.Error("assignment failed",
			"order_id", order.ID,
			"courier_id", courier.ID,
			"error", err,
		)
		return
	}

	d.log.Info("assigned",
		"order_id", order.ID,
		"courier_id", courier.ID,
	)
}

func (d *Dispatcher) assign(
	ctx context.Context,
	order *modules.Order,
	courier *modules.Courier,
) error {
	assignment := &modules.Assignment{
		ID:         uuid.NewString(),
		OrderID:    order.ID,
		CourierID:  courier.ID,
		AssignedAt: time.Now().UTC(),
	}

	if err := d.assignmentRepo.Create(ctx, assignment); err != nil {
		return err
	}

	_, err := d.orderUsecase.UpdateStatus(ctx, order.ID, modules.UpdateOrderStatusRequest{
		Status: modules.OrderStatusAssigned,
	})
	if err != nil {
		return err
	}

	_, err = d.courierUsecase.UpdateStatus(ctx, courier.ID, usecase.UpdateStatusRequest{
		Status: modules.StatusBusy,
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *Dispatcher) Pool() *Pool {
	return d.pool
}
