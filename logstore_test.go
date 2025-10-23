package main

import (
	"testing"
)

func TestLogStore(t *testing.T) {
	t.Run("Initialization", func(t *testing.T) {
		t.Run("EmptyStore", func(t *testing.T) {
			// Arrange
			store := createTestLogStore(10)
			
			// Act & Assert
			assertStoreLength(t, store, 0)
			
			logs := store.Slice()
			assertSliceLength(t, logs, 0, "empty store")
		})
		
		t.Run("WithCapacity", func(t *testing.T) {
			// Arrange & Act
			capacity := 100
			store := createTestLogStore(capacity)
			
			// Assert
			assertIntEqual(t, store.capacity, capacity, "store capacity")
			assertStoreLength(t, store, 0)
		})
	})
	
	t.Run("AppendOperations", func(t *testing.T) {
		t.Run("WithinCapacity", func(t *testing.T) {
			// Arrange
			store := createTestLogStore(10)
			
			// Act
			for i := 0; i < 5; i++ {
				wrapped := store.Append(createTestLogEntry("test"))
				assertWrapResult(t, wrapped, false, "append within capacity")
			}
			
			// Assert
			assertStoreLength(t, store, 5)
		})
		
		t.Run("BeyondCapacity", func(t *testing.T) {
			// Arrange
			store := createTestLogStore(3)
			
			// Fill to capacity
			for i := 0; i < 3; i++ {
				wrapped := store.Append(createTestLogEntry("msg"))
				assertWrapResult(t, wrapped, false, "fill to capacity")
			}
			
			// Act - trigger wrap
			wrapped := store.Append(createTestLogEntry("overflow"))
			
			// Assert
			assertWrapResult(t, wrapped, true, "append beyond capacity")
			assertStoreLength(t, store, 3)
		})
		
		t.Run("AtExactCapacity", func(t *testing.T) {
			// Arrange
			capacity := 100
			store := createTestLogStore(capacity)
			
			// Act - add exactly capacity entries
			for i := 0; i < capacity; i++ {
				wrapped := store.Append(createTestLogEntry("test"))
				assertWrapResult(t, wrapped, false, "fill to exact capacity")
			}
			
			// Assert
			assertStoreLength(t, store, capacity)
			
			// Act - add one more to trigger wrap
			wrapped := store.Append(createTestLogEntry("overflow"))
			
			// Assert
			assertWrapResult(t, wrapped, true, "exceed exact capacity")
			assertStoreLength(t, store, capacity)
		})
	})
	
	t.Run("Ordering", func(t *testing.T) {
		t.Run("BeforeWrap", func(t *testing.T) {
			// Arrange
			store := createTestLogStore(5)
			expected := []string{"A", "B", "C"}
			
			// Act
			for _, msg := range expected {
				store.Append(createTestLogEntry(msg))
			}
			
			// Assert
			logs := store.Slice()
			assertSliceLength(t, logs, len(expected), "slice before wrap")
			
			for i, want := range expected {
				assertStringEqual(t, logs[i].Message, want)
			}
		})
		
		t.Run("AfterWrap", func(t *testing.T) {
			// Arrange
			store := createTestLogStore(3)
			
			// Fill buffer: A, B, C
			for _, msg := range []string{"A", "B", "C"} {
				store.Append(createTestLogEntry(msg))
			}
			
			// Act - wrap with D, E
			store.Append(createTestLogEntry("D"))
			store.Append(createTestLogEntry("E"))
			
			// Assert - after wrapping twice: oldest is C, then D, then E
			logs := store.Slice()
			expected := []string{"C", "D", "E"}
			
			assertSliceLength(t, logs, len(expected), "slice after wrap")
			
			for i, want := range expected {
				assertStringEqual(t, logs[i].Message, want)
			}
		})
		
		t.Run("MultipleWraps", func(t *testing.T) {
			// Arrange
			store := createTestLogStore(2)
			
			// Act - add many entries to test multiple wraps
			entries := []string{"A", "B", "C", "D", "E", "F"}
			for _, msg := range entries {
				store.Append(createTestLogEntry(msg))
			}
			
			// Assert - should only have last 2 entries
			logs := store.Slice()
			expected := []string{"E", "F"}
			
			assertSliceLength(t, logs, len(expected), "slice after multiple wraps")
			
			for i, want := range expected {
				assertStringEqual(t, logs[i].Message, want)
			}
		})
	})
}

// Benchmark tests for performance validation
func BenchmarkLogStore_Append(b *testing.B) {
	store := createTestLogStore(5000)
	entry := createTestLogEntry("benchmark test message")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Append(entry)
	}
}

func BenchmarkLogStore_Slice(b *testing.B) {
	store := createTestLogStore(5000)
	
	// Fill store
	for i := 0; i < 5000; i++ {
		store.Append(createTestLogEntry("test"))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Slice()
	}
}

func BenchmarkLogStore_AppendWithWrap(b *testing.B) {
	store := createTestLogStore(1000)
	entry := createTestLogEntry("benchmark")
	
	// Pre-fill to trigger wrapping
	for i := 0; i < 1500; i++ {
		store.Append(createTestLogEntry("prefill"))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Append(entry)
	}
}
