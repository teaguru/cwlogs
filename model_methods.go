package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI layout constants to replace magic numbers
const (
	uiReservedHeight = 6 // Header + status + borders + padding
	contentPadding   = 8 // Left/right padding for content
)

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// ANSI escape sequence regex for cleaning colored text before search
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// fetchLogs fetches logs from CloudWatch
func (m *logModel) fetchLogs() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return loadingMsg(true) },
		func() tea.Msg {
			// Create context with timeout to prevent UI freezing
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Duration(m.config.APITimeout)*time.Second)
			defer cancel()

			endTime := time.Now()
			var startTime time.Time
			var limit int32

			if m.initialLoad {
				// For initial load: use current time range (may be expanded)
				startTime = endTime.Add(-time.Duration(m.currentTimeRange) * time.Hour)
				limit = int32(m.config.LogsPerFetch)
			} else {
				// For refresh: get recent logs with adaptive window
				// Shorter window when following for near-realtime updates
				window := 2 * time.Minute
				if m.followMode {
					window = 1 * time.Minute // Near-realtime when following
				}
				startTime = endTime.Add(-window)
				limit = int32(m.config.LogsPerFetch)
			}

			input := &cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: aws.String(m.logGroup),
				StartTime:    aws.Int64(startTime.UnixMilli()),
				EndTime:      aws.Int64(endTime.UnixMilli()),
				Limit:        aws.Int32(limit),
			}

			// Only use NextToken for initial load pagination, not for refresh
			if m.initialLoad && m.lastToken != nil {
				input.NextToken = m.lastToken
			}

			output, err := m.client.FilterLogEvents(ctx, input)
			if err != nil {
				return err
			}

			var logs []logEntry
			for _, event := range output.Events {
				if event.Timestamp != nil && event.Message != nil {
					timestamp := time.UnixMilli(*event.Timestamp)
					logs = append(logs, makeLogEntry(timestamp, *event.Message, m.config))
				}
			}

			// Return both logs and pagination info
			return logsWithTokenMsg{logs, output.NextToken, m.initialLoad}
		},
	)
}

// fetchHistoryLogs loads older logs by extending the time range
func (m *logModel) fetchHistoryLogs() tea.Cmd {
	return func() tea.Msg {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(),
			time.Duration(m.config.APITimeout)*time.Second)
		defer cancel()

		// Get older logs (extend back further)
		endTime := time.Now().Add(-time.Duration(m.config.LogTimeRange) * time.Hour)
		startTime := endTime.Add(-time.Duration(m.config.LogTimeRange) * time.Hour)

		input := &cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(m.logGroup),
			StartTime:    aws.Int64(startTime.UnixMilli()),
			EndTime:      aws.Int64(endTime.UnixMilli()),
			Limit:        aws.Int32(m.config.LogsPerFetch),
		}

		output, err := m.client.FilterLogEvents(ctx, input)
		if err != nil {
			return err
		}

		var logs []logEntry
		for _, event := range output.Events {
			if event.Timestamp != nil && event.Message != nil {
				timestamp := time.UnixMilli(*event.Timestamp)
				logs = append(logs, makeLogEntry(timestamp, *event.Message, m.config))
			}
		}

		return logsWithTokenMsg{logs: logs, nextToken: nil, isInitial: false}
	}
}

// expandSearchWindow expands the search time window when no logs are found
func (m *logModel) expandSearchWindow() tea.Cmd {
	return func() tea.Msg {
		m.searchAttempt++

		// Progressive expansion: 2h -> 24h -> 7d -> 30d
		switch m.searchAttempt {
		case 1:
			m.currentTimeRange = 24 // 24 hours
		case 2:
			m.currentTimeRange = 24 * 7 // 7 days
		case 3:
			m.currentTimeRange = 24 * 30 // 30 days
		default:
			// No more expansion
			return noLogsFoundMsg{timeRange: m.currentTimeRange, canExpand: false}
		}

		return noLogsFoundMsg{timeRange: m.currentTimeRange, canExpand: true}
	}
}

