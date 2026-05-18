package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"swift-gopher/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORS_AllowAll(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CORS([]string{"*"}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected *, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_Preflight(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CORS([]string{"*"}))
	r.OPTIONS("/test", func(c *gin.Context) {})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS preflight, got %d", w.Code)
	}
}

func TestCORS_SpecificOrigin_Allowed(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CORS([]string{"http://localhost:4200"}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Origin")
	if got != "http://localhost:4200" {
		t.Errorf("expected http://localhost:4200, got %q", got)
	}
}

func TestCORS_SpecificOrigin_Blocked(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CORS([]string{"http://localhost:4200"}))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Origin")
	if got != "" {
		t.Errorf("evil.com should NOT get CORS header, got %q", got)
	}
}

func TestRateLimit_AllowsNormalRequests(t *testing.T) {
	rl := middleware.NewRateLimiter(10, 20)
	r := gin.New()
	r.Use(middleware.RateLimit(rl))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	rl := middleware.NewRateLimiter(2, 2)
	r := gin.New()
	r.Use(middleware.RateLimit(rl))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1111"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d should pass, got %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:1111"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("3rd request should be 429, got %d", w.Code)
	}
}

func TestRateLimit_DifferentIPs_Independent(t *testing.T) {
	rl := middleware.NewRateLimiter(1, 1)
	r := gin.New()
	r.Use(middleware.RateLimit(rl))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	ips := []string{"1.1.1.1:0", "2.2.2.2:0", "3.3.3.3:0"}
	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("first request from %s should pass, got %d", ip, w.Code)
		}
	}
}

func TestIdempotency_SkipsNonPOST(t *testing.T) {
	r := gin.New()
	r.Use(middleware.Idempotency(nil))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Idempotency-Key", "some-key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET with idempotency key should pass, got %d", w.Code)
	}
}

func TestIdempotency_NilRedis_Passthrough(t *testing.T) {
	r := gin.New()
	r.Use(middleware.Idempotency(nil))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "test-key-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with nil Redis, got %d", w.Code)
	}
}

func TestIdempotency_NoKey_Passthrough(t *testing.T) {
	r := gin.New()
	r.Use(middleware.Idempotency(nil))
	r.POST("/test", func(c *gin.Context) { c.Status(201) })

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 without idempotency key, got %d", w.Code)
	}
}
