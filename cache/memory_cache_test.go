package cache

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache(t *testing.T) {
	t.Parallel()

	defer runtime.GC()

	// The "InitCache" test case verifies the initialization of a MemoryCache instance.
	// This test ensures that the MemoryCache is correctly created with the given parameters,
	// particularly focusing on the context and TTL (time-to-live) configuration.
	t.Run("InitCache", func(t *testing.T) {
		// Create a new context with cancellation capability.
		// This context will be used for the cache and can be canceled when the test completes.
		ctx, cancel := context.WithCancel(context.Background())

		// Ensure the cancel function is called after the test finishes.
		// This step is important for cleanup and to avoid potential resource leaks.
		defer cancel()

		// Define the TTL (time-to-live) for cache entries.
		// TTL specifies how long each cache entry will be kept before it is expired.
		ttl := 1 * time.Second

		// Initialize a new instance of MemoryCache with the specified context and TTL.
		// This call creates a cache that will use the provided context and have entries
		// that expire after the TTL duration.
		cache := NewMemoryCache[string, string](ctx, ttl)

		// Assert that the cache instance is not nil.
		// This check confirms that the MemoryCache was successfully created.
		assert.NotNil(t, cache, "Expected cache to be initialized but it was nil")

		// Assert that the TTL of the cache matches the expected TTL.
		// This verifies that the cache was initialized with the correct expiration duration.
		assert.Equal(t, ttl, cache.ttl, "Expected cache TTL to match the provided TTL")
	})

	// Begin the test case titled "Set and Fetch" which aims to validate the functionality of
	// the MemoryCache instance by checking the process of setting and retrieving a cache entry.
	// This test ensures that the MemoryCache correctly stores a key-value pair with a specified TTL
	// and that the value can be accurately retrieved within the TTL period. The test also verifies
	// that the cache behaves as expected by confirming the presence and correctness of the stored value.
	t.Run("Set and Fetch", func(t *testing.T) {
		// Create a new instance of MemoryCache designed to handle string keys and integer values.
		// The cache is initialized with a TTL (time-to-live) of 1 second, meaning that stored items
		// will expire and be removed from the cache after 1 second.
		cache := NewMemoryCache[string, int](context.Background(), 1*time.Second)

		// Insert a key-value pair ("key1", 42) into the cache with a TTL of 1 second.
		// This action will store the integer value 42 under the key "key1" in the cache.
		cache.Set("key1", 42, time.Second)

		// Retrieve the value associated with the key "key1" from the cache.
		// The cache should return the value if it is still valid and within the TTL period.
		element, ok := cache.Get("key1")

		// Verify that the retrieved element is not nil, ensuring that the value was successfully
		// fetched from the cache. If the element were nil, it would indicate a problem with the
		// retrieval process or that the item may have expired.
		assert.NotNil(t, element, "Expected element to be non-nil")

		// Check that the retrieved value matches the expected value 42.
		// This ensures that the correct value was stored and retrieved from the cache.
		assert.Equal(t, 42, element, "Expected element to be 42")

		// Confirm that the key "key1" exists in the cache by asserting that the `ok` value is true.
		// This validation ensures that the cache properly handled the insertion and retrieval
		// of the key-value pair.
		assert.Equal(t, true, ok, "Expected key 'key1' to exist in cache")
	})
}