// getTimeRangeText converts hours to human-readable text
func (m *logModel) getTimeRangeText(hours int) string {
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	} else if hours < 24*7 {
		days := hours / 24
		return fmt.Sprintf("%d days", days)
	} else if hours < 24*30 {
		weeks := hours / (24 * 7)
		return fmt.Sprintf("%d weeks", weeks)
	} else {
		months := hours / (24 * 30)
		return fmt.Sprintf("%d months", months)
	}
}

// performSearch performs a search on the current logs
func (m *logModel) performSearch() {
	if m.searchQuery == "" {
		m.matches = nil
		m.highlighted = make(map[int]string)
		return
	}

	// Skip reprocessing if the query hasn't changed AND we don't need lazy reprocessing
	if m.searchQuery == m.lastSearchQuery && !m.needsLazyReprocess {
		return
	}
	m.lastSearchQuery = m.searchQuery

	// If we need lazy reprocessing, force complete reprocessing before search
	// to ensure search results are accurate
	if m.needsLazyReprocess {
		m.forceCompleteReprocess()
	}

	regex, err := regexp.Compile("(?i)" + regexp.QuoteMeta(m.searchQuery))
	if err != nil {
		// Set status message for regex error
		m.statusMessage = fmt.Sprintf("Invalid search pattern: %v", err)
		return
	}

	m.searchRegex = regex
	m.matches = []int{}

	// Always search full buffer, not just visible slice
	logs := m.safeLogs()
	for i, log := range logs {
		// Always search in the display text (what user sees) to ensure highlighting works
		displayText := stripANSI(log.Raw)
		
		if regex.MatchString(displayText) {
			m.matches = append(m.matches, i)
		}
	}

	if len(m.matches) > 0 {
		// Start at the LAST match (most recent/latest log)
		m.currentMatch = len(m.matches) - 1
		// Defensive bounds check
		if m.currentMatch >= 0 && m.currentMatch < len(m.matches) {
			m.cursor = m.matches[m.currentMatch]
		}
		m.followMode = false // Prevent tick from overwriting cursor
		m.centerOnCursor()   // Center viewport on found match
		m.statusMessage = fmt.Sprintf("Found %d matches", len(m.matches))
	} else {
		m.statusMessage = fmt.Sprintf("No matches found for '%s'", m.searchQuery)
	}

	// Precompute highlighted lines
	m.applyHighlights()
}

// nextMatch moves to the next search match (backward in time to older logs)
func (m *logModel) nextMatch() {
	if len(m.matches) == 0 {
		return
	}
	// Go backward in the matches array (to older logs)
	m.currentMatch = (m.currentMatch - 1 + len(m.matches)) % len(m.matches)
	
	// Bounds check before accessing matches array
	if m.currentMatch >= 0 && m.currentMatch < len(m.matches) {
		m.cursor = m.matches[m.currentMatch]
	}
	m.followMode = false        // Disable follow persistently
	m.centerOnCursor()          // Center viewport on match
	m.applyHighlights()         // Refresh all highlights with new current match
}

// prevMatch moves to the previous search match (forward in time to newer logs)
func (m *logModel) prevMatch() {
	if len(m.matches) == 0 {
		return
	}
	// Go forward in the matches array (to newer logs)
	m.currentMatch = (m.currentMatch + 1) % len(m.matches)
	
	// Bounds check before accessing matches array
	if m.currentMatch >= 0 && m.currentMatch < len(m.matches) {
		m.cursor = m.matches[m.currentMatch]
	}
	m.followMode = false        // Disable follow persistently
	m.centerOnCursor()          // Center viewport on match
	m.applyHighlights()         // Refresh all highlights with new current match
}

