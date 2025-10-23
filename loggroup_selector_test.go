package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test log group selector initialization
func TestLogGroupSelectorInit(t *testing.T) {
	logGroups := []string{"/aws/lambda/test1", "/aws/lambda/test2", "/aws/lambda/test3"}
	config := NewUIConfig()
	
	model := newLogGroupSelector(logGroups, config)
	
	if len(model.logGroups) != 3 {
		t.Errorf("Expected 3 log groups, got %d", len(model.logGroups))
	}
	
	if model.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", model.cursor)
	}
	
	if model.selected != "" {
		t.Errorf("Expected empty selection, got %q", model.selected)
	}
	
	if model.changeRegion {
		t.Error("Expected changeRegion to be false initially")
	}
	
	if model.quit {
		t.Error("Expected quit to be false initially")
	}
}

// Test navigation in log group selector
func TestLogGroupSelectorNavigation(t *testing.T) {
	logGroups := []string{"/aws/lambda/test1", "/aws/lambda/test2", "/aws/lambda/test3"}
	config := NewUIConfig()
	
	model := newLogGroupSelector(logGroups, config)
	
	// Test down navigation
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(keyMsg)
	
	if m, ok := updatedModel.(*logGroupSelectorModel); ok {
		if m.cursor != 1 {
			t.Errorf("Expected cursor at 1 after down, got %d", m.cursor)
		}
	} else {
		t.Error("Expected model to be *logGroupSelectorModel type")
	}
	
	// Test up navigation
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ = updatedModel.Update(keyMsg)
	
	if m, ok := updatedModel.(*logGroupSelectorModel); ok {
		if m.cursor != 0 {
			t.Errorf("Expected cursor at 0 after up, got %d", m.cursor)
		}
	} else {
		t.Error("Expected model to be *logGroupSelectorModel type")
	}
}

// Test region change in log group selector
func TestLogGroupSelectorRegionChange(t *testing.T) {
	logGroups := []string{"/aws/lambda/test1", "/aws/lambda/test2"}
	config := NewUIConfig()
	
	model := newLogGroupSelector(logGroups, config)
	
	// Test region change key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, cmd := model.Update(keyMsg)
	
	if m, ok := updatedModel.(*logGroupSelectorModel); ok {
		if !m.changeRegion {
			t.Error("Expected changeRegion to be true after 'r' key")
		}
	} else {
		t.Error("Expected model to be *logGroupSelectorModel type")
	}
	
	// Should return tea.Quit command
	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}
}

// Test selection in log group selector
func TestLogGroupSelectorSelection(t *testing.T) {
	logGroups := []string{"/aws/lambda/test1", "/aws/lambda/test2"}
	config := NewUIConfig()
	
	model := newLogGroupSelector(logGroups, config)
	
	// Move cursor to second item
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(keyMsg)
	
	// Select current item
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := updatedModel.Update(keyMsg)
	
	if m, ok := updatedModel.(*logGroupSelectorModel); ok {
		if m.selected != "/aws/lambda/test2" {
			t.Errorf("Expected selected '/aws/lambda/test2', got %q", m.selected)
		}
	} else {
		t.Error("Expected model to be *logGroupSelectorModel type")
	}
	
	// Should return tea.Quit command
	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}
}
