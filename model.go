package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	tea "github.com/charmbracelet/bubbletea"
)

// Message types for the TUI
type tickMsg time.Time
type loadingMsg bool
type logsWithTokenMsg struct {
	logs      []LogEntry
	nextToken *string
	isInitial bool
}
type noLogsFoundMsg struct {
	timeRange int
	canExpand bool
}

type delayedSearchMsg struct {
	query string
}

// logModel represents the state of the log viewer TUI
type logModel struct {
	profile          string
	logGroup         string
	store            *LogStore           // Ring buffer for bounded memory
	logs             []LogEntry          // Deprecated: use safeLogs() instead
	client           *cloudwatchlogs.Client
	config           *UIConfig
	cursor           int
	searchMode       bool
	searchQuery      string
	searchRegex      *regexp.Regexp
	matches          []int
	currentMatch     int
	height           int
	width            int
	lastToken        *string
	loading          bool
	lastError        error
	initialLoad      bool
	fetchCount       int
	followMode       bool
	searchAttempt    int
	currentTimeRange int
	statusMessage    string
	lastFormatState  bool                // Track last format state to avoid unnecessary reprocessing
	highlighted      map[int]string      // Cache of highlighted lines (by index)
	lastSearchQuery  string              // Track last search query to avoid reprocessing
}

// safeLogs returns logs safely, never panics
func (m *logModel) safeLogs() []LogEntry {
	if m.store == nil {
		return nil
	}
	return m.store.Slice()
}

// fixCursor keeps cursor in valid range and handles follow mode
func (m *logModel) fixCursor() {
	logs := m.safeLogs()
	if len(logs) == 0 {
		m.cursor = 0
		return
	}
	
	// Clamp cursor to valid range
	if m.cursor >= len(logs) {
		m.cursor = len(logs) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	
	// Follow mode: always track latest
	if m.followMode {
		m.cursor = len(logs) - 1
	}
}

// centerOnCursor recenters viewport on the current cursor
func (m *logModel) centerOnCursor() {
	m.followMode = false // Always stop following during search navigation
	m.fixCursor()
}

// Init initializes the model
func (m *logModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchLogs(),
		tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

// Update handles messages and updates the model
func (m *logModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height - 3
		m.width = msg.Width

	case tea.KeyMsg:
		key := msg.String()

		// Handle global keys that work in any mode
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "J":
			cmd := m.toggleFormat()
			return m, cmd
		case "F":
			m.toggleFollow()
			return m, nil // Return immediately after handling global keys
		}

		if m.searchMode {
			switch key {
			case "enter":
				m.searchMode = false
				m.performSearch()
				m.followMode = false
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.matches = nil
				m.followMode = false
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
			default:
				if len(key) == 1 && key != "J" && key != "F" {
					m.searchQuery += key
				}
			}
		} else {
			switch key {
			case "/":
				m.searchMode = true
				m.searchQuery = ""
				m.followMode = false
			case "esc":
				// Clear search results and return to normal browsing
				m.clearSearchState()
			case "n":
				m.nextMatch()
				m.followMode = false
			case "N":
				m.prevMatch()
				m.followMode = false
			case "up", "k":
				m.cursor--
				m.followMode = false
				m.fixCursor()
			case "down", "j":
				m.cursor++
				m.fixCursor()
			case "pageup", "ctrl+b":
				m.cursor -= m.height
				m.followMode = false
				m.fixCursor()
			case "pagedown", "ctrl+f":
				m.cursor += m.height
				m.fixCursor()
			case "g":
				m.cursor = 0
				m.followMode = false
				m.fixCursor()
			case "G":
				m.followMode = true
				m.fixCursor()
			case "H":
				return m, m.loadMoreHistory()
			case "end":
				// Jump to latest logs (same as G but more intuitive)
				m.followMode = true
				m.fixCursor()
			}
		}

	case tickMsg:
		return m, tea.Batch(
			m.fetchLogs(),
			tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)

	case delayedSearchMsg:
		// Handle delayed search after buffer rollover
		m.searchQuery = msg.query
		m.performSearch()

	case loadingMsg:
		m.loading = bool(msg)

	case error:
		m.loading = false
		m.lastError = msg

	// This case is now handled by logsWithTokenMsg with isInitial: false

	case logsWithTokenMsg:
		m.loading = false
		m.lastError = nil

		if len(msg.logs) > 0 {
			// Clear status message when logs are found
			m.statusMessage = ""

			// Append to ring buffer and detect if it wrapped
			wrapped := false
			for _, log := range msg.logs {
				if m.store.Append(log) {
					wrapped = true
				}
			}
			
			// If buffer wrapped, invalidate search state
			if wrapped {
				oldQuery := m.searchQuery
				m.clearSearchState()
				m.statusMessage = "Log buffer rolled over"
				
				// Update logs slice first
				m.logs = m.safeLogs()
				m.fixCursor()
				
				// Delay re-search until next frame for consistency
				if oldQuery != "" {
					m.searchQuery = oldQuery
					return m, func() tea.Msg {
						// Small delay to ensure store state is stable
						time.Sleep(50 * time.Millisecond)
						return delayedSearchMsg{oldQuery}
					}
				}
			}

			// Update logs slice for compatibility (will be removed later)
			m.logs = m.safeLogs()

			// Fix cursor after appending logs
			m.fixCursor()
			
			// Always scroll to bottom when follow mode is on
			if m.followMode {
				logs := m.safeLogs()
				if len(logs) > 0 {
					m.cursor = len(logs) - 1
				}
			}
		}

		// Handle pagination for initial load (limit to 3 batches max for speed)
		if msg.isInitial && msg.nextToken != nil && len(msg.logs) > 0 && m.fetchCount < 3 {
			m.lastToken = msg.nextToken
			m.fetchCount++
			// Continue fetching more logs if we have a token and got results
			return m, m.fetchLogs()
		} else {
			// Check if we found no logs and should expand search
			logs := m.safeLogs()
			if msg.isInitial && len(logs) == 0 && m.searchAttempt < 3 {
				return m, m.expandSearchWindow()
			}

			// Initial load complete or no more logs
			m.initialLoad = false
			m.lastToken = nil
			m.fetchCount = 0
		}

	case noLogsFoundMsg:
		m.loading = false
		timeRangeText := m.getTimeRangeText(msg.timeRange)
		if msg.canExpand {
			// Update status message and try expanding the search window
			m.statusMessage = fmt.Sprintf("No logs found, expanding search to %s...", timeRangeText)
			return m, m.fetchLogs()
		} else {
			// No more expansion possible, show final message
			m.lastError = fmt.Errorf("no logs found in the last %s", timeRangeText)
			m.statusMessage = ""
		}
	}

	return m, nil
}