// View renders the TUI
func (m *logModel) View() string {
	// Crash-proof rendering: skip frame instead of panic
	defer func() { recover() }()

	// Guard against invalid dimensions
	if m.width <= 0 || m.height <= 0 {
		return "Initializing..."
	}

	// Header
	header := m.config.HeaderStyle().Render(fmt.Sprintf("CloudWatch Logs: %s", m.logGroup))

	// Build status line
	var statusBar string

	switch {
	case m.searchMode:
		statusBar = m.config.SearchStyle().
			Render(fmt.Sprintf("Search: %s_ (follow disabled)", m.searchQuery))
	case len(m.matches) > 0:
		statusBar = m.config.MatchStyle().
			Render(fmt.Sprintf("Matches: %d/%d (follow disabled) | n=next, N=prev, /=new search",
				m.currentMatch+1, len(m.matches)))
	case m.formatStatusMsg != "":
		statusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Render(fmt.Sprintf("⚙️  %s", m.formatStatusMsg))
	case m.statusMessage != "":
		statusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(fmt.Sprintf("🔍 %s", m.statusMessage))
	case m.loading:
		statusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("⏳ Loading logs...")
	case m.lastError != nil:
		statusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("❌ Error: %v", m.lastError))
	}

	// Always render the controls menu below the status
	controls := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		PaddingTop(1).
		Render(renderControlsBar(m))

	if statusBar != "" {
		statusBar = fmt.Sprintf("%s\n%s", statusBar, controls)
	} else {
		statusBar = controls
	}

	// Get logs safely
	logs := m.safeLogs()
	if len(logs) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, statusBar, "", "No logs yet")
	}

	// Calculate viewport
	viewportHeight := m.height - uiReservedHeight
	start := m.cursor - viewportHeight/2 // Center cursor in viewport
	if start < 0 {
		start = 0
	}
	end := start + viewportHeight // Reserve space for UI elements
	if end > len(logs) {
		end = len(logs)
		start = end - viewportHeight
		if start < 0 {
			start = 0
		}
	}

	// Build log content with clean visual isolation
	var logContent strings.Builder
	logContent.Grow(4096)

	for i := start; i < end && i < len(logs); i++ {
		entry := logs[i]
		line := entry.Raw

		// 1) Use precomputed highlight if present
		if hl, ok := m.highlighted[i]; ok {
			line = hl
		}

		// 2) Soft-wrap BEFORE styling so styles don't get re-rendered
		if m.width > contentPadding {
			line = lipgloss.NewStyle().
				MaxWidth(m.width - contentPadding).
				Render(line)
		}

		// 3) Split multi-line entries to isolate per-visual line rendering
		subLines := strings.Split(line, "\n")

		for j, sub := range subLines {
			var rendered string

			// 4) Apply cursor ONLY to the selected logical row (first subline)
			if i == m.cursor && j == 0 {
				if _, hasHighlight := m.highlighted[i]; hasHighlight {
					// Don't overwrite highlights - just add cursor indicator
					rendered = "▌ " + sub
				} else {
					// No highlights, safe to apply cursor style
					rendered = m.config.CursorStyle().Render(sub)
				}
			} else {
				// Base zebra for this logical row (NOT per subline)
				base := m.config.EvenRowStyle()
				if i%2 != 0 {
					base = m.config.OddRowStyle()
				}
				rendered = base.Render(sub)
			}

			logContent.WriteString(rendered)
			if j < len(subLines)-1 {
				logContent.WriteString("\n")
			}
		}

		// 5) Separator between logical rows
		if i < end-1 && i < len(logs)-1 {
			logContent.WriteString("\n")
		}
	}

	// Stable border with fixed width calculation
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("8")).
		Width(m.width - 4). // Fixed width: terminal width minus border chars
		Padding(0, 1)

	borderedLogs := borderStyle.Render(logContent.String())

	// Combine all elements
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		statusBar,
		"",
		borderedLogs,
	)
}

