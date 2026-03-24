package usecase

import (
	"swift-gopher/internal/repository"
	"swift-gopher/pkg/modules"
	"time"
)

type AuthUsecase interface {
	Register(req modules.RegisterRequest) (*modules.User, error)
	Login(req modules.LoginRequest) (*modules.TokenPair, error)
	Refresh(refreshToken string) (*modules.TokenPair, error)
	ValidateAccessToken(token string) (*modules.Claims, error)
}

type Usecases struct {
	AuthUsecase
	CourierUsecase
}

func NewUsecases(repos *repository.Repositories, jwtSecret string, accessTTL, refreshTTL time.Duration) *Usecases {
	return &Usecases{
		AuthUsecase:    NewAuthUsecase(repos.AuthRepository, jwtSecret, accessTTL, refreshTTL),
		CourierUsecase: NewCourierUsecase(repos.CourierRepository),
	}
}
