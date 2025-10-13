package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/ini.v1"
)

func main() {
	// Load configuration
	uiConfig := NewUIConfig()
	
	// Display welcome message without border
	welcomeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")).
		Padding(1, 0)
	
	welcome := welcomeStyle.Render("☁️  AWS CloudWatch Logs Viewer")
	fmt.Println(welcome)
	
	// Select AWS profile
	profile, err := selectAWSProfile(uiConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting AWS profile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Selected AWS profile: %s\n", profile)

	// List CloudWatch log groups
	logGroups, err := listLogGroups(profile)
	if err != nil {
		if strings.Contains(err.Error(), "SSO session has expired") || strings.Contains(err.Error(), "sso session") {
			fmt.Fprintf(os.Stderr, "\n❌ AWS SSO session expired\n")
			fmt.Fprintf(os.Stderr, "Please run: aws sso login --profile <root-account>\n")
			fmt.Fprintf(os.Stderr, "Then try again with profile '%s'\n", profile)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error listing CloudWatch log groups: %v\n", err)
		os.Exit(1)
	}

	if len(logGroups) == 0 {
		fmt.Println("No log groups found")
		return
	}

	// Display log group selection title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(1, 0)
	
	title := titleStyle.Render("📋 CloudWatch Log Group Selection")
	fmt.Println(title)
	fmt.Println()
	
	// Let user select a log group
	var chosenLogGroup string
	prompt := &survey.Select{
		Message: "Select CloudWatch log group:",
		Options: logGroups,
		PageSize: uiConfig.LogGroupPageSize,
	}
	survey.AskOne(prompt, &chosenLogGroup)

	// Display success message and controls without border
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Padding(1, 0)
	
	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Padding(0, 0)
	
	successMsg := successStyle.Render(fmt.Sprintf("✅ Connected to: %s → %s", profile, chosenLogGroup))
	controls := controlsStyle.Render(
		"Controls: ↑↓/j/k=scroll, PgUp/PgDn=fast scroll, g/G=top/bottom\n" +
		"          /=search, n/N=next/prev match, J=format, F=follow, H=history, q=quit\n\n" +
		"💡 Auto-follow turns OFF when you scroll up, ON when you reach bottom\n" +
		"💡 Use mouse to select text for copy/paste (Ctrl+C in most terminals)")
	
	fmt.Println(successMsg)
	fmt.Println()
	fmt.Println(controls)
	
	err = startLogViewer(profile, chosenLogGroup, uiConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting log viewer: %v\n", err)
		os.Exit(1)
	}
}

// Read AWS config, extract profiles, and let user select one.
func selectAWSProfile(uiConfig *UIConfig) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".aws", "config")
	cfg, err := ini.Load(path)
	if err != nil {
		return "", fmt.Errorf("failed to read AWS config file at %s: %v", path, err)
	}

	var profiles []string
	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == "DEFAULT" {
			profiles = append(profiles, "default")
			continue
		}
		profiles = append(profiles, trimProfilePrefix(name))
	}

	if len(profiles) == 0 {
		return "", fmt.Errorf("no AWS profiles found in %s", path)
	}

	// Display profile selection title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(1, 0)
	
	title := titleStyle.Render("🔐 AWS Profile Selection")
	fmt.Println(title)
	fmt.Println()
	
	var chosen string
	prompt := &survey.Select{
		Message: "Select AWS profile:",
		Options: profiles,
		PageSize: uiConfig.ProfilePageSize,
	}
	err = survey.AskOne(prompt, &chosen)
	if err != nil {
		return "", fmt.Errorf("failed to select profile: %v", err)
	}

	return chosen, nil
}

// List CloudWatch log groups for the selected profile
func listLogGroups(profile string) ([]string, error) {
	ctx := context.Background()
	
	// Load AWS config with the selected profile
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration for profile '%s': %v", profile, err)
	}

	// Create CloudWatch Logs client
	client := cloudwatchlogs.NewFromConfig(cfg)

	// List log groups
	var logGroups []string
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, &cloudwatchlogs.DescribeLogGroupsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list CloudWatch log groups: %v", err)
		}

		for _, logGroup := range output.LogGroups {
			if logGroup.LogGroupName != nil {
				logGroups = append(logGroups, *logGroup.LogGroupName)
			}
		}
	}

	return logGroups, nil
}

func trimProfilePrefix(name string) string {
	if strings.HasPrefix(name, "profile ") {
		return strings.TrimPrefix(name, "profile ")
	}
	return name
}

// isJSON checks if a string is valid JSON
func isJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
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

