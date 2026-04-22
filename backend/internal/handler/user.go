package handler

import (
	"net/http"
	"swift-gopher/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetMyProfile(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)

	profile, err := h.usecases.UserUsecase.GetMyProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, profile)
}
