package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

func (h *Handler) Register(c *gin.Context) {
	var req modules.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = modules.RoleClient
	}

	user, err := h.usecases.AuthUsecase.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
		case errors.Is(err, usecase.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		default:
			h.log.Error("register", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) Login(c *gin.Context) {
	var req modules.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.usecases.AuthUsecase.Login(req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		h.log.Error("login", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req modules.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.usecases.AuthUsecase.Refresh(req.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		h.log.Error("refresh", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
