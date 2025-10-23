package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Pre-compiled regex patterns for better performance
var (
	// Common Log Format: IP - - [timestamp] "METHOD path protocol" status size "referer" "user-agent"
	accessLogRegex = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*) ([^"]*)" (\d+) (\S+) "([^"]*)" "([^"]*)"`)

	// JSON object detection within log messages - simplified to avoid catastrophic backtracking
	jsonRegex = regexp.MustCompile(`\{[^{}]{0,1000}\}`)
)

// AccessLogEntry represents a parsed access log entry
type AccessLogEntry struct {
	IP        string
	Timestamp string
	Method    string
	Path      string
	Protocol  string
	Status    string
	Size      string
	Referer   string
	UserAgent string
}

// logEntry represents a log entry with original and formatted versions
type logEntry struct {
	Timestamp       time.Time
	OriginalMessage string // Store the original unformatted message
	Message         string // Store the formatted message
	Raw             string // Store the complete display line
}

// isJSON checks if a string is valid JSON with safety checks
func isJSON(s string) bool {
	// Quick checks to avoid expensive JSON parsing
	if len(s) == 0 {
		return false
	}
	
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	
	// Avoid parsing very large JSON to prevent performance issues
	if len(s) > 5000 {
		return false
	}
	
	// Quick check for obvious JSON patterns
	first := s[0]
	if first == '{' || first == '[' || first == '"' || 
	   s == "null" || s == "true" || s == "false" ||
	   (first >= '0' && first <= '9') || first == '-' {
		var js json.RawMessage
		return json.Unmarshal([]byte(s), &js) == nil
	}
	
	return false
}

// formatJSON pretty-prints JSON with syntax highlighting
func formatJSON(jsonStr string, indent string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return jsonStr // Return original if not valid JSON
	}

	formatted, err := json.MarshalIndent(obj, "", indent)
	if err != nil {
		return jsonStr // Return original if formatting fails
	}

	return string(formatted)
}

// parseAccessLog attempts to parse common access log formats (Apache/Nginx)
func parseAccessLog(logLine string) *AccessLogEntry {
	matches := accessLogRegex.FindStringSubmatch(logLine)

	if len(matches) >= 10 {
		return &AccessLogEntry{
			IP:        matches[1],
			Timestamp: matches[2],
			Method:    matches[3],
			Path:      matches[4],
			Protocol:  matches[5],
			Status:    matches[6],
			Size:      matches[7],
			Referer:   matches[8],
			UserAgent: matches[9],
		}
	}

	return nil
}

// formatAccessLog formats an access log entry with colors and structure
func formatAccessLog(entry *AccessLogEntry, config *UIConfig) string {
	if !config.ColorizeFields {
		return fmt.Sprintf("%s %s %s %s %s",
			entry.IP, entry.Method, entry.Path, entry.Status, entry.Size)
	}

	// Color coding based on HTTP status
	var statusStyle lipgloss.Style
	switch {
	case strings.HasPrefix(entry.Status, "2"):
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green for 2xx
	case strings.HasPrefix(entry.Status, "3"):
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Yellow for 3xx
	case strings.HasPrefix(entry.Status, "4"):
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red for 4xx
	case strings.HasPrefix(entry.Status, "5"):
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true) // Dark red for 5xx
	default:
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White for others
	}

	// Method colors with better distinction
	var methodStyle lipgloss.Style
	switch entry.Method {
	case "GET":
		methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) // Blue
	case "POST":
		methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true) // Magenta
	case "PUT":
		methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true) // Cyan
	case "DELETE":
		methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red
	default:
		methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White
	}

	// Better styling for different elements
	ipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))    // Cyan for IP
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White for path
	sizeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray for size

	// Simple, clean single-line format
	return fmt.Sprintf("%s %s %s %s %s",
		ipStyle.Render(entry.IP),
		methodStyle.Render(entry.Method),
		pathStyle.Render(entry.Path),
		statusStyle.Render(entry.Status),
		sizeStyle.Render(entry.Size))
}

// formatLogMessage formats a log message, applying appropriate formatting
func formatLogMessage(message string, config *UIConfig) string {
	// Fast path: if no formatting is enabled, return as-is
	if !config.ParseAccessLogs && !config.PrettyPrintJSON {
		return strings.TrimSpace(message)
	}

	message = strings.TrimSpace(message)

	// Skip formatting for very long messages to prevent performance issues
	const maxMessageLength = 10000
	if len(message) > maxMessageLength {
		return message
	}

	// Try access log parsing first
	if config.ParseAccessLogs {
		if accessEntry := parseAccessLog(message); accessEntry != nil {
			return formatAccessLog(accessEntry, config)
		}
	}

	// Fall back to JSON formatting
	if config.PrettyPrintJSON {
		// Check if the entire message is JSON (but only for reasonable sizes)
		if len(message) < 5000 && isJSON(message) {
			return formatJSON(message, config.JSONIndent)
		}

		// Look for JSON objects within the message (with safety limits)
		if len(message) < 5000 {
			// Use a defer/recover to catch any potential panics from regex
			func() {
				defer func() {
					if recover() != nil {
						// If regex panics, just return the original message
						return
					}
				}()
				
				message = jsonRegex.ReplaceAllStringFunc(message, func(match string) string {
					if len(match) < 2000 && isJSON(match) {
						return formatJSON(match, config.JSONIndent)
					}
					return match
				})
			}()
		}
	}

	return message
}

// makeLogEntry creates log entries consistently
func makeLogEntry(ts time.Time, originalMsg string, cfg *UIConfig) logEntry {
	formatted := formatLogMessage(originalMsg, cfg)
	return logEntry{
		Timestamp:       ts,
		OriginalMessage: originalMsg,
		Message:         formatted,
		Raw:             fmt.Sprintf("[%s] %s", ts.Format("15:04:05"), formatted),
	}
}
