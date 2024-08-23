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
