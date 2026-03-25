package usecase

import (
	"context"
	"swift-gopher/internal/repository"
	"log/slog"
	"swift-gopher/pkg/modules"
	"time"
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
    ListPendingOrders(ctx context.Context) ([]*modules.Order, error)
    UpdateStatus(ctx context.Context, id string, req modules.UpdateOrderStatusRequest) (*modules.Order, error)
}

type Usecases struct {
	AuthUsecase
	OrderUsecase
}

func NewUsecases(repos *repository.Repositories, jwtSecret string, accessTTL, refreshTTL time.Duration) *Usecases {
	return &Usecases{
		AuthUsecase: NewAuthUsecase(repos.AuthRepository, jwtSecret, accessTTL, refreshTTL),
		OrderUsecase: NewOrderUsecase(repos.OrderRepository, slog.Default()),
	}
}
