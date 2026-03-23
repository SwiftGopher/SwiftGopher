package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"swift-gopher/internal/handler"
	"swift-gopher/internal/repository"
	"swift-gopher/internal/repository/_postgres"
	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

func Run() {
	appCfg := initAppConfig()
	dbCfg := initPostgreConfig()

	logger := newLogger(appCfg.Env)

	ctx := context.Background()
	pg := _postgres.NewPGXDialect(ctx, dbCfg)
	defer pg.DB.Close()

	repositories := repository.NewRepositories(pg)

	usecases := usecase.NewUsecases(
		repositories,
		appCfg.JWTSecret,
		appCfg.AccessTokenTTL,
		appCfg.RefreshTokenTTL,
	)

	h := handler.NewHandler(usecases, logger)
	router := h.InitRoutes()

	srv := &http.Server{
		Addr:         ":" + appCfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", appCfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced shutdown", "error", err)
	}

	logger.Info("server exited gracefully")
}

func initAppConfig() *modules.AppConfig {
	accessTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		accessTTL = 15 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "168h"))
	if err != nil {
		refreshTTL = 168 * time.Hour
	}
	workerSec, err := time.ParseDuration(getEnv("WORKER_INTERVAL", "30s"))
	if err != nil {
		workerSec = 30 * time.Second
	}

	return &modules.AppConfig{
		Env:             getEnv("ENV", "development"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		JWTSecret:       getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
		WorkerInterval:  workerSec,
	}
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        getEnv("DB_HOST", "localhost"),
		Port:        getEnv("DB_PORT", "5432"),
		Username:    getEnv("DB_USERNAME", "postgres"),
		Password:    getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "swiftgopher"),
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newLogger(env string) *slog.Logger {
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
