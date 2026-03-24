package repository

import (
	"context"
	"swift-gopher/internal/repository/_postgres"
	postgresAssignment "swift-gopher/internal/repository/_postgres/assignment"
	postgresAuth "swift-gopher/internal/repository/_postgres/auth"
	postgresCourier "swift-gopher/internal/repository/_postgres/courier"
	"swift-gopher/pkg/modules"
	"time"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *modules.User) error
	GetUserByEmail(ctx context.Context, email string) (*modules.User, error)
	GetUserByID(ctx context.Context, id string) (*modules.User, error)
}

type CourierRepository interface {
	Create(ctx context.Context, c *modules.Courier) error
	GetByID(ctx context.Context, id string) (*modules.Courier, error)
	GetByUserID(ctx context.Context, userID string) (*modules.Courier, error)
	List(ctx context.Context) ([]*modules.Courier, error)
	ListFree(ctx context.Context) ([]*modules.Courier, error)
	UpdateStatus(ctx context.Context, id string, status modules.CourierStatus) error
	UpdateLocation(ctx context.Context, id string, lat, lng float64) error
}

type AssignmentRepository interface {
	Create(ctx context.Context, a *modules.Assignment) error
	GetByOrderID(ctx context.Context, orderID string) (*modules.Assignment, error)
	Complete(ctx context.Context, orderID string, completedAt time.Time) error
}

type Repositories struct {
	AuthRepository
	CourierRepository
	AssignmentRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		AuthRepository:       postgresAuth.NewAuthRepository(db),
		CourierRepository:    postgresCourier.NewCourierRepository(db),
		AssignmentRepository: postgresAssignment.NewAssignmentRepository(db),
	}
}
