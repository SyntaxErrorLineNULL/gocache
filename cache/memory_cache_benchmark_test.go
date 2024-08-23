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
