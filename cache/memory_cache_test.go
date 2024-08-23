package cache

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache(t *testing.T) {
	t.Parallel()

	defer runtime.GC()

	// InitCache test case verifies the initialization of a MemoryCache instance.
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

	// Set and Fetch validates the behavior of the MemoryCache when setting and retrieving a value
	// associated with a specific key. This test ensures that the cache correctly handles insertion
	// and retrieval of values, updating existing entries as expected. The test sets a key with a value
	// and verifies that the updated value is correctly retrieved, confirming that the cache operates
	// as intended when handling single entries and their TTLs.
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

	// ReInsertion tests the behavior of MemoryCache when setting a key multiple times with different values.
	// This test ensures that the cache correctly updates the value associated with a key when a new value
	// is set, and that the updated value is correctly retrieved. The test involves setting the same key twice
	// with different values and verifying that the cache reflects the latest value as expected.
	t.Run("ReInsertion", func(t *testing.T) {
		// Create a new MemoryCache instance with string keys and integer values.
		// The cache is initialized with a TTL (time-to-live) of 1 second, meaning that entries will expire
		// after this duration. The TTL parameter is not directly used in this test but is part of the cache setup.
		cache := NewMemoryCache[string, int](context.Background(), 1*time.Second)

		// Set a key "key1" with the value 42 and a TTL of 3 seconds.
		// This stores the value 42 under the key "key1" in the cache with the specified TTL.
		cache.Set("key1", 42, 3*time.Second)

		// Immediately set the same key "key1" with a new value 57 and the same TTL of 3 seconds.
		// This updates the value associated with "key1" to 57. The TTL remains unchanged.
		cache.Set("key1", 57, 3*time.Second)

		// Retrieve the value associated with the key "key1" from the cache.
		// This operation fetches the value currently stored under "key1", which should be the updated value 57.
		element, ok := cache.Get("key1")

		// Assert that the retrieved element is not nil.
		// This confirms that the key "key1" exists in the cache and that a value was successfully retrieved.
		assert.NotNil(t, element, "Expected element to be non-nil")

		// Assert that the retrieved value matches the expected value 57.
		// This ensures that the cache correctly updated the value when the key was set a second time.
		assert.Equal(t, 57, element, "Expected element to be 57")

		// Assert that the key "key1" exists in the cache by checking that `ok` is true.
		// This verifies that the retrieval operation was successful and that the key is present in the cache.
		assert.True(t, ok, "Expected key 'key1' to exist in cache")
	})

	// Contains tests the behavior of MemoryCache when checking if keys are present in the cache.
	// This test ensures that the cache correctly reports the presence of keys after they have been set,
	// and verifies that the values associated with those keys are accurate. The test involves setting
	// key-value pairs with and without TTL, then confirming their presence and correctness in the cache.
	t.Run("Contains", func(t *testing.T) {
		// Create a new context with a cancellation function for managing the lifecycle of the cache.
		// The context will be used to signal when the cache operations should be terminated or cleaned up.
		ctx, cancel := context.WithCancel(context.Background())
		// Ensure that the context cancellation function is called when the test completes.
		// This will release resources associated with the context and prevent potential leaks.
		defer cancel()

		// Initialize a new MemoryCache instance with string keys and string values.
		// The cache is created with a TTL (Time-To-Live) of 1 second, specifying how long items should remain in the cache before expiring.
		// This cache will use the context to manage its lifecycle and ensure it adheres to the TTL constraints.
		cache := NewMemoryCache[string, string](ctx, 1*time.Second)

		// Set a key "key1" with the value "value1" and no TTL (TTL of 0).
		// This stores the value "value1" under the key "key1" in the cache without any expiration time.
		cache.Set("key1", "value1", 0)
		// Assert that the cache contains the key "key1".
		// This verifies that the key was successfully added to the cache and is present.
		assert.True(t, cache.Contains("key1"), "Expected cache to contain key 'key1'")

		// Fetch the value associated with "key1" from the cache.
		// This retrieves the value stored under "key1" to confirm that it matches the expected value.
		val, ok := cache.Get("key1")
		// Assert that the key "key1" exists in the cache by checking that `ok` is true.
		// This ensures that the retrieval operation was successful and that the key is present.
		assert.True(t, ok, "Expected key 'key1' to exist in cache")
		// Assert that the value retrieved for "key1" is "value1".
		// This confirms that the correct value is associated with the key.
		assert.Equal(t, "value1", val, "Expected value for 'key1' to be 'value1'")

		// Set another key "key2" with the value "value2" and a TTL of 1 second.
		// This stores the value "value2" under the key "key2" in the cache with a specified expiration time.
		cache.Set("key2", "value2", time.Second)
		// Assert that the cache contains the key "key2".
		// This verifies that the key was successfully added to the cache and is present.
		assert.True(t, cache.Contains("key2"), "Expected cache to contain key 'key2'")

		// Fetch the value associated with "key2" from the cache.
		// This retrieves the value stored under "key2" to confirm that it matches the expected value.
		_, ok = cache.Get("key2")
		// Assert that the key "key2" exists in the cache by checking that `ok` is true.
		// This ensures that the retrieval operation was successful and that the key is present.
		assert.True(t, ok, "Expected key 'key2' to exist in cache")
	})

	// NotFound tests the behavior of the MemoryCache when attempting to retrieve a value
	// associated with a key that does not exist in the cache. This test is designed to validate
	// that the cache correctly handles requests for keys that have not been stored. By checking
	// that the cache returns an indication that the key is not present, we ensure that the
	// cache's `Get` method properly differentiates between existing and non-existing keys,
	// thereby confirming its correctness in managing cache entries and reporting their status.
	t.Run("NotFound", func(t *testing.T) {
		// Create a new MemoryCache instance with string keys and integer values.
		// The cache is initialized with a TTL (Time-To-Live) of 1 second, specifying that cache entries will expire after this duration.
		cache := NewMemoryCache[string, int](context.Background(), 1*time.Second)

		// Attempt to fetch a value associated with the key "nonexistent" from the cache.
		// Since the key "nonexistent" was not previously set, we expect the cache to return a zero value and a boolean indicating absence.
		_, ok := cache.Get("nonexistent")
		// Assert that the result indicates the key does not exist in the cache.
		// The `ok` variable should be `false` to confirm that the key "nonexistent" was not found in the cache.
		// This test verifies that the cache correctly handles and reports missing keys.
		assert.Equal(t, false, ok, "Expected key 'nonexistent' not to exist in cache")
	})

	// GetNil tests the behavior of the MemoryCache when attempting to retrieve a value
	// associated with a key that is `nil`. This scenario is designed to validate how the
	// cache handles requests for keys that are not properly initialized or are explicitly set as `nil`.
	// By passing a `nil` key to the cache's `Get` method, this test ensures that the cache
	// correctly identifies the absence of such a key and handles it appropriately.
	// We expect the cache to return a zero value and a boolean indicating that the key is not present.
	t.Run("GetNil", func(t *testing.T) {
		// Create a new MemoryCache instance with interface{} as the key type and integer as the value type.
		// The cache is initialized with a TTL (Time-To-Live) of 1 second, specifying that cache entries will expire after this duration.
		cache := NewMemoryCache[interface{}, int](context.Background(), 1*time.Second)

		// Attempt to fetch a value associated with a `nil` key from the cache.
		// Since `nil` is not a valid key for caching purposes, we expect the cache to return a zero value
		// and a boolean indicating that the key is not found.
		_, ok := cache.Get(nil)

		// Assert that the result indicates the key `nil` does not exist in the cache.
		// The `ok` variable should be `false` to confirm that the cache handles `nil` keys correctly
		// by indicating that such a key is not present in the cache.
		assert.Equal(t, false, ok, "Expected key 'nil' not to exist in cache")
	})

	// RemoveKeyFromCache verifies that the MemoryCache correctly removes a key-value pair.
	// It tests the behavior of the Remove method by setting a key-value pair in the cache,
	// removing it, and then checking that the key no longer exists in the cache.
	t.Run("Remove", func(t *testing.T) {
		// Create a new MemoryCache instance configured to store string keys and integer values.
		// The cache is initialized with a global Time-To-Live (TTL) of 1 second, meaning
		// all entries in the cache will expire if they are not accessed within this time frame.
		cache := NewMemoryCache[string, int](context.Background(), 1*time.Second)

		// Insert a key-value pair into the cache using the Set method.
		// Here, "key1" is the key, and 42 is the associated value.
		// The TTL for this specific key-value pair is set to 1 second,
		// after which the entry would automatically expire if not removed earlier.
		cache.Set("key1", 42, time.Second)

		// Attempt to remove the key "key1" from the cache by calling the Remove method.
		// The Remove method should return a boolean value indicating whether the removal was successful.
		// If the key exists in the cache, it should return true, signifying the key has been removed.
		ok := cache.Remove("key1")

		// Use an assertion to verify that the Remove method returned true.
		// This check ensures that the key "key1" was present in the cache and was successfully removed.
		assert.Equal(t, true, ok, "Expected Remove to return true, indicating the key 'key1' was successfully removed from the cache")

		// After removing the key, attempt to retrieve the value associated with "key1" from the cache using the Get method.
		// Since "key1" was just removed, the Get method should not find it and should return a false boolean value.
		_, ok = cache.Get("key1")

		// Use an assertion to verify that the Get method returned false.
		// This ensures that "key1" no longer exists in the cache, confirming that the removal was effective.
		assert.False(t, ok, "Expected key 'key1' to be removed from cache, but it was still found")
	})

	// RemoveAndContains tests the behavior of the MemoryCache when removing a key-value pair and then checking for its existence.
	// This test case first ensures that attempting to remove a non-existent key returns false, indicating that nothing was removed.
	// It then adds a key-value pair to the cache without setting a TTL (Time-To-Live), removes it, and verifies that the key no longer exists in the cache.
	// The purpose of this test is to confirm that the cache correctly handles removal operations and accurately reflects the state of its contents after modifications.
	t.Run("RemoveAndContains", func(t *testing.T) {
		// Set up a new context to control the lifecycle of the MemoryCache instance.
		// The cancel function allows the context to be terminated early if needed.
		ctx, cancel := context.WithCancel(context.Background())
		// Ensure the cancel function is called to clean up the context once the test is done.
		defer cancel()

		// Initialize a new MemoryCache instance that will store string keys and values.
		// The cache is created with a TTL (Time-To-Live) of 1 second, meaning entries will expire after this period unless specified otherwise.
		cache := NewMemoryCache[string, string](ctx, 1*time.Second)

		// Attempt to remove a key ("nonexistent") that doesn't exist in the cache.
		// Since the key is not present, the Remove method should return false, indicating that no item was removed.
		assert.False(t, cache.Remove("nonexistent"), "Expected Remove to return false for non-existent key")

		// Add a key-value pair ("key1", "value1") to the cache without specifying a TTL.
		// By not setting a TTL, this entry is set to last indefinitely, or until manually removed.
		cache.Set("key1", "value1", 0)

		// Remove the previously added key "key1" from the cache.
		// The Remove method should return true, indicating that the key existed and was successfully removed.
		assert.True(t, cache.Remove("key1"), "Expected Remove to return true for existing key")

		// After removing "key1", check if it still exists in the cache using the Contains method.
		// The Contains method should return false, confirming that "key1" has been completely removed.
		assert.False(t, cache.Contains("key1"), "Expected key 'key1' to be removed from cache")
	})

	// RemoveNotFoundRecord tests the MemoryCache's handling of removal attempts for keys that do not exist in the cache.
	// This test case ensures that the cache's Remove method behaves correctly when it is asked to remove a key that has never been added.
	// Specifically, it verifies that the method does not alter the cache's state and returns false, clearly indicating that the removal operation
	// had no effect because the key was not found. This is crucial for maintaining the integrity of the cache's contents and ensuring that
	// no unintended side effects occur when attempting to remove non-existent items.
	t.Run("RemoveNotFoundRecord", func(t *testing.T) {
		// Initialize a new MemoryCache instance with string keys and integer values.
		// The cache is set with a TTL (Time-To-Live) of 1 second, meaning any entries (if they existed) would expire after this period.
		cache := NewMemoryCache[string, int](context.Background(), 1*time.Second)

		// Attempt to remove a key "nonexistent" which hasn't been added to the cache.
		// Since the key doesn't exist, the Remove method should return false, signifying that nothing was removed.
		ok := cache.Remove("nonexistent")

		// Verify that the return value of the Remove method is false.
		// The assertion checks that the removal operation correctly identifies the key's absence.
		assert.Equal(t, false, ok, "Expected Remove to return false for non-existent key")
	})

	// ExpirationRecord verifies the correct behavior of MemoryCache when handling item expiration.
	// This test ensures that items stored in the cache are properly expired and removed from memory after their TTL (Time-To-Live) has elapsed.
	// Specifically, it checks that an item added with a short TTL is no longer retrievable after the TTL has passed, confirming that
	// the expiration mechanism of the cache is functioning as expected.
	t.Run("ExpirationRecord", func(t *testing.T) {
		// Initialize a new MemoryCache instance with string keys and integer values.
		// The cache is created with a global TTL of 150 milliseconds, meaning that items without a specific TTL will expire after this time.
		cache := NewMemoryCache[string, int](context.Background(), 150*time.Millisecond)

		// Add an item to the cache with the key "key1" and value 42, setting a custom TTL of 100 milliseconds.
		// This TTL is shorter than the cache's global TTL, ensuring the item will expire 100 milliseconds after being added.
		cache.Set("key1", 42, 100*time.Millisecond)

		// Pause execution for 300 milliseconds to allow sufficient time for the item's TTL to expire.
		// This delay ensures that the item should no longer exist in the cache when retrieval is attempted.
		<-time.After(300 * time.Millisecond)

		// Attempt to retrieve the item associated with the key "key1" from the cache.
		// Since the item's TTL has expired, the cache should return false, indicating the item is no longer present.
		_, ok := cache.Get("key1")

		// Validate that the item is not found in the cache by asserting that `ok` is false.
		// This confirms that the item was correctly expired and removed from the cache.
		assert.Equal(t, false, ok)
	})
}

