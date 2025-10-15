package main

import "github.com/charmbracelet/lipgloss"

/*
CloudWatch Logs Viewer Configuration

This file contains all configurable settings for the log viewer.
Adjust these values based on your needs:

PERFORMANCE TUNING:
- For high-volume logs: Increase LogsPerFetch (500-1000), decrease LogTimeRange (1-3 hours)
- For sparse logs: Increase LogTimeRange (6-12 hours), keep LogsPerFetch moderate (200-500)
- For slow connections: Increase APITimeout (20-30 seconds)
- For limited memory: Decrease MaxLogBuffer (5000-10000)

DISPLAY PREFERENCES:
- Disable PrettyPrintJSON if you prefer raw JSON (faster rendering)
- Disable ParseAccessLogs if you don't have web server logs
- Adjust ProfilePageSize based on how many AWS accounts you have
- Modify Colors section for different terminal themes

TERMINAL COLOR CODES:
0=black, 1=red, 2=green, 3=yellow, 4=blue, 5=magenta, 6=cyan, 7=white
8=bright_black, 9=bright_red, 10=bright_green, 11=bright_yellow
12=bright_blue, 13=bright_magenta, 14=bright_cyan, 15=bright_white
245=gray, 246=light_gray, etc.
*/

// UI Configuration - All settings for the CloudWatch logs viewer
type UIConfig struct {
	// ========== SELECTION UI SETTINGS ==========
	// How many items to show in selection menus before scrolling
	ProfilePageSize  int // AWS profiles shown at once (recommended: 20-50)
	LogGroupPageSize int // Log groups shown at once (recommended: 10-20)
	
	// ========== LOG VIEWER DISPLAY SETTINGS ==========
	DefaultHeight    int // Initial terminal height in lines (auto-adjusts to actual terminal)
	DefaultWidth     int // Initial terminal width in chars (auto-adjusts to actual terminal)
	
	// ========== PERFORMANCE & FETCHING SETTINGS ==========
	RefreshInterval  int   // How often to fetch new logs (seconds) - lower = more real-time but more API calls
	MaxLogBuffer     int   // Maximum logs kept in memory - higher = more history but more RAM usage
	LogsPerFetch     int32 // Logs fetched per API call - higher = fewer API calls but slower initial load
	LogTimeRange     int   // How far back to look for logs (hours) - increase if logs are sparse
	APITimeout       int   // AWS API call timeout (seconds) - increase for slow connections
	
	// ========== LOG FORMATTING SETTINGS ==========
	PrettyPrintJSON  bool   // Auto-detect and pretty-print JSON in log messages
	JSONIndent       string // Indentation for JSON formatting (e.g., "  " for 2 spaces, "\t" for tabs)
	ParseAccessLogs  bool   // Auto-detect and colorize Apache/Nginx access logs
	ColorizeFields   bool   // Apply color coding to parsed log fields (status codes, methods, etc.)
	
	// ========== COLOR SCHEME ==========
	Colors ColorScheme
}

type ColorScheme struct {
	// ========== HEADER & UI COLORS ==========
	HeaderColor string // Color for the main header showing log group name
	
	// ========== SEARCH INTERFACE COLORS ==========
	SearchColor string // Color for search input text
	MatchColor  string // Color for match counter display
	
	// ========== LOG LINE DISPLAY COLORS ==========
	EvenRowColor   string // Text color for even-numbered log lines (zebra striping)
	OddRowColor    string // Text color for odd-numbered log lines (zebra striping)
	CursorBgColor  string // Background color for currently selected log line
	CursorFgColor  string // Text color for currently selected log line
	
	// ========== SEARCH MATCH HIGHLIGHTING ==========
	MatchBgColor string // Background color for search matches within log text
	MatchFgColor string // Text color for search matches within log text
}

// NewUIConfig creates the default configuration with optimized settings for most use cases
func NewUIConfig() *UIConfig {
	return &UIConfig{
		// ========== SELECTION UI SETTINGS ==========
		ProfilePageSize:  40, // Show 40 AWS profiles at once (good for orgs with many accounts)
		LogGroupPageSize: 30, // Show 30 log groups at once (more visible options)
		
		// ========== LOG VIEWER DISPLAY SETTINGS ==========
		DefaultHeight:   24, // Standard terminal height (will auto-adjust to actual size)
		DefaultWidth:    80, // Standard terminal width (will auto-adjust to actual size)
		
		// ========== PERFORMANCE & FETCHING SETTINGS ==========
		RefreshInterval: 5,    // Refresh every 5 seconds (good balance of real-time vs API usage)
		MaxLogBuffer:    5000,  // Keep 5k logs in memory (good balance of history vs speed)
		LogsPerFetch:    500,   // Fetch 500 logs per API call (faster initial load)
		LogTimeRange:    2,     // Look back 2 hours (fast loading, covers recent activity)
		APITimeout:      10,    // 10 second timeout (faster failure for slow responses)
		
		// ========== LOG FORMATTING SETTINGS ==========
		PrettyPrintJSON: true,   // Enable JSON pretty-printing by default
		JSONIndent:      "  ",   // Use 2 spaces for JSON indentation (readable but compact)
		ParseAccessLogs: true,   // Enable access log parsing by default
		ColorizeFields:  true,   // Enable field colorization for better readability
		
		// ========== COLOR SCHEME ==========
		Colors: ColorScheme{
			// Header & UI colors
			HeaderColor: "12", // Blue - stands out for log group identification
			
			// Search interface colors  
			SearchColor: "11", // Yellow - bright and attention-grabbing for search mode
			MatchColor:  "10", // Green - positive color for successful matches
			
			// Log line display colors (zebra striping for readability)
			EvenRowColor:  "245", // Light gray - subtle for alternating rows
			OddRowColor:   "15",  // White - high contrast with even rows
			CursorBgColor: "8",   // Dark gray - clear selection indicator
			CursorFgColor: "15",  // White - high contrast text on cursor
			
			// Search match highlighting
			MatchBgColor: "10", // Green background - makes matches pop out
			MatchFgColor: "0",  // Black text - ensures readability on green background
		},
	}
}

// Style helpers
func (c *UIConfig) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.Colors.HeaderColor))
}

func (c *UIConfig) SearchStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Colors.SearchColor))
}

func (c *UIConfig) MatchStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Colors.MatchColor))
}

func (c *UIConfig) EvenRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Colors.EvenRowColor))
}

func (c *UIConfig) OddRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Colors.OddRowColor))
}

func (c *UIConfig) CursorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color(c.Colors.CursorBgColor)).Foreground(lipgloss.Color(c.Colors.CursorFgColor))
}

func (c *UIConfig) HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color(c.Colors.MatchBgColor)).Foreground(lipgloss.Color(c.Colors.MatchFgColor))
}
