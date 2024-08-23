package cache

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkMemoryCacheSet benchmarks the performance of the Set method in MemoryCache.
func BenchmarkMemoryCacheSet(b *testing.B) {
	// Create a new MemoryCache instance with a TTL of 1 minute.
	cache := NewMemoryCache[string, int](context.Background(), 1*time.Minute)

	// Reset the benchmark timer to exclude setup time.
	b.ResetTimer()

	// Run the benchmark to measure the performance of the Set method.
	for i := 0; i < b.N; i++ {
		// Set an item in the cache with a unique key and a TTL of 1 minute.
		cache.Set(fmt.Sprintf("key%d", i), i, 1*time.Minute)
	}

	// Forcing the garbage collector (GC) to run to clear memory after benchmarking.
	runtime.GC()
}

// BenchmarkMemoryCacheGet benchmarks the performance of the Fetch (Get) method in MemoryCache.
func BenchmarkMemoryCacheGet(b *testing.B) {
	// Create a new MemoryCache instance with a TTL of 1 hour.
	cache := NewMemoryCache[string, int](context.Background(), 1*time.Hour)

	// Populate the cache with benchmark data using goroutines.
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		// Add an item to the cache in a goroutine to simulate concurrent access.
		go func(key string, value int, ttl time.Duration) {
			cache.Set(key, value, ttl)
			wg.Done()
		}(fmt.Sprintf("key%d", i), i, 10*time.Minute) // Each item has a TTL of 10 minutes.
	}

	// Wait for all goroutines to finish populating the cache.
	wg.Wait()

	// Reset the benchmark timer to exclude setup time.
	b.ResetTimer()

	// Benchmark the Fetch operation.
	for i := 0; i < b.N; i++ {
		// Retrieve an item from the cache with a unique key.
		cache.Get(fmt.Sprintf("key%d", i))
	}

	// Forcing the garbage collector (GC) to run to clear memory after benchmarking.
	runtime.GC()
}

// BenchmarkMemoryCacheDeleteExpiredData benchmark test to add a large amount of data to the cache and measure deleteExpiredData performance.
func BenchmarkMemoryCacheDeleteExpiredData(b *testing.B) {
	// Initialize a new MemoryCache instance with a TTL.
	cache := NewMemoryCache[int, int](context.Background(), DefaultTTL)

	// Number of items to add to the cache.
	numItems := 100000

	// Add items to the cache with varied expiration times.
	for i := 0; i < numItems; i++ {
		var expiration time.Duration

		// Assign a shorter TTL to some items (e.g., 100 milliseconds).
		if i%100 == 0 {
			expiration = 100 * time.Millisecond
		} else {
			// Assign random expiration times between 1 millisecond and 10 minutes to other items.
			expiration = time.Duration((i%600)+1) * time.Millisecond
		}

		cache.Set(i, i, expiration)
	}

	// Run the benchmark to measure the performance of deleteExpiredData for items with short TTL.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.deleteExpiredData()
	}

	// Forcing the GC to run to clear all the memory
	runtime.GC()
}
