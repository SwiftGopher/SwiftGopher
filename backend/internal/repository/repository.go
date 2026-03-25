package repository

import (
	"context"

	"swift-gopher/internal/repository/_postgres"
	postgresAuth "swift-gopher/internal/repository/_postgres/auth"
	postgresOrders "swift-gopher/internal/repository/_postgres/orders"
	"swift-gopher/pkg/modules"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *modules.User) error
	GetUserByEmail(ctx context.Context, email string) (*modules.User, error)
	GetUserByID(ctx context.Context, id string) (*modules.User, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *modules.Order) error
	GetByID(ctx context.Context, id string) (*modules.Order, error)
	List(ctx context.Context) ([]*modules.Order, error)
	ListByStatus(ctx context.Context, status modules.OrderStatus) ([]*modules.Order, error)
	UpdateStatus(ctx context.Context, id string, status modules.OrderStatus) error
	RecordHistory(ctx context.Context, history *modules.OrderHistory) error
}

type Repositories struct {
	AuthRepository
	OrderRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		AuthRepository:  postgresAuth.NewAuthRepository(db),
		OrderRepository: postgresOrders.NewOrderRepository(db),
	}
}