package main

import (
	"strings"
	"testing"
	"time"
)

// Test JSON detection
func TestIsJSON(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{`{"key":"value"}`, true},
		{`{"nested":{"key":123}}`, true},
		{`[1,2,3]`, true},
		{`not json`, false},
		{`{incomplete`, false},
		{``, false},
	}
	
	for _, tt := range tests {
		result := isJSON(tt.input)
		if result != tt.valid {
			t.Errorf("isJSON(%q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

// Test JSON formatting
func TestFormatJSON(t *testing.T) {
	input := `{"key":"value","number":42}`
	result := formatJSON(input, "  ")
	
	if !strings.Contains(result, "\n") {
		t.Error("Expected formatted JSON to contain newlines")
	}
	
	if !strings.Contains(result, `"key"`) {
		t.Error("Expected formatted JSON to preserve keys")
	}
}

// Test access log parsing
func TestParseAccessLog(t *testing.T) {
	// Valid Apache/Nginx common log format
	logLine := `192.168.1.1 - - [20/Oct/2025:14:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"`
	
	entry := parseAccessLog(logLine)
	if entry == nil {
		t.Fatal("Expected valid access log entry, got nil")
	}
	
	if entry.IP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", entry.IP)
	}
	
	if entry.Method != "GET" {
		t.Errorf("Expected method GET, got %s", entry.Method)
	}
	
	if entry.Path != "/api/users" {
		t.Errorf("Expected path /api/users, got %s", entry.Path)
	}
	
	if entry.Status != "200" {
		t.Errorf("Expected status 200, got %s", entry.Status)
	}
}

// Test access log parsing with invalid input
func TestParseAccessLog_Invalid(t *testing.T) {
	invalidLines := []string{
		"not an access log",
		"",
		"192.168.1.1 incomplete",
	}
	
	for _, line := range invalidLines {
		entry := parseAccessLog(line)
		if entry != nil {
			t.Errorf("Expected nil for invalid line %q, got %+v", line, entry)
		}
	}
}

// Test log message formatting with raw mode
func TestFormatLogMessage_Raw(t *testing.T) {
	config := &UIConfig{
		ParseAccessLogs: false,
		PrettyPrintJSON: false,
	}
	
	input := "  some log message  "
	result := formatLogMessage(input, config)
	
	if result != "some log message" {
		t.Errorf("Expected trimmed message, got %q", result)
	}
}

// Test log message formatting with JSON
func TestFormatLogMessage_JSON(t *testing.T) {
	config := &UIConfig{
		ParseAccessLogs: false,
		PrettyPrintJSON: true,
		JSONIndent:      "  ",
	}
	
	input := `{"level":"info","msg":"test"}`
	result := formatLogMessage(input, config)
	
	if !strings.Contains(result, "\n") {
		t.Error("Expected JSON to be formatted with newlines")
	}
	
	if !strings.Contains(result, `"level"`) {
		t.Error("Expected JSON to preserve content")
	}
}

// Test log entry creation
func TestMakeLogEntry(t *testing.T) {
	config := &UIConfig{
		ParseAccessLogs: false,
		PrettyPrintJSON: false,
	}
	
	ts := time.Date(2025, 10, 20, 14, 30, 0, 0, time.UTC)
	msg := "test message"
	
	entry := makeLogEntry(ts, msg, config)
	
	if entry.Timestamp != ts {
		t.Errorf("Expected timestamp %v, got %v", ts, entry.Timestamp)
	}
	
	if entry.OriginalMessage != msg {
		t.Errorf("Expected original message %q, got %q", msg, entry.OriginalMessage)
	}
	
	if !strings.Contains(entry.Raw, "14:30:00") {
		t.Errorf("Expected Raw to contain formatted time, got %q", entry.Raw)
	}
	
	if !strings.Contains(entry.Raw, msg) {
		t.Errorf("Expected Raw to contain message, got %q", entry.Raw)
	}
}

// Test empty message handling
func TestFormatLogMessage_Empty(t *testing.T) {
	config := &UIConfig{
		ParseAccessLogs: true,
		PrettyPrintJSON: true,
	}
	
	result := formatLogMessage("", config)
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}
