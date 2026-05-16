package usecase

import (
	"context"
	"swift-gopher/internal/repository"
	"swift-gopher/pkg/modules"

	"github.com/redis/go-redis/v9"
)

type userUsecase struct {
	repo  repository.UserRepository
	cache *redis.Client
}

func NewUserUsecase(r repository.UserRepository, cache *redis.Client) *userUsecase {
	return &userUsecase{
		repo:  r,
		cache: cache,
	}
}

func (u *userUsecase) GetMyProfile(ctx context.Context, userId string) (*modules.User, error) {
	return u.repo.GetMyProfile(ctx, userId)
}
