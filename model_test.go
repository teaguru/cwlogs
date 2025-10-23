package main

import (
	"testing"
	"time"
)

// Test model initialization
func TestNewLogModel(t *testing.T) {
	config := NewUIConfig()
	model := newLogModel("test-group", config)
	
	if model.logGroup != "test-group" {
		t.Errorf("Expected log group 'test-group', got %q", model.logGroup)
	}
	
	if model.store == nil {
		t.Error("Expected store to be initialized")
	}
	
	if model.store.Len() != 0 {
		t.Errorf("Expected empty store, got %d entries", model.store.Len())
	}
	
	if !model.followMode {
		t.Error("Expected follow mode to be enabled by default")
	}
	
	if model.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", model.cursor)
	}
}

// Test cursor management
func TestCursorManagement(t *testing.T) {
	config := NewUIConfig()
	model := newLogModel("test", config)
	
	// Add some test logs
	for i := 0; i < 5; i++ {
		entry := logEntry{
			Timestamp:       time.Now(),
			Message:         "test message",
			OriginalMessage: "test message",
			Raw:            "test message",
		}
		model.store.Append(entry)
	}
	
	// Test cursor bounds
	model.cursor = 10 // Out of bounds
	model.fixCursor()
	
	if model.cursor >= model.store.Len() {
		t.Errorf("Cursor should be bounded, got %d for %d logs", model.cursor, model.store.Len())
	}
	
	// Test negative cursor
	model.cursor = -5
	model.fixCursor()
	
	if model.cursor < 0 {
		t.Errorf("Cursor should not be negative, got %d", model.cursor)
	}
}

// Test follow mode behavior
func TestFollowMode(t *testing.T) {
	config := NewUIConfig()
	model := newLogModel("test", config)
	
	// Add logs
	for i := 0; i < 3; i++ {
		entry := logEntry{
			Timestamp:       time.Now(),
			Message:         "test message",
			OriginalMessage: "test message",
			Raw:            "test message",
		}
		model.store.Append(entry)
	}
	
	// Enable follow mode
	model.followMode = true
	model.fixCursor()
	
	// Cursor should be at the end
	expectedCursor := model.store.Len() - 1
	if model.cursor != expectedCursor {
		t.Errorf("In follow mode, cursor should be at %d, got %d", expectedCursor, model.cursor)
	}
	
	// Disable follow mode
	model.followMode = false
	model.cursor = 0
	model.fixCursor()
	
	// Cursor should stay at 0
	if model.cursor != 0 {
		t.Errorf("With follow mode off, cursor should stay at 0, got %d", model.cursor)
	}
}

// Test search state management
func TestSearchState(t *testing.T) {
	config := NewUIConfig()
	model := newLogModel("test", config)
	
	// Set up search state
	model.searchQuery = "test query"
	model.matches = []int{0, 2, 4}
	model.currentMatch = 1
	model.highlighted = make(map[int]string)
	model.highlighted[0] = "highlighted"
	
	// Clear search state
	model.clearSearchState()
	
	// Note: clearSearchState keeps searchQuery but clears matches and highlights
	if len(model.matches) != 0 {
		t.Errorf("Expected empty matches, got %d", len(model.matches))
	}
	
	if model.currentMatch != 0 {
		t.Errorf("Expected currentMatch 0, got %d", model.currentMatch)
	}
	
	if len(model.highlighted) != 0 {
		t.Errorf("Expected empty highlights, got %d", len(model.highlighted))
	}
}

// Test safe logs access
func TestSafeLogs(t *testing.T) {
	config := NewUIConfig()
	model := newLogModel("test", config)
	
	// Empty store
	logs := model.safeLogs()
	if len(logs) != 0 {
		t.Errorf("Expected empty logs, got %d", len(logs))
	}
	
	// Add some logs
	for i := 0; i < 3; i++ {
		entry := logEntry{
			Timestamp:       time.Now(),
			Message:         "test message",
			OriginalMessage: "test message",
			Raw:            "test message",
		}
		model.store.Append(entry)
	}
	
	logs = model.safeLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

// Test back to log groups functionality
func TestBackToLogGroups(t *testing.T) {
	config := NewUIConfig()
	model := newLogModel("test", config)
	
	// Initially should not want to go back
	if model.backToLogGroups {
		t.Error("Expected backToLogGroups to be false initially")
	}
	
	// Simulate back key press by sending backToLogGroupsMsg
	msg := backToLogGroupsMsg{}
	updatedModel, cmd := model.Update(msg)
	
	// Should set the flag and return quit command
	if logModel, ok := updatedModel.(*logModel); ok {
		if !logModel.backToLogGroups {
			t.Error("Expected backToLogGroups to be true after back message")
		}
	} else {
		t.Error("Expected model to be *logModel type")
	}
	
	// Should return tea.Quit command
	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}
}



// Helper functions for model creation
func newLogModel(logGroup string, config *UIConfig) *logModel {
	return &logModel{
		logGroup:     logGroup,
		store:        newLogStore(config.MaxLogBuffer),
		followMode:   true,
		cursor:       0,
		currentMatch: -1,
		highlighted:  make(map[int]string),
		config:       config,
	}
}
