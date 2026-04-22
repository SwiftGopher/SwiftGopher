package usecase

import (
	"context"
	"swift-gopher/internal/repository"
	"swift-gopher/pkg/modules"
)

type userUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(r repository.UserRepository) *userUsecase {
	return &userUsecase{repo: r}
}

func (u *userUsecase) GetMyProfile(ctx context.Context, userId string) (*modules.User, error) {
	return u.repo.GetMyProfile(ctx, userId)
}
