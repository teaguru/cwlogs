package main

import (
	"testing"
)

// Test default configuration values
func TestDefaultConfig(t *testing.T) {
	config := NewUIConfig()
	
	if config.RefreshInterval != 5 {
		t.Errorf("Expected RefreshInterval 5, got %d", config.RefreshInterval)
	}
	
	if config.LogsPerFetch != 500 {
		t.Errorf("Expected LogsPerFetch 500, got %d", config.LogsPerFetch)
	}
	
	if config.MaxLogBuffer != 5000 {
		t.Errorf("Expected MaxLogBuffer 5000, got %d", config.MaxLogBuffer)
	}
	
	if !config.ParseAccessLogs {
		t.Error("Expected ParseAccessLogs to be true by default")
	}
	
	if !config.PrettyPrintJSON {
		t.Error("Expected PrettyPrintJSON to be true by default")
	}
}

// Test JSON indent configuration
func TestJSONIndentConfig(t *testing.T) {
	config := NewUIConfig()
	
	if config.JSONIndent != "  " {
		t.Errorf("Expected JSONIndent '  ', got %q", config.JSONIndent)
	}
}

// Test time range configuration
func TestTimeRangeConfig(t *testing.T) {
	config := NewUIConfig()
	
	if config.LogTimeRange != 2 {
		t.Errorf("Expected LogTimeRange 2, got %d", config.LogTimeRange)
	}
}
