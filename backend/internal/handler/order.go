package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
	"swift-gopher/internal/middleware"
)

func (h *Handler) CreateOrder(c *gin.Context) {
    claims := middleware.ClaimsFromContext(c)
    if claims == nil {
        h.log.Error("CreateOrder: no claims found in context")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    userID := claims.UserID

    var req modules.CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    order, err := h.usecases.OrderUsecase.CreateOrder(c.Request.Context(), userID, req)
    if err != nil {
        h.handleCreateOrderError(c, err)
        return
    }
    c.JSON(http.StatusCreated, order)
}

func (h *Handler) handleCreateOrderError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, usecase.ErrMissingAddress), 
         errors.Is(err, usecase.ErrInvalidPrice):
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    default:
        h.log.Error("CreateOrder failed", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}

func (h *Handler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing order id"})
		return
	}

	order, err := h.usecases.OrderUsecase.GetOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		h.log.Error("GetOrder failed", "order_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	orders, err := h.usecases.OrderUsecase.ListOrders(c.Request.Context())
	if err != nil {
		h.log.Error("ListOrders failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing order id"})
		return
	}

	var req modules.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	order, err := h.usecases.OrderUsecase.UpdateStatus(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		case errors.Is(err, usecase.ErrInvalidStatus):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			h.log.Error("UpdateOrderStatus failed", "order_id", id, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}