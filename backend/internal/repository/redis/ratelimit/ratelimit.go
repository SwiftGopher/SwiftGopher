package ratelimit

import (
	"context"
	"fmt"
	"time"

	"swift-gopher/pkg/redisclient"
)

type RedisRateLimiter struct {
	rdb    *redisclient.Client
	limit  int
	window time.Duration
}

func New(rdb *redisclient.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{rdb: rdb, limit: limit, window: window}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, ip string) (bool, error) {
	windowID := time.Now().Unix() / int64(r.window.Seconds())
	key := fmt.Sprintf("rl:%s:%d", ip, windowID)

	count, err := r.rdb.IncrBy(ctx, key, 1)
	if err != nil {
		return true, err
	}

	if count == 1 {
		if err := r.rdb.Expire(ctx, key, r.window*2); err != nil {
			r.rdb.Del(ctx, key)
		}
	}

	return count <= int64(r.limit), nil
}
