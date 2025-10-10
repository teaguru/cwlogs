package main

import "github.com/charmbracelet/lipgloss"

// UI Configuration
type UIConfig struct {
	// Survey settings
	ProfilePageSize  int
	LogGroupPageSize int
	
	// Log viewer settings
	DefaultHeight    int
	DefaultWidth     int
	RefreshInterval  int // seconds
	MaxLogBuffer     int
	LogsPerFetch     int32
	LogTimeRange     int // hours
	
	// Colors
	Colors ColorScheme
}

type ColorScheme struct {
	// Header colors
	HeaderColor string
	
	// Search colors
	SearchColor     string
	SearchHighlight string
	MatchColor      string
	
	// Log line colors
	EvenRowColor   string
	OddRowColor    string
	CursorBgColor  string
	CursorFgColor  string
	
	// Search match colors
	MatchBgColor string
	MatchFgColor string
}

// Default configuration
func NewUIConfig() *UIConfig {
	return &UIConfig{
		// Survey settings
		ProfilePageSize:  40,
		LogGroupPageSize: 15,
		
		// Log viewer settings
		DefaultHeight:   24,
		DefaultWidth:    80,
		RefreshInterval: 5,
		MaxLogBuffer:    1000,
		LogsPerFetch:    50,
		LogTimeRange:    1,
		
		// Colors
		Colors: ColorScheme{
			// Header
			HeaderColor: "12", // Blue
			
			// Search
			SearchColor:     "11", // Yellow
			MatchColor:      "10", // Green
			
			// Log lines
			EvenRowColor:  "245", // Light gray
			OddRowColor:   "15",  // White
			CursorBgColor: "8",   // Dark gray
			CursorFgColor: "15",  // White
			
			// Search matches
			MatchBgColor: "10", // Green background
			MatchFgColor: "0",  // Black text
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
