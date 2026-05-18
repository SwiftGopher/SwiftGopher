package tokenblacklist

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBlacklist struct {
	client *redis.Client
}

func New(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{client: client}
}

func (b *TokenBlacklist) key(token string) string {
	return fmt.Sprintf("blacklist:%s", token)
}

func (b *TokenBlacklist) Revoke(ctx context.Context, token string, expiresIn time.Duration) error {
	if expiresIn <= 0 {
		return nil
	}
	return b.client.Set(ctx, b.key(token), "1", expiresIn).Err()
}

func (b *TokenBlacklist) IsRevoked(ctx context.Context, token string) (bool, error) {
	exists, err := b.client.Exists(ctx, b.key(token)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
