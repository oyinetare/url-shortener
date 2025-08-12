package cache

import (
	"sync"
	"time"
)

// Compile-time check that InMemoryCache implements Cache interface
var _ CacheInterface = (*InMemoryCache)(nil)

// in-memory caching with cinfigurable ttl eviction strategy
// deletion happens async with automatic cleanup i.e. a background goroutine that removes expired items periodically
// thread-safe using sync.RWMutex for concurrent access
// also logs cache hits & misses for monitoring

// CacheItem represents a cached URL mapping
type CacheItem struct {
	LongURL   string
	ExpiresAt time.Time
}

// InMemoryCache provides a simple thread-safe in-memory cache
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]*CacheItem
	ttl   time.Duration
}

// NewInMemoryCache creates a new in-memory cache with the given TTL
func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	// default to 1 hour if no TTL provided
	if ttl <= 0 {
		ttl = time.Hour
	}

	cache := &InMemoryCache{
		items: make(map[string]*CacheItem),
		ttl:   ttl,
	}

	// start cleanup routine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a long URL from the cache by short code
func (c *InMemoryCache) Get(shortCode string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[shortCode]

	// doesnt exist
	if !exists {
		return "", false
	}

	// cache item expired
	if time.Now().After(item.ExpiresAt) {
		return "", false
	}

	return item.LongURL, true
}

// Set stores a URL mapping in the cache
func (c *InMemoryCache) Set(shortCode, longURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[shortCode] = &CacheItem{
		LongURL:   longURL,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes an item from the cache
func (c *InMemoryCache) Delete(shortCode string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, shortCode)
}

func (c *InMemoryCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}

func (c *InMemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		now := time.Now()
		if now.After(item.ExpiresAt) {

			delete(c.items, key)
		}
	}
}

// cleanup periodically removes expired items
// Non-blocking main code: The cleanup runs in the background without blocking your API
// Regular intervals: Guaranteed to run at consistent intervals
// Automatic scheduling: No need to manually manage timing
// Goroutine-safe: Channels handle synchronization between goroutines
func (c *InMemoryCache) cleanupExpired() {
	// ticker.C channel that runs in a separate goroutine
	// and sends the current time to the channel C at regular intervals
	ticker := time.NewTicker(c.ttl / 2)
	// Running cleanup at half the TTL interval is a common pattern to ensure
	// expired items don't stay in memory too long
	// frequent enough to prevent significant memory buildup
	// but not as frequent to waste CPU cycles
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}

	// OLD IMPLEMENTATION - CLEANUP MOVED TO cleanup()
	// loop on channel ticker.C to receives values from it
	// for range ticker.C {
	// 	// c.mu.Lock()
	// 	// now := time.Now()
	// 	// for key, item := range c.items {
	// 	// 	if now.After(item.ExpiresAt) {
	// 	// 		delete(c.items, key)
	// 	// 	}
	// 	// }

	// 	// // since in a loop, cant use defer as all the unlocks to pile up until the function exits
	// 	// // (which is never, since it's an infinite loop)
	// 	// c.mu.Unlock()
	// }

}
