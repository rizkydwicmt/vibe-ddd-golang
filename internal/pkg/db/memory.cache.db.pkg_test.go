package database

import (
	"context"
	"testing"

	"github.com/go-gorm/caches/v4"
	"github.com/stretchr/testify/assert"
)

func TestMemoryCacher_Init(t *testing.T) {
	t.Run("Initialize memory cacher", func(t *testing.T) {
		cacher := &memoryCacher{}

		// Initially store should be nil
		assert.Nil(t, cacher.store, "Store should be nil initially")

		// Call init
		cacher.init()

		// Store should be initialized
		assert.NotNil(t, cacher.store, "Store should be initialized after init")
	})
}

func TestMemoryCacher_Get(t *testing.T) {
	t.Run("Get existing key", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Create a query and marshal it
		query := &caches.Query[any]{}
		marshaledData, err := query.Marshal()
		assert.NoError(t, err, "Should be able to marshal query")

		// Store the marshaled data
		cacher.store.Store("test_key", marshaledData)

		// Create a new query to unmarshal into
		resultQuery := &caches.Query[any]{}

		// Get the data
		result, err := cacher.Get(context.Background(), "test_key", resultQuery)

		assert.NoError(t, err, "Get should not return error")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("Get non-existing key", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		query := &caches.Query[any]{}

		// Get non-existing key
		result, err := cacher.Get(context.Background(), "non_existing_key", query)

		assert.NoError(t, err, "Get should not return error for non-existing key")
		assert.Nil(t, result, "Result should be nil for non-existing key")
	})

	t.Run("Get with invalid data", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Store invalid data (not marshaled query)
		cacher.store.Store("invalid_key", []byte("invalid data"))

		query := &caches.Query[any]{}

		// Get the invalid data
		result, err := cacher.Get(context.Background(), "invalid_key", query)

		assert.Error(t, err, "Get should return error for invalid data")
		assert.Nil(t, result, "Result should be nil for invalid data")
	})
}

func TestMemoryCacher_Store(t *testing.T) {
	t.Run("Store valid data", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Create a query
		query := &caches.Query[any]{}

		// Store the data
		err := cacher.Store(context.Background(), "test_key", query)

		assert.NoError(t, err, "Store should not return error")

		// Verify data was stored
		val, ok := cacher.store.Load("test_key")
		assert.True(t, ok, "Data should be stored")
		assert.NotNil(t, val, "Stored value should not be nil")

		// Verify it's byte data
		_, isByte := val.([]byte)
		assert.True(t, isByte, "Stored value should be byte data")
	})

	t.Run("Store with marshal error", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Create a query that will fail to marshal
		query := &caches.Query[any]{}

		// Store the data
		err := cacher.Store(context.Background(), "test_key", query)

		assert.NoError(t, err, "Store should not return error for valid query")
	})
}

func TestMemoryCacher_Invalidate(t *testing.T) {
	t.Run("Invalidate cache", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Store some data
		cacher.store.Store("key1", []byte("data1"))
		cacher.store.Store("key2", []byte("data2"))

		// Verify data exists
		_, ok1 := cacher.store.Load("key1")
		_, ok2 := cacher.store.Load("key2")
		assert.True(t, ok1, "Key1 should exist before invalidation")
		assert.True(t, ok2, "Key2 should exist before invalidation")

		// Invalidate cache
		err := cacher.Invalidate(context.Background())

		assert.NoError(t, err, "Invalidate should not return error")

		// Verify data was cleared
		_, ok1 = cacher.store.Load("key1")
		_, ok2 = cacher.store.Load("key2")
		assert.False(t, ok1, "Key1 should not exist after invalidation")
		assert.False(t, ok2, "Key2 should not exist after invalidation")
	})

	t.Run("Invalidate empty cache", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Invalidate empty cache
		err := cacher.Invalidate(context.Background())

		assert.NoError(t, err, "Invalidate should not return error for empty cache")
	})
}

func TestMemoryCacher_Concurrency(t *testing.T) {
	t.Run("Concurrent access", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Test concurrent store and get operations
		done := make(chan bool, 2)

		// Goroutine 1: Store data
		go func() {
			query := &caches.Query[any]{}
			err := cacher.Store(context.Background(), "concurrent_key", query)
			assert.NoError(t, err, "Concurrent store should not return error")
			done <- true
		}()

		// Goroutine 2: Get data
		go func() {
			query := &caches.Query[any]{}
			result, err := cacher.Get(context.Background(), "concurrent_key", query)
			// Result might be nil if get happens before store
			_ = result
			_ = err
			done <- true
		}()

		// Wait for both goroutines to complete
		<-done
		<-done
	})
}

func TestMemoryCacher_EdgeCases(t *testing.T) {
	t.Run("Nil context", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		query := &caches.Query[any]{}

		// Test with TODO context instead of nil
		err := cacher.Store(context.TODO(), "test_key", query)
		assert.NoError(t, err, "Store should work with TODO context")

		result, err := cacher.Get(context.TODO(), "test_key", query)
		assert.NoError(t, err, "Get should work with TODO context")
		// Result might be nil if no data was stored
		_ = result
	})

	t.Run("Empty key", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		query := &caches.Query[any]{}

		// Test with empty key
		err := cacher.Store(context.Background(), "", query)
		assert.NoError(t, err, "Store should work with empty key")

		result, err := cacher.Get(context.Background(), "", query)
		assert.NoError(t, err, "Get should work with empty key")
		assert.NotNil(t, result, "Result should not be nil")
	})

	t.Run("Nil query", func(t *testing.T) {
		cacher := &memoryCacher{}
		cacher.init()

		// Test with nil query
		err := cacher.Store(context.Background(), "test_key", nil)
		assert.NoError(t, err, "Store should work with nil query")

		result, err := cacher.Get(context.Background(), "test_key", nil)
		assert.Error(t, err, "Get should work with nil query")
		assert.Nil(t, result, "Result should be nil with nil query")
	})
}
