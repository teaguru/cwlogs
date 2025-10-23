package main

import (
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		// Arrange & Act
		config := NewUIConfig()
		
		// Assert - Performance settings
		assertIntEqual(t, config.RefreshInterval, 5, "refresh interval")
		assertIntEqual(t, int(config.LogsPerFetch), 500, "logs per fetch")
		assertIntEqual(t, config.MaxLogBuffer, 5000, "max log buffer")
		assertIntEqual(t, config.LogTimeRange, 2, "log time range")
		assertIntEqual(t, config.APITimeout, 10, "API timeout")
		
		// Assert - UI settings
		assertIntEqual(t, config.ProfilePageSize, 40, "profile page size")
		assertIntEqual(t, config.LogGroupPageSize, 30, "log group page size")
		assertIntEqual(t, config.DefaultHeight, 24, "default height")
		assertIntEqual(t, config.DefaultWidth, 80, "default width")
		
		// Assert - Formatting settings
		assertBoolEqual(t, config.ParseAccessLogs, true, "parse access logs")
		assertBoolEqual(t, config.PrettyPrintJSON, true, "pretty print JSON")
		assertBoolEqual(t, config.ColorizeFields, true, "colorize fields")
		assertStringEqual(t, config.JSONIndent, "  ")
	})
	
	t.Run("JSONConfiguration", func(t *testing.T) {
		// Arrange & Act
		config := NewUIConfig()
		
		// Assert
		assertStringEqual(t, config.JSONIndent, "  ")
		assertBoolEqual(t, config.PrettyPrintJSON, true, "JSON pretty printing enabled")
	})
	
	t.Run("PerformanceSettings", func(t *testing.T) {
		// Arrange & Act
		config := NewUIConfig()
		
		// Assert - Reasonable defaults for performance
		if config.RefreshInterval < 1 || config.RefreshInterval > 60 {
			t.Errorf("RefreshInterval should be between 1-60 seconds, got %d", config.RefreshInterval)
		}
		
		if config.LogsPerFetch < 100 || config.LogsPerFetch > 1000 {
			t.Errorf("LogsPerFetch should be between 100-1000, got %d", config.LogsPerFetch)
		}
		
		if config.MaxLogBuffer < 1000 || config.MaxLogBuffer > 10000 {
			t.Errorf("MaxLogBuffer should be between 1000-10000, got %d", config.MaxLogBuffer)
		}
	})
	
	t.Run("UISettings", func(t *testing.T) {
		// Arrange & Act
		config := NewUIConfig()
		
		// Assert - UI dimensions are reasonable
		if config.DefaultHeight < 10 || config.DefaultHeight > 100 {
			t.Errorf("DefaultHeight should be between 10-100, got %d", config.DefaultHeight)
		}
		
		if config.DefaultWidth < 40 || config.DefaultWidth > 200 {
			t.Errorf("DefaultWidth should be between 40-200, got %d", config.DefaultWidth)
		}
		
		// Assert - Page sizes are reasonable
		if config.ProfilePageSize < 5 || config.ProfilePageSize > 100 {
			t.Errorf("ProfilePageSize should be between 5-100, got %d", config.ProfilePageSize)
		}
		
		if config.LogGroupPageSize < 5 || config.LogGroupPageSize > 100 {
			t.Errorf("LogGroupPageSize should be between 5-100, got %d", config.LogGroupPageSize)
		}
	})
	
	t.Run("ColorConfiguration", func(t *testing.T) {
		// Arrange & Act
		config := NewUIConfig()
		
		// Assert - Color settings exist and are non-empty
		if config.Colors.HeaderColor == "" {
			t.Error("HeaderColor should not be empty")
		}
		
		if config.Colors.SearchColor == "" {
			t.Error("SearchColor should not be empty")
		}
		
		if config.Colors.MatchColor == "" {
			t.Error("MatchColor should not be empty")
		}
		
		// Assert - Color methods work (just verify they don't panic)
		_ = config.HeaderStyle()
		_ = config.SearchStyle()
		_ = config.MatchStyle()
		_ = config.CursorStyle()
	})
}

// Benchmark config creation (should be fast)
func BenchmarkConfig_NewUIConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewUIConfig()
	}
}
