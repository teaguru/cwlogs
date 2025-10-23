package main

import (
	"testing"
	"time"
)

// Test empty store
func TestLogStore_Empty(t *testing.T) {
	store := NewLogStore(10)
	if store.Len() != 0 {
		t.Errorf("Expected length 0, got %d", store.Len())
	}
	
	logs := store.Slice()
	if len(logs) != 0 {
		t.Errorf("Expected empty slice, got %d entries", len(logs))
	}
}

// Test append without wraparound
func TestLogStore_AppendNoWrap(t *testing.T) {
	store := NewLogStore(10)
	
	for i := 0; i < 5; i++ {
		wrapped := store.Append(LogEntry{
			Timestamp: time.Now(),
			Message:   "test",
			Raw:       "test",
		})
		if wrapped {
			t.Errorf("Should not wrap on entry %d", i)
		}
	}
	
	if store.Len() != 5 {
		t.Errorf("Expected length 5, got %d", store.Len())
	}
}

// Test append with wraparound
func TestLogStore_AppendWrap(t *testing.T) {
	store := NewLogStore(3)
	
	// Fill to capacity
	for i := 0; i < 3; i++ {
		wrapped := store.Append(LogEntry{
			Timestamp: time.Now(),
			Message:   "msg",
			Raw:       "msg",
		})
		if wrapped {
			t.Errorf("Should not wrap before capacity, entry %d", i)
		}
	}
	
	// Next append should wrap
	wrapped := store.Append(LogEntry{
		Timestamp: time.Now(),
		Message:   "overflow",
		Raw:       "overflow",
	})
	if !wrapped {
		t.Error("Expected wraparound signal")
	}
	
	if store.Len() != 3 {
		t.Errorf("Length should stay at capacity, got %d", store.Len())
	}
}

// Test slice ordering before wraparound
func TestLogStore_SliceOrdering(t *testing.T) {
	store := NewLogStore(5)
	
	for i := 0; i < 3; i++ {
		store.Append(LogEntry{
			Message: string(rune('A' + i)),
			Raw:     string(rune('A' + i)),
		})
	}
	
	logs := store.Slice()
	if len(logs) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(logs))
	}
	
	for i := 0; i < 3; i++ {
		expected := string(rune('A' + i))
		if logs[i].Message != expected {
			t.Errorf("Entry %d: expected %s, got %s", i, expected, logs[i].Message)
		}
	}
}

// Test slice ordering after wraparound
func TestLogStore_SliceOrderingAfterWrap(t *testing.T) {
	store := NewLogStore(3)
	
	// Fill buffer: A, B, C
	for i := 0; i < 3; i++ {
		store.Append(LogEntry{
			Message: string(rune('A' + i)),
			Raw:     string(rune('A' + i)),
		})
	}
	
	// Wrap with D, E
	store.Append(LogEntry{Message: "D", Raw: "D"})
	store.Append(LogEntry{Message: "E", Raw: "E"})
	
	logs := store.Slice()
	if len(logs) != 3 {
		t.Fatalf("Expected 3 entries after wrap, got %d", len(logs))
	}
	
	// After wrapping twice: oldest is C, then D, then E
	expected := []string{"C", "D", "E"}
	for i, exp := range expected {
		if logs[i].Message != exp {
			t.Errorf("Entry %d: expected %s, got %s", i, exp, logs[i].Message)
		}
	}
}

// Test capacity boundary
func TestLogStore_CapacityBoundary(t *testing.T) {
	capacity := 100
	store := NewLogStore(capacity)
	
	// Add exactly capacity entries
	for i := 0; i < capacity; i++ {
		store.Append(LogEntry{Message: "test", Raw: "test"})
	}
	
	if store.Len() != capacity {
		t.Errorf("Expected length %d, got %d", capacity, store.Len())
	}
	
	// Add one more to trigger wrap
	wrapped := store.Append(LogEntry{Message: "overflow", Raw: "overflow"})
	if !wrapped {
		t.Error("Expected wrap signal at capacity+1")
	}
	
	if store.Len() != capacity {
		t.Errorf("Length should remain at capacity %d, got %d", capacity, store.Len())
	}
}
