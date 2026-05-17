package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"swift-gopher/internal/handler"
	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

func newTestHandler() *handler.Handler {
	authUC := newTestAuthUsecase()
	orderUC := newTestOrderUsecase()
	courierUC := newTestCourierUsecase()
	usecases := &usecase.Usecases{
		AuthUsecase:    authUC,
		OrderUsecase:   orderUC,
		CourierUsecase: courierUC,
	}
	return handler.NewHandler(usecases, newTestLogger())
}

func TestHealthCheck(t *testing.T) {
	h := newTestHandler()
	r := h.InitRoutes()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "ok", response["status"])
}

func TestHandlerRegister_Success(t *testing.T) {
	h := newTestHandler()
	r := h.InitRoutes()

	reqBody := modules.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     modules.RoleClient,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response modules.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test@example.com", response.Email)
}

func TestHandlerRegister_InvalidRole(t *testing.T) {
	h := newTestHandler()
	r := h.InitRoutes()

	reqBody := modules.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "invalid",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerLogin_Success(t *testing.T) {
	// First register a user
	h := newTestHandler()
	r := h.InitRoutes()

	regBody := modules.RegisterRequest{
		Email:    "login@example.com",
		Password: "password123",
		Role:     modules.RoleClient,
	}
	body, _ := json.Marshal(regBody)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Now login
	loginBody := modules.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}
	body, _ = json.Marshal(loginBody)

	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response modules.TokenPair
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
}

func TestCreateOrder_Unauthorized(t *testing.T) {
	h := newTestHandler()
	r := h.InitRoutes()

	reqBody := modules.CreateOrderRequest{
		PickupAddress:   "A",
		DeliveryAddress: "B",
		Price:           10.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
