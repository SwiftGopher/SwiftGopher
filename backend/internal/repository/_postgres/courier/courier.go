package courier

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"swift-gopher/internal/repository/_postgres"
	"swift-gopher/pkg/modules"
)

var ErrCourierNotFound = errors.New("courier not found")

type courierRepository struct {
	db *_postgres.Dialect
}

func NewCourierRepository(db *_postgres.Dialect) *courierRepository {
	return &courierRepository{db: db}
}

func (r *courierRepository) Create(ctx context.Context, c *modules.Courier) error {
	_, err := r.db.DB.Exec(ctx,
		`INSERT INTO couriers (id, user_id, transport_type, status, current_lat, current_lng)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		c.ID, c.UserID, c.TransportType, c.Status, c.CurrentLat, c.CurrentLng,
	)
	if err != nil {
		return fmt.Errorf("courierRepo.Create: %w", err)
	}
	return nil
}

func (r *courierRepository) GetByID(ctx context.Context, id string) (*modules.Courier, error) {
	row := r.db.DB.QueryRow(ctx,
		`SELECT id, user_id, transport_type, status, current_lat, current_lng FROM couriers WHERE id = $1`,
		id,
	)
	return scanCourier(row)
}

func (r *courierRepository) GetByUserID(ctx context.Context, userID string) (*modules.Courier, error) {
	row := r.db.DB.QueryRow(ctx,
		`SELECT id, user_id, transport_type, status, current_lat, current_lng FROM couriers WHERE user_id = $1`,
		userID,
	)
	return scanCourier(row)
}

func (r *courierRepository) List(ctx context.Context) ([]*modules.Courier, error) {
	rows, err := r.db.DB.Query(ctx,
		`SELECT id, user_id, transport_type, status, current_lat, current_lng FROM couriers ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("courierRepo.List: %w", err)
	}
	defer rows.Close()
	return collectCouriers(rows)
}

func (r *courierRepository) ListFree(ctx context.Context) ([]*modules.Courier, error) {
	rows, err := r.db.DB.Query(ctx,
		`SELECT id, user_id, transport_type, status, current_lat, current_lng FROM couriers WHERE status = 'free'`,
	)
	if err != nil {
		return nil, fmt.Errorf("courierRepo.ListFree: %w", err)
	}
	defer rows.Close()
	return collectCouriers(rows)
}

func (r *courierRepository) UpdateStatus(ctx context.Context, id string, status modules.CourierStatus) error {
	_, err := r.db.DB.Exec(ctx,
		`UPDATE couriers SET status = $1 WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("courierRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *courierRepository) UpdateLocation(ctx context.Context, id string, lat, lng float64) error {
	_, err := r.db.DB.Exec(ctx,
		`UPDATE couriers SET current_lat = $1, current_lng = $2 WHERE id = $3`,
		lat, lng, id,
	)
	if err != nil {
		return fmt.Errorf("courierRepo.UpdateLocation: %w", err)
	}
	return nil
}

func scanCourier(row pgx.Row) (*modules.Courier, error) {
	var c modules.Courier
	err := row.Scan(&c.ID, &c.UserID, &c.TransportType, &c.Status, &c.CurrentLat, &c.CurrentLng)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCourierNotFound
		}
		return nil, fmt.Errorf("scanCourier: %w", err)
	}
	return &c, nil
}

func collectCouriers(rows pgx.Rows) ([]*modules.Courier, error) {
	var couriers []*modules.Courier
	for rows.Next() {
		var c modules.Courier
		if err := rows.Scan(&c.ID, &c.UserID, &c.TransportType, &c.Status, &c.CurrentLat, &c.CurrentLng); err != nil {
			return nil, fmt.Errorf("collectCouriers: %w", err)
		}
		couriers = append(couriers, &c)
	}
	return couriers, rows.Err()
}
