package auth

import (
	"context"
	"errors"
	"fmt"

	"swift-gopher/internal/repository/_postgres"
	"swift-gopher/pkg/modules"
)

var ErrUserNotFound = errors.New("user not found")

type authRepository struct {
	db *_postgres.Dialect
}

func NewAuthRepository(db *_postgres.Dialect) *authRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(ctx context.Context, u *modules.User) error {
	_, err := r.db.DB.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, role, created_at)
         VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Email, u.PasswordHash, u.Role, u.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("authRepo.CreateUser: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*modules.User, error) {
	row := r.db.DB.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`,
		email,
	)
	return scanUser(row)
}

func (r *authRepository) GetUserByID(ctx context.Context, id string) (*modules.User, error) {
	row := r.db.DB.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`,
		id,
	)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*modules.User, error) {
	var u modules.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scanUser: %w", err)
	}
	return &u, nil
}