// copyCurrentLine copies the current log line to clipboard
func (m *logModel) copyCurrentLine() tea.Cmd {
	logs := m.safeLogs()
	if len(logs) == 0 || m.cursor < 0 || m.cursor >= len(logs) {
		return func() tea.Msg {
			return statusMsg("No log line to copy")
		}
	}
	
	// Get the current log line (use original unformatted message for copying)
	logLine := logs[m.cursor].OriginalMessage
	
	// Copy to clipboard using different methods based on OS
	return func() tea.Msg {
		if err := copyToClipboard(logLine); err != nil {
			return statusMsg(fmt.Sprintf("Failed to copy: %v", err))
		}
		return statusMsg("Log line copied to clipboard")
	}
}

// applyHighlights precomputes highlighted lines for search matches
func (m *logModel) applyHighlights() {
	if m.searchRegex == nil {
		m.highlighted = make(map[int]string)
		return
	}

	logs := m.safeLogs()
	m.highlighted = make(map[int]string, len(m.matches))
	
	// Get current match index for special highlighting
	var currentMatchIdx int = -1
	if len(m.matches) > 0 && m.currentMatch >= 0 && m.currentMatch < len(m.matches) {
		currentMatchIdx = m.matches[m.currentMatch]
	}
	
	for _, idx := range m.matches {
		if idx < 0 || idx >= len(logs) {
			continue
		}
		
		// Always highlight in the display text (what user sees)
		displayText := stripANSI(logs[idx].Raw)

		// Apply highlights - use different style for current match
		var highlighted string
		if idx == currentMatchIdx {
			// Current match gets special highlighting
			highlighted = m.searchRegex.ReplaceAllStringFunc(displayText, func(match string) string {
				return m.config.CurrentMatchStyle().Render(match)
			})
		} else {
			// Regular matches get normal highlighting
			highlighted = m.searchRegex.ReplaceAllStringFunc(displayText, func(match string) string {
				return m.config.HighlightStyle().Render(match)
			})
		}

		m.highlighted[idx] = highlighted
	}
}



// renderControlsBar builds the controls/help line shown at the bottom of the TUI.
// It reuses the same menu for both formatted and raw modes.
func renderControlsBar(m *logModel) string {
	var formatStatus string
	if m.config.ParseAccessLogs {
		formatStatus = "Formatted"
	} else {
		formatStatus = "Raw"
	}

	followStatus := "OFF"
	if m.followMode && !m.searchMode && len(m.matches) == 0 {
		followStatus = "ON"
	}

	logInfo := ""
	logs := m.safeLogs()
	if len(logs) > 0 {
		logInfo = fmt.Sprintf(" | %d/%d logs", m.cursor+1, len(logs))
	}

	// Build help text with priority order (most important commands first)
	essentialControls := fmt.Sprintf("b back, q quit%s", logInfo)
	
	// Try different levels of detail based on available width
	fullControls := fmt.Sprintf(
		"/ search, Esc clear, n/N next, c copy, J fmt (%s), F follow (%s), H hist, %s",
		formatStatus, followStatus, essentialControls,
	)
	
	mediumControls := fmt.Sprintf(
		"/ search, n/N next, c copy, J fmt (%s), F follow (%s), %s",
		formatStatus, followStatus, essentialControls,
	)
	
	shortControls := fmt.Sprintf(
		"/ search, c copy, J fmt (%s), F follow (%s), %s",
		formatStatus, followStatus, essentialControls,
	)
	
	// Choose the longest version that fits
	var controlsText string
	if len(fullControls) <= m.width {
		controlsText = fullControls
	} else if len(mediumControls) <= m.width {
		controlsText = mediumControls
	} else if len(shortControls) <= m.width {
		controlsText = shortControls
	} else {
		controlsText = essentialControls // Always show back and quit
	}

	return lipgloss.NewStyle().
		MaxWidth(m.width).
		Render(controlsText)
}
