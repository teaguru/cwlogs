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
	logs      []logEntry
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

type backToLogGroupsMsg struct{}
type clearFormatStatusMsg struct{}
type statusMsg string

// Commands
func backToLogGroupsCmd() tea.Msg {
	return backToLogGroupsMsg{}
}

func clearFormatStatusCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearFormatStatusMsg{}
	})
}

// logModel represents the state of the log viewer TUI
type logModel struct {
	profile          string
	logGroup         string
	store            *logStore  // Ring buffer for bounded memory
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
	statusMessage       string
	formatStatusMsg     string         // Separate message for format toggle status
	lastFormatState     bool           // Track last format state to avoid unnecessary reprocessing
	needsLazyReprocess  bool           // Flag to indicate lazy reprocessing is needed
	lastLazyReprocess   int            // Last cursor position where lazy reprocessing occurred
	highlighted         map[int]string // Cache of highlighted lines (by index)
	lastSearchQuery     string         // Track last search query to avoid reprocessing
	backToLogGroups     bool           // Flag to indicate user wants to go back to log group selection
}

// safeLogs returns logs safely, never panics
func (m *logModel) safeLogs() []logEntry {
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

	// Perform lazy reprocessing when cursor moves
	m.lazyReprocessNearby()
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
			cmd := m.toggleFollow()
			return m, cmd
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
			case "b", "backspace":
				// Return to log group selection
				return m, backToLogGroupsCmd
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
				// Start tick cycle for follow mode
				return m, tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				})
			case "H":
				return m, m.fetchHistoryLogs()
			case "c":
				// Copy current log line to clipboard
				return m, m.copyCurrentLine()
			case "end":
				// Jump to latest logs (same as G but more intuitive)
				m.followMode = true
				m.fixCursor()
				// Start tick cycle for follow mode
				return m, tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				})
			}
		}

	case tickMsg:
		// Only fetch logs and schedule next tick if follow mode is enabled
		if m.followMode {
			return m, tea.Batch(
				m.fetchLogs(),
				tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				}),
			)
		}
		// If follow mode is disabled, don't schedule another tick
		return m, nil

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

				m.fixCursor()

				// Delay re-search until next frame for consistency
				if oldQuery != "" {
					m.searchQuery = oldQuery
					return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
						return delayedSearchMsg{oldQuery}
					})
				}
			}

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

	case backToLogGroupsMsg:
		// Set flag to indicate user wants to go back to log group selection
		m.backToLogGroups = true
		return m, tea.Quit

	case clearFormatStatusMsg:
		// Clear the format status message after timeout
		m.formatStatusMsg = ""
		return m, nil
		
	case statusMsg:
		// Set status message (will be displayed in status bar)
		m.statusMessage = string(msg)
		return m, nil
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

	// Mark that we need to reprocess other logs lazily
	m.needsLazyReprocess = true

	// Recalculate highlights if search is active
	if len(m.matches) > 0 {
		m.applyHighlights()
	}

	// Update format status message
	if m.config.ParseAccessLogs {
		m.formatStatusMsg = "Formatted mode enabled"
	} else {
		m.formatStatusMsg = "Raw mode enabled"
	}

	// Return command to clear the format status after 2 seconds
	return clearFormatStatusCmd()
}

// reprocessVisibleLogs regenerates visible logs based on the current format setting
func (m *logModel) reprocessVisibleLogs() {
	logs := m.safeLogs()
	if len(logs) == 0 {
		return
	}

	// Calculate visible range with generous buffer for smooth scrolling
	const uiReservedHeight = 6 // Header + status + borders
	viewportHeight := m.height - uiReservedHeight
	bufferSize := viewportHeight * 2 // 2x viewport for smooth scrolling
	
	start := max(0, m.cursor - bufferSize)
	end := min(len(logs), m.cursor + bufferSize)

	// Only reprocess visible range + buffer for better performance
	for i := start; i < end; i++ {
		entry := makeLogEntry(logs[i].Timestamp, logs[i].OriginalMessage, m.config)
		m.store.UpdateEntry(i, entry)
	}
}

// lazyReprocessNearby reprocesses logs near the cursor when needed
func (m *logModel) lazyReprocessNearby() {
	if !m.needsLazyReprocess {
		return
	}

	logs := m.safeLogs()
	if len(logs) == 0 {
		return
	}

	// Throttle lazy reprocessing - only reprocess if cursor moved significantly
	const minCursorDelta = 10
	if abs(m.cursor - m.lastLazyReprocess) < minCursorDelta {
		return
	}

	// Calculate a small range around cursor for lazy reprocessing
	const uiReservedHeight = 6
	viewportHeight := m.height - uiReservedHeight
	batchSize := max(30, viewportHeight/2) // Smaller batch size for better performance
	
	start := max(0, m.cursor - batchSize/2)
	end := min(len(logs), start + batchSize)

	// Reprocess small batch around cursor
	for i := start; i < end; i++ {
		entry := makeLogEntry(logs[i].Timestamp, logs[i].OriginalMessage, m.config)
		m.store.UpdateEntry(i, entry)
	}

	// Update last reprocess position
	m.lastLazyReprocess = m.cursor

	// Check if we've processed all logs
	if start == 0 && end == len(logs) {
		m.needsLazyReprocess = false
	}
}

// forceCompleteReprocess reprocesses all logs immediately (used before search)
func (m *logModel) forceCompleteReprocess() {
	logs := m.safeLogs()
	if len(logs) == 0 {
		return
	}

	// Reprocess all logs to ensure search accuracy
	for i := 0; i < len(logs); i++ {
		entry := makeLogEntry(logs[i].Timestamp, logs[i].OriginalMessage, m.config)
		m.store.UpdateEntry(i, entry)
	}

	// Mark lazy reprocessing as complete
	m.needsLazyReprocess = false
}

// clearSearchState clears all search-related state
func (m *logModel) clearSearchState() {
	m.searchRegex = nil
	m.matches = nil
	m.currentMatch = 0
	m.highlighted = make(map[int]string) // Clear highlight cache
	m.lastSearchQuery = ""              // Clear last search query to force re-search
	// Keep searchQuery and searchMode so user can re-search if needed
}

// toggleFollow toggles auto-follow mode and returns a command to start/stop ticking
func (m *logModel) toggleFollow() tea.Cmd {
	m.followMode = !m.followMode
	if m.followMode {
		// Jump to the latest log immediately
		logs := m.safeLogs()
		if len(logs) > 0 {
			m.cursor = len(logs) - 1
		}
		// Start the tick cycle for follow mode
		return tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	// Follow mode disabled - no tick needed
	return nil
}
