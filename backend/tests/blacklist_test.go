package tests

import (
	"context"
	"testing"
	"time"
)

type mockBlacklist struct {
	data map[string]time.Time
}

func newMockBlacklist() *mockBlacklist {
	return &mockBlacklist{data: make(map[string]time.Time)}
}

func (b *mockBlacklist) Revoke(_ context.Context, token string, ttl time.Duration) error {
	b.data[token] = time.Now().Add(ttl)
	return nil
}

func (b *mockBlacklist) IsRevoked(_ context.Context, token string) (bool, error) {
	expiry, exists := b.data[token]
	if !exists {
		return false, nil
	}
	if time.Now().After(expiry) {
		delete(b.data, token)
		return false, nil
	}
	return true, nil
}

func TestBlacklist_NotRevokedByDefault(t *testing.T) {
	bl := newMockBlacklist()
	revoked, err := bl.IsRevoked(context.Background(), "some.jwt.token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revoked {
		t.Error("fresh token should not be revoked")
	}
}

func TestBlacklist_RevokedAfterLogout(t *testing.T) {
	bl := newMockBlacklist()
	token := "eyJhbGciOiJIUzI1NiJ9.payload.sig"

	if err := bl.Revoke(context.Background(), token, 1*time.Hour); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}

	revoked, err := bl.IsRevoked(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !revoked {
		t.Error("token should be revoked after logout")
	}
}

func TestBlacklist_ExpiredTokenNotRevoked(t *testing.T) {
	bl := newMockBlacklist()
	token := "expiring.token"

	bl.Revoke(context.Background(), token, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	revoked, err := bl.IsRevoked(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revoked {
		t.Error("token should not be revoked after TTL expiry")
	}
}

func TestBlacklist_DifferentTokensIndependent(t *testing.T) {
	bl := newMockBlacklist()

	bl.Revoke(context.Background(), "token-A", 1*time.Hour)

	revokedA, _ := bl.IsRevoked(context.Background(), "token-A")
	revokedB, _ := bl.IsRevoked(context.Background(), "token-B")

	if !revokedA {
		t.Error("token-A should be revoked")
	}
	if revokedB {
		t.Error("token-B should NOT be revoked")
	}
}
