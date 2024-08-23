package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
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
	cache := &MemoryCache[K, V]{
		parentCtx:         ctx,
		list:              list.New(),
		items:             make(map[K]*list.Element),
		expirationBuckets: make(map[K]*list.Element),
		ttl:               ttl,
	}

	// Increment the WaitGroup counter to track the collector goroutine.
	cache.wg.Add(1)

	// Start the collector goroutine to periodically delete expired items.
	go func() {
		// Decrement the WaitGroup counter when the goroutine exits.
		defer cache.wg.Done()
		cache.garbageCollection()
	}()

	return cache
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

// Contains checks if a given key exists in the cache.
// This method acquires a lock to ensure thread safety while accessing the cache
// and then checks if the key is present in the map of cached items.
func (m *MemoryCache[K, V]) Contains(key K) bool {
	// Acquire a read lock to ensure thread-safe access to the cache's data structures.
	m.mutex.RLock()
	// Ensure the read lock is released when the function returns.
	defer m.mutex.RUnlock()

	// Check if the key exists in the cache by looking it up in the items map.
	_, ok := m.items[key]
	// Return true if the key exists, false otherwise.
	return ok
}

// Remove removes an item from the cache by its key.
// This method is thread-safe and ensures that the item is properly removed from both
// the list and the map, as well as from the expiration buckets map.
// It returns true if the item was found and removed, and false if the key does not exist in the cache.
func (m *MemoryCache[K, V]) Remove(key K) bool {
	// Lock the mutex to ensure thread-safe access to the cache's data structures.
	m.mutex.Lock()
	// Ensure the mutex is unlocked when the function returns.
	defer m.mutex.Unlock()

	// Check if the key exists in the cache.
	if element, ok := m.items[key]; ok {
		// Remove the item from the list.
		m.list.Remove(element)
		// Remove the item from the items map.
		delete(m.items, key)
		// Remove the item from the expirationBuckets map.
		delete(m.expirationBuckets, key)
		// Return true indicating the item was successfully removed.
		return true
	}

	// Return false if the key does not exist in the cache.
	return false
}

// Len returns the number of items currently stored in the cache.
// This method is thread-safe and uses a read lock to ensure that the count
// is accurate without blocking other read operations.
func (m *MemoryCache[K, V]) Len() int {
	// Acquire a read lock to ensure thread-safe access to the cache's data structures.
	m.mutex.RLock()
	// Ensure the read lock is released when the function returns.
	defer m.mutex.RUnlock()

	// Return the length of the list, which represents the number of items in the cache.
	return m.list.Len()
}

// garbageCollection is a method that periodically deletes expired items from the cache.
// It runs in its own goroutine and uses a ticker to trigger the cleanup process
// at regular intervals defined by DefaultTTL.
func (m *MemoryCache[K, V]) garbageCollection() {
	// Create a ticker that ticks at intervals defined by DefaultTTL.
	ticker := time.NewTicker(m.ttl)
	// Stop the ticker when the method exits to release resources.
	defer ticker.Stop()

	for {
		select {
		// Trigger the deletion of expired data at each tick.
		case <-ticker.C:
			m.deleteExpiredData()

		// Exit the collector when the parent context is done (e.g., when the application is shutting down).
		case <-m.parentCtx.Done():
			return
		}
	}
}

// deleteExpiredData removes all expired items from the cache.
// It acquires write lock to ensure thread safety during the cleanup process.
func (m *MemoryCache[K, V]) deleteExpiredData() {
	// Lock the mutex to ensure thread-safe access to the cache's data structures.
	m.mutex.Lock()
	// Ensure the mutex is unlocked when the function returns.
	defer m.mutex.Unlock()

	// Iterate through the list from the back to the front.
	for element := m.list.Back(); element != nil; element = element.Prev() {
		// Retrieve the cache item from the list element.
		item := element.Value.(*Item[K, V])
		// Check if the item has expired.
		if item.ExpiresAt.Before(time.Now()) {
			// Remove the expired item from the list.
			m.list.Remove(element)
			// Remove the expired item from the items map.
			delete(m.items, item.Key)
			// Remove the expired item from the expirationBuckets map.
			delete(m.expirationBuckets, item.Key)
		} else {
			// Break the loop if we encounter a non-expired item
			break
		}
	}
}
