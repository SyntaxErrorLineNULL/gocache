package cache

import (
	"container/list"
	"context"
	"golang.org/x/sync/singleflight"
	"sync"
	"time"
)

// DefaultTTL is the default time-to-live duration for cache items.
const DefaultTTL = 1 * time.Hour

// MemoryCache represents an in-memory cache with TTL support.
type MemoryCache[K comparable, V any] struct {
	// parentCtx is the parent context used to manage the cache's goroutines.
	parentCtx context.Context
	// List to maintain the order of items.
	list *list.List
	// Map for quick lookups.
	items map[K]*list.Element
	// to search for expired data, we use the map. To reduce the impact on the data warehouse
	expirationBuckets map[K]*list.Element
	// wg is a WaitGroup for managing goroutines.
	wg sync.WaitGroup
	// ttl represents the default time-to-live duration for cache items.
	ttl time.Duration
	// Mutex for thread safety.
	mutex sync.RWMutex
	// New field for singleflight.Group
	group singleflight.Group
}

// NewMemoryCache creates a new MemoryCache instance.
// It initializes the cache with the provided context and starts a background
// goroutine that periodically deletes expired items from the cache.
func NewMemoryCache[K comparable, V any](ctx context.Context, ttl time.Duration) *MemoryCache[K, V] {
	// If no TTL is provided (ttl == 0), use the default TTL.
	if ttl == 0 {
		ttl = DefaultTTL
	}

	// Initialize a new MemoryCache instance with an empty list and map.
	return &MemoryCache[K, V]{
		parentCtx:         ctx,
		list:              list.New(),
		items:             make(map[K]*list.Element),
		expirationBuckets: make(map[K]*list.Element),
		ttl:               ttl,
	}
}
