package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, o *Order) error {
	query := `
		INSERT INTO orders (id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		o.ID, o.ClientID, o.PickupAddress, o.DeliveryAddress,
		string(o.Status), o.Price, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("order.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Order, error) {
	query := `
		SELECT id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at
		FROM orders
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)
	o := &Order{}
	err := row.Scan(
		&o.ID, &o.ClientID, &o.PickupAddress, &o.DeliveryAddress,
		&o.Status, &o.Price, &o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("order.repository.GetByID: %w", err)
	}
	return o, nil
}

func (r *repository) List(ctx context.Context) ([]*Order, error) {
	query := `
		SELECT id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("order.repository.List: %w", err)
	}
	defer rows.Close()

	return scanOrders(rows)
}

func (r *repository) ListByStatus(ctx context.Context, status Status) ([]*Order, error) {
	query := `
		SELECT id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("order.repository.ListByStatus: %w", err)
	}
	defer rows.Close()

	return scanOrders(rows)
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status Status) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2`

	ct, err := r.db.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("order.repository.UpdateStatus: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *repository) RecordHistory(ctx context.Context, h *OrderHistory) error {
	query := `
		INSERT INTO order_history (id, order_id, old_status, new_status, changed_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query,
		h.ID, h.OrderID, string(h.OldStatus), string(h.NewStatus), h.ChangedAt,
	)
	if err != nil {
		return fmt.Errorf("order.repository.RecordHistory: %w", err)
	}
	return nil
}

func scanOrders(rows pgx.Rows) ([]*Order, error) {
	var orders []*Order
	for rows.Next() {
		o := &Order{}
		if err := rows.Scan(
			&o.ID, &o.ClientID, &o.PickupAddress, &o.DeliveryAddress,
			&o.Status, &o.Price, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("order.repository.scanOrders: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("order.repository.scanOrders rows.Err: %w", err)
	}
	return orders, nil
}
