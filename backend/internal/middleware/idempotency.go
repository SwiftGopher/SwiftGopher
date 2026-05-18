package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type idempotencyRecord struct {
	Response   []byte `json:"response"`
	StatusCode int    `json:"status_code"`
}

type bodyWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (bw *bodyWriter) Write(b []byte) (int, error) {
	bw.body.Write(b)
	return bw.ResponseWriter.Write(b)
}

func (bw *bodyWriter) WriteHeader(code int) {
	bw.statusCode = code
	bw.ResponseWriter.WriteHeader(code)
}

func Idempotency(cache *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		if cache == nil {
			c.Next()
			return
		}

		key := c.GetHeader("X-Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		claims := ClaimsFromContext(c)
		if claims == nil {
			c.Next()
			return
		}

		redisKey := fmt.Sprintf("idem:%s:%s", claims.UserID, key)
		ctx := context.Background()

		existing, err := cache.Get(ctx, redisKey).Bytes()
		if err == nil {
			var rec idempotencyRecord
			if jsonErr := json.Unmarshal(existing, &rec); jsonErr == nil {
				c.Data(rec.StatusCode, "application/json", rec.Response)
				c.Abort()
				return
			}
		}

		bw := &bodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}
		c.Writer = bw

		c.Next()

		if bw.statusCode >= 200 && bw.statusCode < 300 {
			rec := idempotencyRecord{
				Response:   bw.body.Bytes(),
				StatusCode: bw.statusCode,
			}
			data, _ := json.Marshal(rec)
			cache.SetNX(ctx, redisKey, data, 24*time.Hour)
		}
	}
}