func TestMemoryCacheConcurrency(t *testing.T) {
	// Mark the test to be run in parallel with other tests.
	t.Parallel()

	// Create a cancellable context to control the cache's lifecycle.
	ctx, cancel := context.WithCancel(context.Background())
	// Ensure the context is cancelled at the end of the test.
	defer cancel()

	// Initialize a new instance of MemoryCache with the specified context and TTL.
	// This call creates a cache that will use the provided context and have entries
	// that expire after the TTL duration.
	cache := NewMemoryCache[string, string](ctx, DefaultTTL)

	// Define the number of concurrent goroutines to test cache operations.
	// This sets the level of concurrency by specifying how many goroutines will be created.
	const numGoroutines = 10

	// Create a sync.WaitGroup to wait for all goroutines to finish their execution.
	// This is necessary to ensure that the main test function does not exit before all goroutines complete their work.
	var wg sync.WaitGroup

	// Wait for all goroutines to complete their execution.
	// This call blocks until all goroutines have finished, ensuring that all concurrent operations are completed before the test ends.
	defer wg.Wait()

	// Subtest for concurrent set and get operations.
	// TestConcurrentSetAndGet tests the concurrent setting and getting of cache values.
	// It ensures that the cache can handle multiple goroutines setting and getting values concurrently.
	t.Run("TestConcurrentSetAndGet", func(t *testing.T) {
		// Loop to start multiple goroutines for concurrent cache operations.
		// The loop iterates 'numGoroutines' times to create the specified number of concurrent tasks.
		for i := 0; i < numGoroutines; i++ {
			// Increment the wait group counter by 1 for each new goroutine.
			// This tells the wait group to wait for one more goroutine to complete.
			wg.Add(1)

			// Launch a new goroutine to perform cache operations concurrently.
			// Each goroutine will execute the following function to set and get cache values.
			go func(i int) {
				// Ensure that the wait group counter is decremented when the goroutine completes.
				// This guarantees that the wait group will track this goroutine's completion.
				defer wg.Done()

				// Set the generated key-value pair in the cache.
				// The TTL is set to 0, indicating that the item should not expire automatically.
				cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0)

				// Retrieve the value associated with the generated key from the cache.
				// This simulates a cache lookup operation to check if the set operation was successful.
				val, ok := cache.Get(fmt.Sprintf("key%d", i))
				// Assert that the 'ok' variable is true, indicating that the key was found in the cache.
				// If 'ok' is false, the test will fail and display an error message that includes the key index
				// to identify which key was expected to be present in the cache.
				assert.True(t, ok, fmt.Sprintf("Expected key key %d to exist in cache", i))
				// Assert that the value retrieved from the cache ('val') matches the expected value ('value').
				// This checks whether the value associated with the key is correct.
				// If the values do not match, the test will fail and the error message will display the key index
				// and the expected value, helping to identify discrepancies.
				assert.Equal(t, fmt.Sprintf("value%d", i), val, fmt.Sprintf("Expected value for key key %d to be value %d", i, i))
			}(i)
		}

		// Allow time for all goroutines to complete their operations.
		time.Sleep(2 * time.Second)
	})

	// Subtest for concurrent remove operations.
	// TestConcurrentRemove tests the concurrent removal of cache values.
	// It ensures that the cache can handle multiple goroutines removing values concurrently.
	t.Run("TestConcurrentRemove", func(t *testing.T) {
		// Loop to start multiple goroutines for concurrent operations.
		// This will run the specified number of goroutines to test concurrent cache operations.
		for i := 0; i < numGoroutines; i++ {

			// Increment the wait group counter for each goroutine.
			// This ensures that the main test function waits for all goroutines to complete before finishing.
			wg.Add(1)

			// Launch a new goroutine to perform cache operations concurrently.
			// Each goroutine will handle setting and removing a key-value pair in the cache.
			go func(i int) {
				// Ensure that the wait group counter is decremented when the goroutine completes.
				// This allows the main test function to wait until all concurrent operations are done.
				defer wg.Done()

				// Set a key-value pair in the cache with a key formatted as "keyX" and a value formatted as "valueX".
				// The TTL (Time-To-Live) is set to 0, meaning the entry should persist indefinitely unless explicitly removed.
				cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0)

				// Remove the key-value pair from the cache.
				// This operation should delete the cache entry associated with the key formatted as "keyX".
				cache.Remove(fmt.Sprintf("key%d", i))

				// Attempt to retrieve the value from the cache for the key that was just removed.
				// This will verify whether the removal operation was successful.
				_, ok := cache.Get(fmt.Sprintf("key%d", i))

				// Assert that the value was not found in the cache after removal.
				// The assertion ensures that the cache does not contain the removed key, confirming the removal was effective.
				assert.False(t, ok, fmt.Sprintf("Expected key key%d to be removed from cache", i))
			}(i)
		}

		// Allow time for all goroutines to complete their operations.
		time.Sleep(2 * time.Second)
	})
}
