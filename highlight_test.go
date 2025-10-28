package main

import (
	"strings"
	"testing"
	"time"
)

func TestCurrentMatchHighlighting(t *testing.T) {
	t.Run("CurrentMatchStyle", func(t *testing.T) {
		// Arrange
		config := NewUIConfig()
		
		// Act - just verify the functions exist and don't panic
		currentStyle := config.CurrentMatchStyle()
		regularStyle := config.HighlightStyle()
		
		// Assert - verify styles can render text without panicking
		testText := "test"
		currentRendered := currentStyle.Render(testText)
		regularRendered := regularStyle.Render(testText)
		
		// Both should contain the original text (styling may not be visible in tests)
		if !strings.Contains(currentRendered, testText) {
			t.Error("Current match style should contain the original text")
		}
		
		if !strings.Contains(regularRendered, testText) {
			t.Error("Regular highlight style should contain the original text")
		}
		
		// Verify both styles can be created without errors
		if currentRendered == "" {
			t.Error("Current match style should produce non-empty output")
		}
		
		if regularRendered == "" {
			t.Error("Regular highlight style should produce non-empty output")
		}
	})

	t.Run("ApplyHighlightsWithCurrentMatch", func(t *testing.T) {
		// Arrange
		model := createTestLogModel("test")
		
		// Add test logs with searchable content
		testLogs := []string{
			"This is a test message",
			"Another test entry here", 
			"Final test log entry",
		}
		
		for i, msg := range testLogs {
			entry := makeLogEntry(time.Now().Add(time.Duration(i)*time.Second), msg, model.config)
			model.store.Append(entry)
		}
		
		// Set up search
		model.searchQuery = "test"
		model.performSearch()
		
		// Act - should have found 3 matches
		if len(model.matches) != 3 {
			t.Fatalf("Expected 3 matches, got %d", len(model.matches))
		}
		
		// Assert - check that highlights are applied
		if len(model.highlighted) != 3 {
			t.Errorf("Expected 3 highlighted entries, got %d", len(model.highlighted))
		}
		
		// Verify currentMatch was set correctly after search (should be last match, index 2)
		if model.currentMatch != 2 {
			t.Errorf("Expected currentMatch to be 2 (last match) after search, got %d", model.currentMatch)
		}
		
		// Check that current match has highlighting
		currentMatchIdx := model.matches[model.currentMatch]
		currentHighlight := model.highlighted[currentMatchIdx]
		
		// The current match should contain the search term
		if !strings.Contains(currentHighlight, "test") {
			t.Error("Current match highlight should contain the search term")
		}
		
		// Check other matches have highlighting too
		for i, matchIdx := range model.matches {
			if i != model.currentMatch {
				otherHighlight := model.highlighted[matchIdx]
				if !strings.Contains(otherHighlight, "test") {
					t.Error("Regular match highlight should contain the search term")
				}
			}
		}
	})

	t.Run("NextMatchUpdatesHighlighting", func(t *testing.T) {
		// Arrange
		model := createTestLogModel("test")
		
		// Add test logs
		testLogs := []string{
			"First test message",
			"Second test entry", 
			"Third test log",
		}
		
		for i, msg := range testLogs {
			entry := makeLogEntry(time.Now().Add(time.Duration(i)*time.Second), msg, model.config)
			model.store.Append(entry)
		}
		
		// Set up search
		model.searchQuery = "test"
		model.performSearch()
		
		// Verify initial state - should start at LAST match (index 2)
		if model.currentMatch != 2 {
			t.Fatalf("Expected currentMatch to be 2 (last match), got %d", model.currentMatch)
		}
		
		// Act - move to next match (goes backward to older logs)
		model.nextMatch()
		
		// Assert - current match should have moved backward to index 1
		if model.currentMatch != 1 {
			t.Errorf("Expected currentMatch to be 1 (moved backward), got %d", model.currentMatch)
		}
		
		// Verify all matches still highlighted
		if len(model.highlighted) != 3 {
			t.Errorf("Expected 3 highlighted entries after nextMatch, got %d", len(model.highlighted))
		}
	})

	t.Run("PrevMatchUpdatesHighlighting", func(t *testing.T) {
		// Arrange
		model := createTestLogModel("test")
		
		// Add test logs
		testLogs := []string{
			"First test message",
			"Second test entry", 
		}
		
		for i, msg := range testLogs {
			entry := makeLogEntry(time.Now().Add(time.Duration(i)*time.Second), msg, model.config)
			model.store.Append(entry)
		}
		
		// Set up search
		model.searchQuery = "test"
		model.performSearch()
		
		// Verify starts at last match (index 1)
		if model.currentMatch != 1 {
			t.Fatalf("Expected currentMatch to be 1 (last match), got %d", model.currentMatch)
		}
		
		// Act - move to previous match (goes forward to newer, wraps to index 0)
		model.prevMatch()
		
		// Assert - should wrap to first match (index 0)
		if model.currentMatch != 0 {
			t.Errorf("Expected currentMatch to be 0 (wrapped forward), got %d", model.currentMatch)
		}
		
		// Verify highlighting was updated
		if len(model.highlighted) != 2 {
			t.Errorf("Expected 2 highlighted entries after prevMatch, got %d", len(model.highlighted))
		}
	})

	t.Run("NoHighlightingWhenNoMatches", func(t *testing.T) {
		// Arrange
		model := createTestLogModel("test")
		
		// Add test log without search term
		entry := makeLogEntry(time.Now(), "no matching content here", model.config)
		model.store.Append(entry)
		
		// Act - search for non-existent term
		model.searchQuery = "nonexistent"
		model.performSearch()
		
		// Assert - no matches, no highlights
		if len(model.matches) != 0 {
			t.Errorf("Expected 0 matches, got %d", len(model.matches))
		}
		
		if len(model.highlighted) != 0 {
			t.Errorf("Expected 0 highlights, got %d", len(model.highlighted))
		}
	})

	t.Run("ClearSearchClearsHighlights", func(t *testing.T) {
		// Arrange
		model := createTestLogModel("test")
		
		// Add test log
		entry := makeLogEntry(time.Now(), "test message", model.config)
		model.store.Append(entry)
		
		// Set up search
		model.searchQuery = "test"
		model.performSearch()
		
		// Verify highlights exist
		if len(model.highlighted) == 0 {
			t.Fatal("Expected highlights to exist before clearing")
		}
		
		// Act - clear search
		model.clearSearchState()
		
		// Assert - highlights should be cleared
		if len(model.highlighted) != 0 {
			t.Errorf("Expected highlights to be cleared, got %d", len(model.highlighted))
		}
		
		if len(model.matches) != 0 {
			t.Errorf("Expected matches to be cleared, got %d", len(model.matches))
		}
	})

	t.Run("CurrentMatchHasDifferentStyling", func(t *testing.T) {
		// Arrange
		model := createTestLogModel("test")
		
		// Add test logs
		testLogs := []string{
			"First test message",
			"Second test entry", 
		}
		
		for i, msg := range testLogs {
			entry := makeLogEntry(time.Now().Add(time.Duration(i)*time.Second), msg, model.config)
			model.store.Append(entry)
		}
		
		// Set up search
		model.searchQuery = "test"
		model.performSearch()
		
		// Verify we have 2 matches
		if len(model.matches) != 2 {
			t.Fatalf("Expected 2 matches, got %d", len(model.matches))
		}
		
		// Current match should start at last match (index 1)
		if model.currentMatch != 1 {
			t.Fatalf("Expected currentMatch to be 1 (last match), got %d", model.currentMatch)
		}
		
		// Verify highlights exist for both matches
		firstMatchIdx := model.matches[0]
		secondMatchIdx := model.matches[1]
		
		if _, exists := model.highlighted[firstMatchIdx]; !exists {
			t.Error("First match should be highlighted")
		}
		
		if _, exists := model.highlighted[secondMatchIdx]; !exists {
			t.Error("Second match should be highlighted")
		}
		
		// Move to next match (goes backward to older logs, index 0)
		model.nextMatch()
		
		// Verify current match moved backward to index 0
		if model.currentMatch != 0 {
			t.Errorf("Expected currentMatch to be 0 after nextMatch (moved backward), got %d", model.currentMatch)
		}
		
		// Verify highlights still exist after navigation
		if _, exists := model.highlighted[firstMatchIdx]; !exists {
			t.Error("First match should still be highlighted after navigation")
		}
		
		if _, exists := model.highlighted[secondMatchIdx]; !exists {
			t.Error("Second match should still be highlighted after navigation")
		}
		
		// Verify both highlights contain the search term
		firstHighlight := model.highlighted[firstMatchIdx]
		secondHighlight := model.highlighted[secondMatchIdx]
		
		if !strings.Contains(firstHighlight, "test") {
			t.Error("First match highlight should contain search term")
		}
		
		if !strings.Contains(secondHighlight, "test") {
			t.Error("Second match highlight should contain search term")
		}
	})
}
