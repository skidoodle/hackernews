package cache

import (
	"sync"
	"time"
)

type CacheItem[T any] struct {
	Value      T
	Expiration int64
}

type Cache[T any] struct {
	items    map[string]CacheItem[T]
	mu       sync.RWMutex
	duration time.Duration
}

func New[T any](duration time.Duration) *Cache[T] {
	return &Cache[T]{
		items:    make(map[string]CacheItem[T]),
		duration: duration,
	}
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = CacheItem[T]{
		Value:      value,
		Expiration: time.Now().Add(c.duration).UnixNano(),
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found || time.Now().UnixNano() > item.Expiration {
		var zero T
		return zero, false
	}

	return item.Value, true
}