// toggleFormat toggles log formatting between raw and formatted
func (m *logModel) toggleFormat() tea.Cmd {
	// Flip state
	m.config.ParseAccessLogs = !m.config.ParseAccessLogs
	m.config.ColorizeFields = m.config.ParseAccessLogs
	m.config.PrettyPrintJSON = m.config.ParseAccessLogs
	m.lastFormatState = m.config.ParseAccessLogs

	// Clear search state since the log text changes
	m.clearSearchState()

	// Only reprocess visible logs to prevent hanging
	m.reprocessVisibleLogs()

	// Recalculate highlights if search is active
	if len(m.matches) > 0 {
		m.applyHighlights()
	}

	// Update status message
	if m.config.ParseAccessLogs {
		m.statusMessage = "Formatted mode enabled"
	} else {
		m.statusMessage = "Raw mode enabled"
	}

	// Force UI redraw by returning a no-op command
	return func() tea.Msg { return nil }
}

// reprocessVisibleLogs regenerates all logs based on the current format setting
func (m *logModel) reprocessVisibleLogs() {
	// Get current logs
	logs := m.safeLogs()
	if len(logs) == 0 {
		return
	}
	
	// Reprocess each log entry in place
	// This is fast in raw mode since it just trims whitespace
	newStore := NewLogStore(5000)
	for i := range logs {
		entry := makeLogEntry(logs[i].Timestamp, logs[i].OriginalMessage, m.config)
		newStore.Append(entry)
	}
	
	// Replace store with reprocessed logs
	m.store = newStore
	m.logs = m.safeLogs()
}

// lazyReprocessNearby processes a small batch of logs near the cursor
func (m *logModel) lazyReprocessNearby() {
	// Skip lazy reprocessing to prevent accumulation
	// Only reprocess on explicit format toggle
	return
}

// clearSearchState clears all search-related state
func (m *logModel) clearSearchState() {
	m.searchRegex = nil
	m.matches = nil
	m.currentMatch = 0
	m.highlighted = make(map[int]string) // Clear highlight cache
	// Keep searchQuery and searchMode so user can re-search if needed
}

// toggleFollow toggles auto-follow mode
func (m *logModel) toggleFollow() {
	m.followMode = !m.followMode
	if m.followMode {
		// Jump to the latest log immediately
		logs := m.safeLogs()
		if len(logs) > 0 {
			m.cursor = len(logs) - 1
		}
	}
}

// reprocessLogs reformats all existing logs with current formatting settings
func (m *logModel) reprocessLogs() {
	for i := range m.logs {
		entry := makeLogEntry(m.logs[i].Timestamp, m.logs[i].OriginalMessage, m.config)
		m.logs[i].Message = entry.Message
		m.logs[i].Raw = entry.Raw
	}
}
