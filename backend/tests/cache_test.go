package tests

import (
	"sync"
	"testing"
	"time"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]interface{}
	ttl  map[string]time.Time
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	c.ttl[key] = time.Now().Add(ttl)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if expiry, exists := c.ttl[key]; exists && time.Now().After(expiry) {
		// Expired, but we'll clean up on next access
		return nil, false
	}
	value, exists := c.data[key]
	return value, exists
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	delete(c.ttl, key)
}

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache()

	cache.Set("key1", "value1", time.Minute)

	value, exists := cache.Get("key1")
	if !exists {
		t.Error("expected key to exist")
	}
	if value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	cache := NewCache()

	_, exists := cache.Get("nonexistent")
	if exists {
		t.Error("expected key to not exist")
	}
}

func TestCache_Expiry(t *testing.T) {
	cache := NewCache()

	cache.Set("key1", "value1", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	_, exists := cache.Get("key1")
	if exists {
		t.Error("expected key to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache()

	cache.Set("key1", "value1", time.Minute)
	cache.Delete("key1")

	_, exists := cache.Get("key1")
	if exists {
		t.Error("expected key to be deleted")
	}
}
