package main

import (
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	t.Run("JSONDetection", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  bool
		}{
			{"valid object", `{"key":"value"}`, true},
			{"valid nested object", `{"nested":{"key":123}}`, true},
			{"valid array", `[1,2,3]`, true},
			{"valid empty object", `{}`, true},
			{"valid empty array", `[]`, true},
			{"invalid text", `not json`, false},
			{"incomplete object", `{incomplete`, false},
			{"empty string", ``, false},
			{"null value", `null`, true},
			{"boolean value", `true`, true},
			{"number value", `42`, true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Act
				got := isJSON(tt.input)
				
				// Assert
				assertBoolEqual(t, got, tt.want, "JSON detection")
			})
		}
	})
	
	t.Run("JSONFormatting", func(t *testing.T) {
		t.Run("SimpleObject", func(t *testing.T) {
			// Arrange
			input := `{"key":"value","number":42}`
			
			// Act
			result := formatJSON(input, "  ")
			
			// Assert
			assertStringContains(t, result, "\n")
			assertStringContains(t, result, `"key"`)
			assertStringContains(t, result, `"value"`)
		})
		
		t.Run("NestedObject", func(t *testing.T) {
			// Arrange
			input := `{"outer":{"inner":"value"}}`
			
			// Act
			result := formatJSON(input, "  ")
			
			// Assert
			assertStringContains(t, result, "\n")
			assertStringContains(t, result, `"outer"`)
			assertStringContains(t, result, `"inner"`)
		})
		
		t.Run("Array", func(t *testing.T) {
			// Arrange
			input := `[1,2,3]`
			
			// Act
			result := formatJSON(input, "  ")
			
			// Assert
			assertStringContains(t, result, "\n")
		})
	})
	
	t.Run("AccessLogParsing", func(t *testing.T) {
		t.Run("ValidApacheLog", func(t *testing.T) {
			// Arrange
			logLine := `192.168.1.1 - - [20/Oct/2025:14:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"`
			
			// Act
			entry := parseAccessLog(logLine)
			
			// Assert
			if entry == nil {
				t.Fatal("Expected valid access log entry, got nil")
			}
			
			assertStringEqual(t, entry.IP, "192.168.1.1")
			assertStringEqual(t, entry.Method, "GET")
			assertStringEqual(t, entry.Path, "/api/users")
			assertStringEqual(t, entry.Status, "200")
		})
		
		t.Run("ValidNginxLog", func(t *testing.T) {
			// Arrange
			logLine := `10.0.0.1 - user [01/Jan/2025:12:00:00 +0000] "POST /api/login HTTP/1.1" 401 567 "-" "curl/7.68.0"`
			
			// Act
			entry := parseAccessLog(logLine)
			
			// Assert
			if entry == nil {
				t.Fatal("Expected valid access log entry, got nil")
			}
			
			assertStringEqual(t, entry.IP, "10.0.0.1")
			assertStringEqual(t, entry.Method, "POST")
			assertStringEqual(t, entry.Path, "/api/login")
			assertStringEqual(t, entry.Status, "401")
		})
		
		t.Run("InvalidLogs", func(t *testing.T) {
			tests := []struct {
				name string
				line string
			}{
				{"plain text", "not an access log"},
				{"empty string", ""},
				{"incomplete", "192.168.1.1 incomplete"},
				{"malformed", "malformed log entry"},
			}
			
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					// Act
					entry := parseAccessLog(tt.line)
					
					// Assert
					if entry != nil {
						t.Errorf("Expected nil for invalid line %q, got %+v", tt.line, entry)
					}
				})
			}
		})
	})
	
	t.Run("LogMessageFormatting", func(t *testing.T) {
		t.Run("RawMode", func(t *testing.T) {
			// Arrange
			config := &UIConfig{
				ParseAccessLogs: false,
				PrettyPrintJSON: false,
			}
			input := "  some log message  "
			
			// Act
			result := formatLogMessage(input, config)
			
			// Assert
			assertStringEqual(t, result, "some log message")
		})
		
		t.Run("JSONMode", func(t *testing.T) {
			// Arrange
			config := &UIConfig{
				ParseAccessLogs: false,
				PrettyPrintJSON: true,
				JSONIndent:      "  ",
			}
			input := `{"level":"info","msg":"test"}`
			
			// Act
			result := formatLogMessage(input, config)
			
			// Assert
			assertStringContains(t, result, "\n")
			assertStringContains(t, result, `"level"`)
		})
		
		t.Run("AccessLogMode", func(t *testing.T) {
			// Arrange
			config := &UIConfig{
				ParseAccessLogs: true,
				PrettyPrintJSON: false,
			}
			input := `192.168.1.1 - - [20/Oct/2025:14:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234`
			
			// Act
			result := formatLogMessage(input, config)
			
			// Assert
			// Should be formatted as access log
			assertStringContains(t, result, "GET")
			assertStringContains(t, result, "/api/users")
		})
		
		t.Run("EmptyMessage", func(t *testing.T) {
			// Arrange
			config := &UIConfig{
				ParseAccessLogs: true,
				PrettyPrintJSON: true,
			}
			
			// Act
			result := formatLogMessage("", config)
			
			// Assert
			assertStringEqual(t, result, "")
		})
	})
	
	t.Run("LogEntryCreation", func(t *testing.T) {
		t.Run("BasicEntry", func(t *testing.T) {
			// Arrange
			config := &UIConfig{
				ParseAccessLogs: false,
				PrettyPrintJSON: false,
			}
			ts := time.Date(2025, 10, 20, 14, 30, 0, 0, time.UTC)
			msg := "test message"
			
			// Act
			entry := makeLogEntry(ts, msg, config)
			
			// Assert
			if !entry.Timestamp.Equal(ts) {
				t.Errorf("Expected timestamp %v, got %v", ts, entry.Timestamp)
			}
			assertStringEqual(t, entry.OriginalMessage, msg)
			assertStringContains(t, entry.Raw, "14:30:00")
			assertStringContains(t, entry.Raw, msg)
		})
		
		t.Run("JSONEntry", func(t *testing.T) {
			// Arrange
			config := &UIConfig{
				ParseAccessLogs: false,
				PrettyPrintJSON: true,
				JSONIndent:      "  ",
			}
			ts := time.Now()
			msg := `{"level":"error","message":"test"}`
			
			// Act
			entry := makeLogEntry(ts, msg, config)
			
			// Assert
			assertStringEqual(t, entry.OriginalMessage, msg)
			assertStringContains(t, entry.Message, "level")
			assertStringContains(t, entry.Raw, "level")
		})
	})
}

// Benchmark tests for performance-critical parsing operations
func BenchmarkParser_IsJSON(b *testing.B) {
	input := `{"key":"value","nested":{"array":[1,2,3]}}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isJSON(input)
	}
}

func BenchmarkParser_FormatJSON(b *testing.B) {
	input := `{"key":"value","nested":{"array":[1,2,3]},"number":42}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatJSON(input, "  ")
	}
}

func BenchmarkParser_ParseAccessLog(b *testing.B) {
	logLine := `192.168.1.1 - - [20/Oct/2025:14:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234 "https://example.com" "Mozilla/5.0"`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseAccessLog(logLine)
	}
}

func BenchmarkParser_FormatLogMessage(b *testing.B) {
	config := createTestConfig()
	message := `{"timestamp":"2025-10-20T14:30:00Z","level":"info","message":"test log message"}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatLogMessage(message, config)
	}
}
