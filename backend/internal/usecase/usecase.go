package usecase

import (
	"context"
	"log/slog"
	"swift-gopher/internal/repository"
	"swift-gopher/pkg/modules"
	"time"

	"github.com/redis/go-redis/v9"
)

type AuthUsecase interface {
	Register(req modules.RegisterRequest) (*modules.User, error)
	Login(req modules.LoginRequest) (*modules.TokenPair, error)
	Refresh(refreshToken string) (*modules.TokenPair, error)
	ValidateAccessToken(token string) (*modules.Claims, error)
}

type OrderUsecase interface {
	CreateOrder(ctx context.Context, clientID string, req modules.CreateOrderRequest) (*modules.Order, error)
	GetOrder(ctx context.Context, id string) (*modules.Order, error)
	ListOrders(ctx context.Context) ([]*modules.Order, error)
	ListOrdersFiltered(ctx context.Context, filter modules.OrderFilter) ([]*modules.Order, error)
	ListPendingOrders(ctx context.Context) ([]*modules.Order, error)
	UpdateStatus(ctx context.Context, id string, req modules.UpdateOrderStatusRequest) (*modules.Order, error)
	GetOrderHistory(ctx context.Context, orderID string) ([]*modules.OrderHistory, error)
	GetMyOrders(ctx context.Context, userId string) ([]*modules.Order, error)
	GetCourierOrders(ctx context.Context, courierID string) ([]*modules.Order, error)
}

type CourierUsecase interface {
	GetCourier(ctx context.Context, id string) (*modules.Courier, error)
	GetCourierByUserID(ctx context.Context, userID string) (*modules.Courier, error)

	ListCouriers(ctx context.Context) ([]*modules.Courier, error)
	ListFreeCouriers(ctx context.Context) ([]*modules.Courier, error)
	FindNearestCourier(ctx context.Context, lat, lng float64) (*modules.Courier, error)

	UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*modules.Courier, error)
	UpdateTransport(ctx context.Context, id string, req UpdateTransportRequest) (*modules.Courier, error)
	UpdateLocation(ctx context.Context, id string, req UpdateLocationRequest) (*modules.Courier, error)
}

type UserUsecase interface {
	GetMyProfile(ctx context.Context, userId string) (*modules.User, error)
}

type Usecases struct {
	AuthUsecase
	OrderUsecase
	CourierUsecase
	AssignmentRepo repository.AssignmentRepository
	UserUsecase
}

func NewUsecases(repos *repository.Repositories, jwtSecret string, accessTTL, refreshTTL time.Duration, cache *redis.Client) *Usecases {
	courierUC := NewCourierUsecase(repos.CourierRepository, cache)
	return &Usecases{
		AuthUsecase:    NewAuthUsecase(repos.AuthRepository, repos.CourierRepository, jwtSecret, accessTTL, refreshTTL),
		OrderUsecase:   NewOrderUsecase(repos.OrderRepository, repos.AssignmentRepository, courierUC, slog.Default(), cache),
		CourierUsecase: courierUC,
		AssignmentRepo: repos.AssignmentRepository,
		UserUsecase:    NewUserUsecase(repos.UserRepository, cache),
	}
}
