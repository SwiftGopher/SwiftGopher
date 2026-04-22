package postgresUser

import (
	"context"
	"swift-gopher/internal/repository/_postgres"
	"swift-gopher/pkg/modules"
)

type userRepo struct {
	db *_postgres.Dialect
}

func NewUserRepo(db *_postgres.Dialect) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) GetMyProfile(ctx context.Context, userId string) (*modules.User, error) {
	var user modules.User

	query := `select id, email, role, created_at from users where id = $1`
	err := r.db.DB.QueryRow(ctx, query, userId).Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
