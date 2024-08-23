package cache

import (
	"container/list"
	"context"
	"fmt"
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

// Set adds or updates an item in the cache with the specified key, value, and TTL (time-to-live).
// If the TTL is zero, the default TTL is used. This method is thread-safe and ensures
// that the cache is updated atomically.
func (m *MemoryCache[K, V]) Set(key K, value V, ttl time.Duration) {
	// If TTL is zero, set it to the default TTL value.
	if ttl == 0 {
		ttl = DefaultTTL
	}

	// Create a new item with the given key and value, and set its expiration time.
	// The expiration time is the current time plus the TTL duration.
	item := &Item[K, V]{Key: key, Value: value, ExpiresAt: time.Now().Add(ttl)}

	// Acquire white lock to ensure thread safety while updating the cache.
	// This prevents other goroutines from modifying the cache while this operation is in progress.
	m.mutex.Lock()
	// Ensure that the lock is released when the function completes, even if an error occurs or an early return happens.
	defer m.mutex.Unlock()

	// Add the new item to the front of the linked list (`m.list`).
	// Adding to the front makes it the most recently used item.
	element := m.list.PushFront(item)

	// Map the key to its corresponding list element in the `items` map.
	// This allows for quick lookup of the item by its key.
	m.items[key] = element

	// Also map the key to its corresponding list element in the `expirationBuckets` map.
	// This helps in managing expirations by keeping track of items and their positions.
	m.expirationBuckets[key] = element

	return
}

// Get retrieves an item from the cache by its key.
// If the item is found and has not expired, it is returned along with a boolean true.
// If the item is not found or has expired, the zero value and boolean false are returned.
func (m *MemoryCache[K, V]) Get(key K) (V, bool) {
	// Initialize a variable to hold the result. This will be returned if the key is not found.
	var res V

	// Format the key into a string for use with singleflight.
	keyStr := fmt.Sprintf("%v", key)

	// Use singleflight to ensure that multiple concurrent fetches for the same key are only processed once.
	value, err, _ := m.group.Do(keyStr, func() (interface{}, error) {
		// Acquire a read lock to ensure thread safety while reading from the cache.
		m.mutex.RLock()
		// Retrieve the list element associated with the key.
		element := m.items[key]
		// Release the read lock.
		m.mutex.RUnlock()

		// Check if the item associated with the key exists.
		if element == nil {
			// If the item is not found, return nil and an error (nil error here).
			return nil, nil
		}

		// Retrieve the item from the element.
		item := element.Value.(*Item[K, V])

		// Check if the item has expired.
		if item.ExpiresAt.Before(time.Now()) {
			// Acquire write lock to safely remove the expired item from the cache.
			m.mutex.Lock()

			// Remove expired item from the list and the maps.
			m.list.Remove(element)
			// Delete the item from the items map.
			delete(m.items, key)
			// Delete the item from the expirationBuckets map.
			delete(m.expirationBuckets, key)

			// Release write lock.
			m.mutex.Unlock()

			// Return nil, nil indicating item expired.
			return nil, nil
		}

		// If the item is found and has not expired, move it to the front of the list (most recently used).
		m.mutex.Lock()
		// Move the accessed item to the front of the list (most recently used).
		m.list.MoveToFront(element)
		// Set the result to the value of the item.
		res = item.Value
		// Release write lock.
		m.mutex.Unlock()

		// Return the value associated with the key and true.
		return res, nil
	})

	if err != nil || value == nil {
		// Return the zero value and false if the key was not found or item expired.
		return res, false
	}

	// Assert and return the fetched value and true.
	return value.(V), true
}
