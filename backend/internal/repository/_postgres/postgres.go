package _postgres

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"swift-gopher/pkg/modules"
)

type Dialect struct {
	DB *pgxpool.Pool
}

func NewPGXDialect(ctx context.Context, cfg *modules.PostgreConfig) *Dialect {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(fmt.Sprintf("pgxpool.New: %v", err))
	}
	if err = pool.Ping(ctx); err != nil {
		panic(fmt.Sprintf("pgxpool ping: %v", err))
	}

	AutoMigrate(cfg)
	return &Dialect{DB: pool}
}

func AutoMigrate(cfg *modules.PostgreConfig) {
	sourceURL := "file://database/migrations"
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		panic(fmt.Sprintf("migrate.New: %v", err))
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(fmt.Sprintf("migrate.Up: %v", err))
	}
}
