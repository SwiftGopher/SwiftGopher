package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

type UpdateCourierStatusRequest struct {
	Status modules.CourierStatus `json:"status"`
}

func (h *Handler) ListCouriers(c *gin.Context) {
	couriers, err := h.usecases.CourierUsecase.ListCouriers(c.Request.Context())
	if err != nil {
		h.log.Error("list couriers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if couriers == nil {
		couriers = []*modules.Courier{}
	}
	c.JSON(http.StatusOK, couriers)
}

func (h *Handler) ListFreeCouriers(c *gin.Context) {
	couriers, err := h.usecases.CourierUsecase.ListFreeCouriers(c.Request.Context())
	if err != nil {
		h.log.Error("list free couriers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if couriers == nil {
		couriers = []*modules.Courier{}
	}
	c.JSON(http.StatusOK, couriers)
}

func (h *Handler) UpdateCourierStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCourierStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	courier, err := h.usecases.CourierUsecase.UpdateStatus(c.Request.Context(), id, usecase.UpdateStatusRequest{
		Status: req.Status,
	})
	if err != nil {
		switch err {
		case usecase.ErrCourierNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		case usecase.ErrInvalidStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		default:
			h.log.Error("update courier status", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, courier)
}
