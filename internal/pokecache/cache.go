package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Cache stores URL response bodies with a time-based eviction policy.
type Cache struct {
	mu       sync.Mutex
	entries  map[string]cacheEntry
	interval time.Duration
}

// NewCache returns a cache that reaps entries older than interval on each tick.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
	}
	go c.reapLoop()
	return c
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-c.interval)
		for k, v := range c.entries {
			if v.createdAt.Before(cutoff) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}

// Add stores a copy of val under key.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       append([]byte(nil), val...),
	}
}

// Get returns a copy of the cached value and whether key was present.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), e.val...), true
}
