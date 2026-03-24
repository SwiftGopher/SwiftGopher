package assignment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"swift-gopher/internal/repository/_postgres"
	"swift-gopher/pkg/modules"
)

var ErrAssignmentNotFound = errors.New("assignment not found")

type assignmentRepository struct {
	db *_postgres.Dialect
}

func NewAssignmentRepository(db *_postgres.Dialect) *assignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) Create(ctx context.Context, a *modules.Assignment) error {
	_, err := r.db.DB.Exec(ctx,
		`INSERT INTO assignments (id, order_id, courier_id, assigned_at)
		 VALUES ($1, $2, $3, $4)`,
		a.ID, a.OrderID, a.CourierID, a.AssignedAt,
	)
	if err != nil {
		return fmt.Errorf("assignmentRepo.Create: %w", err)
	}
	return nil
}

func (r *assignmentRepository) GetByOrderID(ctx context.Context, orderID string) (*modules.Assignment, error) {
	row := r.db.DB.QueryRow(ctx,
		`SELECT id, order_id, courier_id, assigned_at, completed_at FROM assignments WHERE order_id = $1`,
		orderID,
	)
	var a modules.Assignment
	err := row.Scan(&a.ID, &a.OrderID, &a.CourierID, &a.AssignedAt, &a.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssignmentNotFound
		}
		return nil, fmt.Errorf("assignmentRepo.GetByOrderID: %w", err)
	}
	return &a, nil
}

func (r *assignmentRepository) Complete(ctx context.Context, orderID string, completedAt time.Time) error {
	_, err := r.db.DB.Exec(ctx,
		`UPDATE assignments SET completed_at = $1 WHERE order_id = $2`,
		completedAt, orderID,
	)
	if err != nil {
		return fmt.Errorf("assignmentRepo.Complete: %w", err)
	}
	return nil
}