// parseAccessLog attempts to parse common access log formats (Apache/Nginx)
func parseAccessLog(logLine string) *AccessLogEntry {
	// Common Log Format: IP - - [timestamp] "METHOD path protocol" status size "referer" "user-agent"
	// Example: 127.0.0.1 - - [10/Oct/2025:15:03:01 +1300] "GET /people/ HTTP/1.0" 200 69151 "https://..." "Mozilla/5.0..."
	
	accessLogRegex := regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*) ([^"]*)" (\d+) (\S+) "([^"]*)" "([^"]*)"`)
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
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)  // Red for 4xx
	case strings.HasPrefix(entry.Status, "5"):
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)  // Dark red for 5xx
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
		methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)  // Red
	default:
		methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White
	}
	
	// Better styling for different elements
	ipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))     // Cyan for IP
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))  // White for path
	sizeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // Gray for size
	
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
	message = strings.TrimSpace(message)
	
	// Try access log parsing first
	if config.ParseAccessLogs {
		if accessEntry := parseAccessLog(message); accessEntry != nil {
			return formatAccessLog(accessEntry, config)
		}
	}
	
	// Fall back to JSON formatting
	if config.PrettyPrintJSON {
		// Check if the entire message is JSON
		if isJSON(message) {
			return formatJSON(message, config.JSONIndent)
		}
		
		// Look for JSON objects within the message
		jsonRegex := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
		message = jsonRegex.ReplaceAllStringFunc(message, func(match string) string {
			if isJSON(match) {
				return formatJSON(match, config.JSONIndent)
			}
			return match
		})
	}
	
	return message
}
type LogEntry struct {
	Timestamp     time.Time
	OriginalMessage string // Store the original unformatted message
	Message       string   // Store the formatted message
	Raw           string   // Store the complete display line
}

type logModel struct {
	profile     string
	logGroup    string
	logs        []LogEntry
	client      *cloudwatchlogs.Client
	config      *UIConfig
	viewport    int
	cursor      int
	searchMode  bool
	searchQuery string
	searchRegex *regexp.Regexp
	matches     []int
	currentMatch int
	height      int
	width       int
	lastToken   *string
	loading     bool
	lastError   error
	initialLoad bool
	fetchCount  int // Track how many fetches we've done
	followMode  bool // Auto-scroll to new logs
	searchAttempt int // Track search window expansion attempts
	currentTimeRange int // Current time range being searched
	statusMessage string // Status message for search expansion
}

type tickMsg time.Time
type logsMsg []LogEntry
type logsErrorMsg error
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

func startLogViewer(profile, logGroupName string, uiConfig *UIConfig) error {
	ctx := context.Background()
	
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration for profile '%s': %v", profile, err)
	}

	client := cloudwatchlogs.NewFromConfig(cfg)
	
	model := logModel{
		profile:          profile,
		logGroup:         logGroupName,
		client:           client,
		config:           uiConfig,
		logs:             []LogEntry{},
		height:           uiConfig.DefaultHeight,
		width:            uiConfig.DefaultWidth,
		initialLoad:      true,
		followMode:       true, // Start in follow mode
		searchAttempt:    0,
		currentTimeRange: uiConfig.LogTimeRange,
	}

	// Use alt-screen mode with mouse support for better display and copy/paste
	p := tea.NewProgram(&model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

func (m *logModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchLogs(),
		tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

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
				// For refresh: only get recent logs (last 10 minutes)
				startTime = endTime.Add(-10 * time.Minute)
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
				return logsErrorMsg(err)
			}

			var logs []LogEntry
			for _, event := range output.Events {
				if event.Timestamp != nil && event.Message != nil {
					timestamp := time.UnixMilli(*event.Timestamp)
					
					// Format the message (apply JSON pretty-printing if applicable)
					formattedMessage := formatLogMessage(*event.Message, m.config)
					
					raw := fmt.Sprintf("[%s] %s", 
						timestamp.Format("15:04:05"), 
						formattedMessage)
					logs = append(logs, LogEntry{
						Timestamp:       timestamp,
						OriginalMessage: *event.Message, // Store original
						Message:         formattedMessage, // Store formatted
						Raw:             raw,
					})
				}
			}

			// Return both logs and pagination info
			return logsWithTokenMsg{logs, output.NextToken, m.initialLoad}
		},
	)
}

func (m *logModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height - 3
		m.width = msg.Width

	case tea.KeyMsg:
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				m.performSearch()
				// Keep follow mode off after search
				m.followMode = false
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.matches = nil
				// Keep follow mode off when canceling search
				m.followMode = false
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
			case "J":
				// Allow format toggle even in search mode
				if m.config.ParseAccessLogs {
					m.config.ParseAccessLogs = false
					m.config.ColorizeFields = false
				} else {
					m.config.ParseAccessLogs = true
					m.config.ColorizeFields = true
				}
				m.reprocessLogs()
			case "F":
				// Allow follow toggle even in search mode
				m.followMode = !m.followMode
			default:
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
				}
			}
		} else {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "/":
				m.searchMode = true
				m.searchQuery = ""
				// Disable follow mode during search to prevent interruption
				m.followMode = false
			case "n":
				m.nextMatch()
				// Keep follow mode off during search navigation
				m.followMode = false
			case "N":
				m.prevMatch()
				// Keep follow mode off during search navigation
				m.followMode = false
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					// Auto-disable follow mode when scrolling up from bottom
					m.followMode = false
				}
			case "down", "j":
				if m.cursor < len(m.logs)-1 {
					m.cursor++
					// Re-enable follow mode when reaching bottom
					if m.cursor == len(m.logs)-1 {
						m.followMode = true
					}
				}
			case "pageup", "ctrl+b":
				// Page up - jump by screen height
				m.cursor -= m.height
				if m.cursor < 0 {
					m.cursor = 0
				}
				// Auto-disable follow mode when paging up
				m.followMode = false
			case "pagedown", "ctrl+f":
				// Page down - jump by screen height
				m.cursor += m.height
				if m.cursor >= len(m.logs) {
					m.cursor = len(m.logs) - 1
					// Re-enable follow mode when reaching bottom
					m.followMode = true
				}
			case "g":
				m.cursor = 0
				// Auto-disable follow mode when going to top
				m.followMode = false
			case "G":
				m.cursor = len(m.logs) - 1
				// Re-enable follow mode when going to bottom
				m.followMode = true
			case "H":
				// Load more history (extend time range)
				return m, m.loadMoreHistory()
			case "F":
				// Toggle follow mode
				m.followMode = !m.followMode
			case "J":
				// Toggle between formatted access logs and raw logs
				if m.config.ParseAccessLogs {
					// Turn off access log parsing (show raw)
					m.config.ParseAccessLogs = false
					m.config.ColorizeFields = false
				} else {
					// Turn on access log parsing (show formatted)
					m.config.ParseAccessLogs = true
					m.config.ColorizeFields = true
				}
				// Reprocess all logs with new formatting
				m.reprocessLogs()
			}
		}

	case tickMsg:
		return m, tea.Batch(
			m.fetchLogs(),
			tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)

	case loadingMsg:
		m.loading = bool(msg)
		
	case logsErrorMsg:
		m.loading = false
		m.lastError = error(msg)
		
	case logsMsg:
		m.loading = false
		m.lastError = nil
		newLogs := []LogEntry(msg)
		if len(newLogs) > 0 {
			// For history loading (H key), prepend older logs to beginning
			// For regular refresh, append new logs to end
			if len(m.logs) > 0 && newLogs[0].Timestamp.Before(m.logs[0].Timestamp) {
				// These are older logs, prepend them
				m.logs = append(newLogs, m.logs...)
				// Don't change cursor position when loading history
			} else {
				// These are newer logs, append them
				m.logs = append(m.logs, newLogs...)
				// Only auto-scroll if follow mode is enabled
				if m.followMode {
					m.cursor = len(m.logs) - 1
				}
				// Otherwise, keep cursor where it was (don't break focus)
			}
			
			// Keep only last N logs based on config
			if len(m.logs) > m.config.MaxLogBuffer {
				if len(newLogs) > 0 && newLogs[0].Timestamp.Before(m.logs[len(m.logs)/2].Timestamp) {
					// If we added old logs, remove from the end
					m.logs = m.logs[:m.config.MaxLogBuffer]
				} else {
					// If we added new logs, remove from the beginning
					m.logs = m.logs[len(m.logs)-m.config.MaxLogBuffer:]
				}
			}
		}
		
	case logsWithTokenMsg:
		m.loading = false
		m.lastError = nil
		
		if len(msg.logs) > 0 {
			// Clear status message when logs are found
			m.statusMessage = ""
			// Always append for initial load and pagination
			m.logs = append(m.logs, msg.logs...)
			// Keep only last N logs based on config
			if len(m.logs) > m.config.MaxLogBuffer {
				m.logs = m.logs[len(m.logs)-m.config.MaxLogBuffer:]
			}
			// Auto-scroll for initial load or if follow mode is enabled
			if msg.isInitial || m.followMode {
				m.cursor = len(m.logs) - 1
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
			if msg.isInitial && len(m.logs) == 0 && m.searchAttempt < 3 {
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

func (m *logModel) performSearch() {
	if m.searchQuery == "" {
		m.matches = nil
		return
	}

	regex, err := regexp.Compile("(?i)" + regexp.QuoteMeta(m.searchQuery))
	if err != nil {
		return
	}

	m.searchRegex = regex
	m.matches = []int{}

	for i, log := range m.logs {
		if regex.MatchString(log.Raw) {
			m.matches = append(m.matches, i)
		}
	}

	if len(m.matches) > 0 {
		m.currentMatch = 0
		m.cursor = m.matches[0]
	}
}

func (m *logModel) nextMatch() {
	if len(m.matches) == 0 {
		return
	}
	m.currentMatch = (m.currentMatch + 1) % len(m.matches)
	m.cursor = m.matches[m.currentMatch]
}

func (m *logModel) prevMatch() {
	if len(m.matches) == 0 {
		return
	}
	m.currentMatch = (m.currentMatch - 1 + len(m.matches)) % len(m.matches)
	m.cursor = m.matches[m.currentMatch]
}

func (m *logModel) loadMoreHistory() tea.Cmd {
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
			return logsErrorMsg(err)
		}

		var logs []LogEntry
		for _, event := range output.Events {
			if event.Timestamp != nil && event.Message != nil {
				timestamp := time.UnixMilli(*event.Timestamp)
				formattedMessage := formatLogMessage(*event.Message, m.config)
				raw := fmt.Sprintf("[%s] %s", 
					timestamp.Format("15:04:05"), 
					formattedMessage)
				logs = append(logs, LogEntry{
					Timestamp:       timestamp,
					OriginalMessage: *event.Message, // Store original
					Message:         formattedMessage, // Store formatted
					Raw:             raw,
				})
			}
		}

		return logsMsg(logs)
	}
}

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

func (m *logModel) reprocessLogs() {
	// Reformat all existing logs with current formatting settings
	for i := range m.logs {
		// Use the original message for reformatting
		formattedMessage := formatLogMessage(m.logs[i].OriginalMessage, m.config)
		m.logs[i].Message = formattedMessage
		m.logs[i].Raw = fmt.Sprintf("[%s] %s", 
			m.logs[i].Timestamp.Format("15:04:05"), 
			formattedMessage)
	}
}

func (m *logModel) View() string {
	// Header
	header := m.config.HeaderStyle().Render(fmt.Sprintf("CloudWatch Logs: %s", m.logGroup))

	// Status bar
	var statusBar string
	if m.searchMode {
		statusBar = m.config.SearchStyle().Render(fmt.Sprintf("Search: %s_ (follow disabled)", m.searchQuery))
	} else if len(m.matches) > 0 {
		statusBar = m.config.MatchStyle().Render(fmt.Sprintf("Matches: %d/%d (follow disabled) | n=next, N=prev, /=new search", m.currentMatch+1, len(m.matches)))
	} else if m.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		statusBar = statusStyle.Render(fmt.Sprintf("🔍 %s", m.statusMessage))
	} else if m.loading {
		loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		statusBar = loadingStyle.Render("⏳ Loading logs...")
	} else if m.lastError != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		statusBar = errorStyle.Render(fmt.Sprintf("❌ Error: %v", m.lastError))
	} else {
		var formatStatus string
		if m.config.ParseAccessLogs {
			formatStatus = "Formatted"
		} else {
			formatStatus = "Raw"
		}
		
		// Show position, total logs, and follow mode
		logInfo := ""
		if len(m.logs) > 0 {
			logInfo = fmt.Sprintf(" | %d/%d logs", m.cursor+1, len(m.logs))
		}
		
		followStatus := "OFF"
		if m.followMode {
			followStatus = "ON"
		}
		
		statusBar = fmt.Sprintf("/ search, J format (%s), F follow (%s), H history, q quit%s", formatStatus, followStatus, logInfo)
	}

	// Calculate viewport
	start := m.cursor - (m.height-6)/2 // Account for header, status, and border
	if start < 0 {
		start = 0
	}
	end := start + (m.height - 6) // Reserve space for UI elements
	if end > len(m.logs) {
		end = len(m.logs)
		start = end - (m.height - 6)
		if start < 0 {
			start = 0
		}
	}

	// Build log content with alternating colors
	var logContent strings.Builder
	for i := start; i < end && i < len(m.logs); i++ {
		line := m.logs[i].Raw
		
		// Apply alternating row colors (zebra striping)
		var lineStyle lipgloss.Style
		if i%2 == 0 {
			lineStyle = m.config.EvenRowStyle()
		} else {
			lineStyle = m.config.OddRowStyle()
		}
		
		// Highlight search matches
		if m.searchRegex != nil {
			line = m.searchRegex.ReplaceAllStringFunc(line, func(match string) string {
				return m.config.HighlightStyle().Render(match)
			})
		}

		// Highlight current line (cursor)
		if i == m.cursor {
			line = m.config.CursorStyle().Render(line)
		} else {
			// Apply alternating colors only if not cursor line
			line = lineStyle.Render(line)
		}

		logContent.WriteString(line)
		if i < end-1 && i < len(m.logs)-1 {
			logContent.WriteString("\n")
		}
	}

	// Create bordered log area
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1).
		Width(m.width - 4).
		Height(m.height - 6)

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
