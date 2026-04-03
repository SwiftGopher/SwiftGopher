package handler

import (
	"errors"
	"net/http"

	"swift-gopher/internal/middleware"
	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateOrder(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req modules.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	order, err := h.usecases.OrderUsecase.CreateOrder(c.Request.Context(), claims.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrMissingAddress):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrInvalidPrice):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			h.log.Error("CreateOrder failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, order)
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
	var filter modules.OrderFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	orders, err := h.usecases.OrderUsecase.ListOrdersFiltered(c.Request.Context(), filter)
	if err != nil {
		h.log.Error("ListOrders failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   orders,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
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
		case errors.Is(err, usecase.ErrInvalidOrderStatus):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			h.log.Error("UpdateOrderStatus failed", "order_id", id, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) GetOrderHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing order id"})
		return
	}

	history, err := h.usecases.OrderUsecase.GetOrderHistory(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		h.log.Error("GetOrderHistory failed", "order_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, history)
}
