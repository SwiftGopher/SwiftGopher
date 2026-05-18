package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"swift-gopher/internal/repository/redis/ratelimit"
	"swift-gopher/internal/repository/redis/tokenblacklist"
	"swift-gopher/pkg/modules"
)

const ClaimsKey = "claims"

type TokenValidator interface {
	ValidateAccessToken(token string) (*modules.Claims, error)
}

func JWT(validator TokenValidator, bl *tokenblacklist.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		token := parts[1]

		if bl != nil {
			revoked, err := bl.IsRevoked(c.Request.Context(), token)
			if err == nil && revoked {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
				return
			}
		}

		claims, err := validator.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("raw_token", token)
		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

func RedisRateLimit(limiter *ratelimit.RedisRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		allowed, err := limiter.Allow(c.Request.Context(), c.ClientIP())
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}

func RequireRole(roles ...modules.Role) gin.HandlerFunc {
	allowed := make(map[modules.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		claims := ClaimsFromContext(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if _, ok := allowed[claims.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) *modules.Claims {
	v, exists := c.Get(ClaimsKey)
	if !exists {
		return nil
	}
	claims, _ := v.(*modules.Claims)
	return claims
}
