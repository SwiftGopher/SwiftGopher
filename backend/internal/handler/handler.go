package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"swift-gopher/internal/middleware"
	"swift-gopher/internal/usecase"
)

type Handler struct {
	usecases *usecase.Usecases
	log      *slog.Logger
}

func NewHandler(usecases *usecase.Usecases, log *slog.Logger) *Handler {
	return &Handler{usecases: usecases, log: log}
}

func (h *Handler) InitRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery(h.log))
	r.Use(middleware.Logger(h.log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
	}

	protected := r.Group("/")
	protected.Use(middleware.JWT(h.usecases.AuthUsecase))
	{
		// orders group → order developer's job
		// couriers group → courier developer's job
	}

	return r
}
