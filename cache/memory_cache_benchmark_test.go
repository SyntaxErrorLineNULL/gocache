package cache

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkMemoryCache_Set benchmarks the performance of the Set method in MemoryCache.
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
