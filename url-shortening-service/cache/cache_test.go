package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInterface(t *testing.T) {
	// test InMemoryCache implements }
	var cache CacheInterface = NewInMemoryCache(1 * time.Hour)

	// set a value
	cache.Set("qwerty", "https://example.com")

	// get should return value set
	longURL, found := cache.Get("qwerty")
	assert.True(t, found)
	assert.Equal(t, "https://example.com", longURL)

	// delete it
	cache.Delete("qwerty")

	// make sure no longer exists
	_, found = cache.Get("qwerty")
	assert.False(t, found)
}

func TestGet(t *testing.T) {
	// dict of items
	// check value is in list
	// check not in list
	// check time expired

	tests := []struct {
		name          string
		setupCache    func() *InMemoryCache
		key           string
		expectedValue string
		expectedOK    bool
	}{
		{
			name: "returns cache item",
			setupCache: func() *InMemoryCache {
				cacheTTL := 1 * time.Hour
				cache := NewInMemoryCache(cacheTTL)
				cache.Set("qwerty", "example.com")
				return cache
			},
			key:           "qwerty",
			expectedValue: "example.com",
			expectedOK:    true,
		},
		{
			name: "returns false for non-existent item",
			setupCache: func() *InMemoryCache {
				cacheTTL := 1 * time.Hour
				cache := NewInMemoryCache(cacheTTL)
				cache.Set("qwerty", "example.com")
				return cache
			},
			key:           "notfound",
			expectedValue: "",
			expectedOK:    false,
		},
		{name: "returns false for expired item",

			setupCache: func() *InMemoryCache {
				cache := &InMemoryCache{
					items: make(map[string]*CacheItem),
					ttl:   1 * time.Hour,
				}
				cache.items["expired"] = &CacheItem{
					LongURL:   "https://example.com",
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}
				return cache
			},
			key:           "expired",
			expectedValue: "",
			expectedOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.setupCache()

			longURL, found := cache.Get(tt.key)

			assert.Equal(t, tt.expectedValue, longURL)
			assert.Equal(t, tt.expectedOK, found)
		})
	}
}

func TestCleanup(t *testing.T) {
	// add mix of expired and valid items
	// create cache
	// run cleanup
	// check results

	cache := &InMemoryCache{
		items: make(map[string]*CacheItem),
		ttl:   1 * time.Hour,
	}

	now := time.Now()
	cache.items["expired1"] = &CacheItem{
		LongURL:   "https://old1.com",
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	cache.items["valid"] = &CacheItem{
		LongURL:   "https://valid.com",
		ExpiresAt: now.Add(1 * time.Hour),
	}
	cache.items["expired2"] = &CacheItem{
		LongURL:   "https://old2.com",
		ExpiresAt: now.Add(-2 * time.Hour),
	}

	cache.cleanup()

	assert.Equal(t, 1, cache.Size())
	_, found := cache.Get("valid")
	assert.True(t, found)
	_, found = cache.Get("expired1")
	assert.False(t, found)
	_, found = cache.Get("expired2")
	assert.False(t, found)
}
