package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"swift-gopher/internal/middleware"
)

func (h *Handler) Logout(c *gin.Context) {
	rawToken, exists := c.Get("raw_token")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	token, ok := rawToken.(string)
	if !ok || token == "" {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if h.blacklist != nil {
		if err := h.blacklist.Revoke(c.Request.Context(), token, 2*time.Hour); err != nil {
			h.log.Warn("logout: failed to revoke token in Redis", "error", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
