package handler

import (
	"net/http"

	"swift-gopher/internal/middleware"
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
	role := c.GetString("role")

	if role == "courier" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

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
	role := c.GetString("role")

	if role == "courier" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

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

func (h *Handler) GetMyCourier(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	courier, err := h.usecases.CourierUsecase.GetCourierByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		return
	}

	c.JSON(http.StatusOK, courier)
}

func (h *Handler) GetMyCourierOrders(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	courier, err := h.usecases.CourierUsecase.GetCourierByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		return
	}

	orders, err := h.usecases.OrderUsecase.GetCourierOrders(c.Request.Context(), courier.ID)
	if err != nil {
		h.log.Error("GetMyCourierOrders failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if orders == nil {
		orders = []*modules.Order{}
	}

	c.JSON(http.StatusOK, orders)
}
func (h *Handler) UpdateMyCourierStatus(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	courier, err := h.usecases.CourierUsecase.GetCourierByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		return
	}

	var req UpdateCourierStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	res, err := h.usecases.CourierUsecase.UpdateStatus(
		c.Request.Context(),
		courier.ID,
		usecase.UpdateStatusRequest{Status: req.Status},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, res)
}
func (h *Handler) UpdateMyCourierTransport(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	courier, err := h.usecases.CourierUsecase.GetCourierByUserID(
		c.Request.Context(),
		claims.UserID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		return
	}

	var req UpdateCourierTransportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	res, err := h.usecases.CourierUsecase.UpdateTransport(
		c.Request.Context(),
		courier.ID,
		usecase.UpdateTransportRequest{
			TransportType: req.TransportType,
		},
	)

	if err != nil {
		switch err {
		case usecase.ErrInvalidTransport:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transport type"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateMyCourierLocation(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	courier, err := h.usecases.CourierUsecase.GetCourierByUserID(
		c.Request.Context(),
		claims.UserID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		return
	}

	var req UpdateCourierLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	res, err := h.usecases.CourierUsecase.UpdateLocation(
		c.Request.Context(),
		courier.ID,
		usecase.UpdateLocationRequest(req),
	)

	if err != nil {
		switch err {
		case usecase.ErrInvalidLocation:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}
