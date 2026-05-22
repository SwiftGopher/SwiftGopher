package repository

import (
	"context"
	"swift-gopher/internal/repository/_postgres"
	postgresAssignment "swift-gopher/internal/repository/_postgres/assignment"
	postgresAuth "swift-gopher/internal/repository/_postgres/auth"
	postgresCourier "swift-gopher/internal/repository/_postgres/courier"
	postgresOrders "swift-gopher/internal/repository/_postgres/order"
	postgresUser "swift-gopher/internal/repository/_postgres/user"
	"swift-gopher/pkg/modules"
	"time"
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
	ListWithFilter(ctx context.Context, filter modules.OrderFilter) ([]*modules.Order, error)
	UpdateStatus(ctx context.Context, id string, status modules.OrderStatus) error
	RecordHistory(ctx context.Context, history *modules.OrderHistory) error
	GetHistory(ctx context.Context, orderID string) ([]*modules.OrderHistory, error)
	GetMyOrders(ctx context.Context, userId string) ([]*modules.Order, error)
	GetByIDs(ctx context.Context, ids []string) ([]*modules.Order, error)
	GetByCourierID(ctx context.Context, courierID string) ([]*modules.Order, error)
}

type CourierRepository interface {
	Create(ctx context.Context, c *modules.Courier) error
	GetByID(ctx context.Context, id string) (*modules.Courier, error)
	GetByUserID(ctx context.Context, userID string) (*modules.Courier, error)

	List(ctx context.Context) ([]*modules.Courier, error)
	ListFree(ctx context.Context) ([]*modules.Courier, error)

	UpdateStatus(ctx context.Context, id string, status modules.CourierStatus) error
	UpdateLocation(ctx context.Context, id string, lat, lng float64) error
	UpdateTransport(ctx context.Context, id string, transport modules.TransportType) error
}

type AssignmentRepository interface {
	Create(ctx context.Context, a *modules.Assignment) error
	GetByOrderID(ctx context.Context, orderID string) (*modules.Assignment, error)
	Complete(ctx context.Context, orderID string, completedAt time.Time) error
	GetByCourierID(ctx context.Context, courierID string) ([]*modules.Assignment, error)
}

type UserRepository interface {
	GetMyProfile(ctx context.Context, userId string) (*modules.User, error)
}

type Repositories struct {
	AuthRepository
	OrderRepository
	CourierRepository
	AssignmentRepository
	UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		AuthRepository:       postgresAuth.NewAuthRepository(db),
		OrderRepository:      postgresOrders.NewOrderRepository(db),
		CourierRepository:    postgresCourier.NewCourierRepository(db),
		AssignmentRepository: postgresAssignment.NewAssignmentRepository(db),
		UserRepository:       postgresUser.NewUserRepo(db),
	}
}
