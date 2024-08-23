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
}
