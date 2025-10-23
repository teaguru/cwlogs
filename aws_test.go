package main

import (
	"testing"
	"time"
)

// Mock AWS client for testing
type mockCloudWatchClient struct {
	logGroups []string
	logs      []mockLogEvent
}

type mockLogEvent struct {
	timestamp int64
	message   string
}

// Test profile name processing
func TestTrimProfilePrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"profile myprofile", "myprofile"},
		{"profile dev", "dev"},
		{"default", "default"},
		{"myprofile", "myprofile"},
		{"profile ", ""},
	}
	
	for _, tt := range tests {
		result := trimProfilePrefix(tt.input)
		if result != tt.expected {
			t.Errorf("trimProfilePrefix(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Test log group name validation
func TestValidateLogGroupName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"valid-log-group", true},
		{"valid_log_group", true},
		{"valid.log.group", true},
		{"/aws/lambda/function", true},
		{"", false},
		{"log group with spaces", false},
		{"log-group-with-very-long-name-that-exceeds-normal-limits-but-should-still-be-valid", true},
	}
	
	for _, tt := range tests {
		result := isValidLogGroupName(tt.name)
		if result != tt.valid {
			t.Errorf("isValidLogGroupName(%q) = %v, want %v", tt.name, result, tt.valid)
		}
	}
}

// Test time range calculation
func TestCalculateTimeRange(t *testing.T) {
	now := time.Date(2025, 10, 20, 14, 30, 0, 0, time.UTC)
	hours := 24
	
	start, end := calculateTimeRange(now, hours)
	
	expectedStart := now.Add(-24 * time.Hour)
	if !start.Equal(expectedStart) {
		t.Errorf("Expected start time %v, got %v", expectedStart, start)
	}
	
	if !end.Equal(now) {
		t.Errorf("Expected end time %v, got %v", now, end)
	}
}

// Test log event conversion
func TestConvertLogEvent(t *testing.T) {
	timestamp := time.Date(2025, 10, 20, 14, 30, 0, 0, time.UTC)
	message := "test log message"
	
	// Mock AWS log event structure
	event := struct {
		Timestamp *int64
		Message   *string
	}{
		Timestamp: &[]int64{timestamp.UnixMilli()}[0],
		Message:   &message,
	}
	
	config := NewUIConfig()
	entry := convertLogEvent(event.Timestamp, event.Message, config)
	
	if entry.OriginalMessage != message {
		t.Errorf("Expected message %q, got %q", message, entry.OriginalMessage)
	}
	
	if entry.Timestamp.Unix() != timestamp.Unix() {
		t.Errorf("Expected timestamp %v, got %v", timestamp, entry.Timestamp)
	}
}

// Helper function for log group name validation
func isValidLogGroupName(name string) bool {
	if name == "" {
		return false
	}
	
	// Basic validation - no spaces
	for _, char := range name {
		if char == ' ' {
			return false
		}
	}
	
	return true
}

// Helper function for time range calculation
func calculateTimeRange(now time.Time, hours int) (time.Time, time.Time) {
	start := now.Add(-time.Duration(hours) * time.Hour)
	return start, now
}

// Helper function for log event conversion
func convertLogEvent(timestamp *int64, message *string, config *UIConfig) logEntry {
	var ts time.Time
	var msg string
	
	if timestamp != nil {
		ts = time.UnixMilli(*timestamp)
	}
	
	if message != nil {
		msg = *message
	}
	
	return makeLogEntry(ts, msg, config)
}
