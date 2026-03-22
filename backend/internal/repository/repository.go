package repository

import (
	"context"

	"swift-gopher/internal/repository/_postgres"
	postgresAuth "swift-gopher/internal/repository/_postgres/auth"
	"swift-gopher/pkg/modules"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *modules.User) error
	GetUserByEmail(ctx context.Context, email string) (*modules.User, error)
	GetUserByID(ctx context.Context, id string) (*modules.User, error)
}

type Repositories struct {
	AuthRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		AuthRepository: postgresAuth.NewAuthRepository(db),
	}
}
