package main

import (
	"flag"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Test data creation helpers
func createTestLogEntry(message string) LogEntry {
	return LogEntry{
		Timestamp:       time.Now(),
		Message:         message,
		OriginalMessage: message,
		Raw:            message,
	}
}

func createTestLogEntryWithTime(message string, timestamp time.Time) LogEntry {
	return LogEntry{
		Timestamp:       timestamp,
		Message:         message,
		OriginalMessage: message,
		Raw:            message,
	}
}

func createTestLogStore(capacity int) *LogStore {
	return NewLogStore(capacity)
}

func createTestConfig() *UIConfig {
	return NewUIConfig()
}

func createTestLogModel(logGroup string) *logModel {
	return &logModel{
		logGroup:     logGroup,
		store:        NewLogStore(5000),
		followMode:   true,
		cursor:       0,
		currentMatch: -1,
		highlighted:  make(map[int]string),
		config:       createTestConfig(),
	}
}

func createTestLogGroupSelector(logGroups []string) *logGroupSelectorModel {
	return newLogGroupSelector(logGroups, createTestConfig())
}

// Assertion helpers
func assertStoreLength(t *testing.T, store *LogStore, expected int) {
	t.Helper()
	if got := store.Len(); got != expected {
		t.Errorf("store length: got %d, want %d", got, expected)
	}
}

func assertWrapResult(t *testing.T, wrapped bool, shouldWrap bool, operation string) {
	t.Helper()
	if wrapped != shouldWrap {
		t.Errorf("%s wrap result: got %v, want %v", operation, wrapped, shouldWrap)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error containing %q, got nil", expectedMsg)
		return
	}
	if expectedMsg != "" && !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func assertStringEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("string mismatch: got %q, want %q", got, want)
	}
}

func assertStringContains(t *testing.T, got, want string) {
	t.Helper()
	if !contains(got, want) {
		t.Errorf("string %q should contain %q", got, want)
	}
}

func assertBoolEqual(t *testing.T, got, want bool, context string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", context, got, want)
	}
}

func assertIntEqual(t *testing.T, got, want int, context string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", context, got, want)
	}
}

func assertSliceLength(t *testing.T, slice interface{}, expected int, context string) {
	t.Helper()
	var length int
	switch s := slice.(type) {
	case []LogEntry:
		length = len(s)
	case []string:
		length = len(s)
	case []int:
		length = len(s)
	default:
		t.Errorf("unsupported slice type for length assertion")
		return
	}
	
	if length != expected {
		t.Errorf("%s slice length: got %d, want %d", context, length, expected)
	}
}

// Setup helpers for command-line tests
func setupTestFlags(args []string) func() {
	oldArgs := os.Args
	os.Args = args
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	
	return func() {
		os.Args = oldArgs
	}
}

func parseTestFlags() (*bool, *string, *string, error) {
	flagVersion := flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	return flagVersion, flagProfile, flagRegion, err
}

// TUI test helpers
func simulateKeyPress(model tea.Model, key string) (tea.Model, tea.Cmd) {
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	if key == "enter" {
		keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	} else if key == "esc" {
		keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	} else if key == "up" {
		keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	} else if key == "down" {
		keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	}
	
	return model.Update(keyMsg)
}

func simulateWindowResize(model tea.Model, width, height int) (tea.Model, tea.Cmd) {
	msg := tea.WindowSizeMsg{Width: width, Height: height}
	return model.Update(msg)
}

// Utility functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(substr) > 0 && indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Test data generators
func generateTestLogEntries(count int, prefix string) []LogEntry {
	entries := make([]LogEntry, count)
	for i := 0; i < count; i++ {
		entries[i] = createTestLogEntry(prefix + string(rune('A'+i)))
	}
	return entries
}

func generateTestLogGroups(count int) []string {
	groups := make([]string, count)
	for i := 0; i < count; i++ {
		groups[i] = "/aws/lambda/test-service-" + string(rune('A'+i))
	}
	return groups
}
