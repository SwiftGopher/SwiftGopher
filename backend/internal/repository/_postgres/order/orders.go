package postgresOrders

import (
	"context"
	"errors"
	"fmt"

	"swift-gopher/internal/repository/_postgres"
	"swift-gopher/pkg/modules"

	"github.com/jackc/pgx/v5"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository struct {
	db *_postgres.Dialect
}

func NewOrderRepository(db *_postgres.Dialect) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *modules.Order) error {
	query := `
		INSERT INTO orders (id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.DB.Exec(ctx, query,
		o.ID, o.ClientID, o.PickupAddress, o.DeliveryAddress,
		string(o.Status), o.Price, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("orderRepository.Create: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*modules.Order, error) {
	query := `
		SELECT id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at
		FROM orders WHERE id = $1`

	row := r.db.DB.QueryRow(ctx, query, id)
	o := &modules.Order{}
	err := row.Scan(
		&o.ID, &o.ClientID, &o.PickupAddress, &o.DeliveryAddress,
		&o.Status, &o.Price, &o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetByID: %w", err)
	}
	return o, nil
}

func (r *OrderRepository) List(ctx context.Context) ([]*modules.Order, error) {
	query := `
		SELECT id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at
		FROM orders ORDER BY created_at DESC`

	rows, err := r.db.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.List: %w", err)
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *OrderRepository) ListByStatus(ctx context.Context, status modules.OrderStatus) ([]*modules.Order, error) {
	query := `
		SELECT id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at
		FROM orders WHERE status = $1 ORDER BY created_at ASC`

	rows, err := r.db.DB.Query(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("orderRepository.ListByStatus: %w", err)
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status modules.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`

	ct, err := r.db.DB.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateStatus: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) RecordHistory(ctx context.Context, h *modules.OrderHistory) error {
	query := `
		INSERT INTO order_history (id, order_id, old_status, new_status, changed_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.DB.Exec(ctx, query,
		h.ID, h.OrderID, string(h.OldStatus), string(h.NewStatus), h.ChangedAt,
	)
	if err != nil {
		return fmt.Errorf("orderRepository.RecordHistory: %w", err)
	}
	return nil
}

func scanOrders(rows pgx.Rows) ([]*modules.Order, error) {
	var orders []*modules.Order
	for rows.Next() {
		o := &modules.Order{}
		if err := rows.Scan(
			&o.ID, &o.ClientID, &o.PickupAddress, &o.DeliveryAddress,
			&o.Status, &o.Price, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanOrders: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanOrders rows.Err: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) GetHistory(ctx context.Context, orderID string) ([]*modules.OrderHistory, error) {
	query := `
		SELECT id, order_id, old_status, new_status, changed_at
		FROM order_history
		WHERE order_id = $1
		ORDER BY changed_at ASC`

	rows, err := r.db.DB.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.GetHistory: %w", err)
	}
	defer rows.Close()

	var history []*modules.OrderHistory
	for rows.Next() {
		h := &modules.OrderHistory{}
		if err := rows.Scan(&h.ID, &h.OrderID, &h.OldStatus, &h.NewStatus, &h.ChangedAt); err != nil {
			return nil, fmt.Errorf("orderRepository.GetHistory scan: %w", err)
		}
		history = append(history, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("orderRepository.GetHistory rows.Err: %w", err)
	}
	return history, nil
}

func (r *OrderRepository) ListWithFilter(ctx context.Context, filter modules.OrderFilter) ([]*modules.Order, error) {
	query := `
		SELECT id, client_id, pickup_address, delivery_address, status, price, created_at, updated_at
		FROM orders
		WHERE ($1 = '' OR status = $1)
		ORDER BY ` + sanitizeSortBy(filter.SortBy) + ` ` + sanitizeSortDir(filter.SortDir) + `
		LIMIT $2 OFFSET $3`

	rows, err := r.db.DB.Query(ctx, query,
		string(filter.Status),
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("orderRepository.ListWithFilter: %w", err)
	}
	defer rows.Close()
	return scanOrders(rows)
}

func sanitizeSortBy(s string) string {
	switch s {
	case "price":
		return "price"
	case "updated_at":
		return "updated_at"
	default:
		return "created_at"
	}
}

func sanitizeSortDir(s string) string {
	if s == "asc" {
		return "ASC"
	}
	return "DESC"
}
