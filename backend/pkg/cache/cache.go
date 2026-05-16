package cache

import "github.com/redis/go-redis/v9"

func NewCache() *redis.Client {
	cache := redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	return cache
}
