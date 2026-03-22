package order

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type claimsGetter interface {
	UserID() string
}

type Handler struct {
	svc Service
	log *slog.Logger
}

func NewHandler(svc Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

type contextKey string

const claimsContextKey contextKey = "claims"

func getClaims(r *http.Request) (userID string, ok bool) {
	type hasClaims interface {
		GetUserID() string
	}
	v := r.Context().Value(claimsContextKey)
	if v == nil {
		return "", false
	}
	if c, ok2 := v.(hasClaims); ok2 {
		return c.GetUserID(), true
	}
	return "", false
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := getClaims(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingAddress):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrInvalidPrice):
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			h.log.Error("CreateOrder failed", "error", err)
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/orders/")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing order id")
		return
	}

	order, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			respondError(w, http.StatusNotFound, "order not found")
			return
		}
		h.log.Error("GetOrder failed", "order_id", id, "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	orders, err := h.svc.ListOrders(r.Context())
	if err != nil {
		h.log.Error("ListOrders failed", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, orders)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	id := strings.TrimSuffix(path, "/status")
	if id == "" || id == path {
		respondError(w, http.StatusBadRequest, "missing order id")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.svc.UpdateStatus(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			respondError(w, http.StatusNotFound, "order not found")
		case errors.Is(err, ErrInvalidStatus):
			respondError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			h.log.Error("UpdateStatus failed", "order_id", id, "error", err)
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func extractID(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(trimmed, "/"); idx != -1 {
		return trimmed[:idx]
	}
	return trimmed
}
