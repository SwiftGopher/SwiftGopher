package handler

import (
	"net/http"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"

	"github.com/gin-gonic/gin"
)

type UpdateCourierStatusRequest struct {
	Status modules.CourierStatus `json:"status"`
}
type UpdateCourierTransportRequest struct {
	TransportType modules.TransportType `json:"transport_type"`
}
type UpdateCourierLocationRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
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

func (h *Handler) UpdateCourierTransport(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCourierTransportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	courier, err := h.usecases.CourierUsecase.UpdateTransport(
		c.Request.Context(),
		id,
		usecase.UpdateTransportRequest{
			TransportType: req.TransportType,
		},
	)

	if err != nil {
		switch err {
		case usecase.ErrCourierNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		case usecase.ErrInvalidTransport:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transport type"})
		default:
			h.log.Error("update courier transport", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, courier)
}

func (h *Handler) UpdateCourierLocation(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCourierLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	courier, err := h.usecases.CourierUsecase.UpdateLocation(
		c.Request.Context(),
		id,
		usecase.UpdateLocationRequest(req),
	)

	if err != nil {
		switch err {
		case usecase.ErrCourierNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		case usecase.ErrInvalidLocation:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location"})
		default:
			h.log.Error("update courier location", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, courier)
}
