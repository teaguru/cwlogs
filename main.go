package main

import (
	"context"
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
			fmt.Fprintf(os.Stderr, "Please run: aws sso login --profile springload\n")
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

	// Let user select a log group
	var chosenLogGroup string
	prompt := &survey.Select{
		Message: "Select CloudWatch log group:",
		Options: logGroups,
		PageSize: uiConfig.LogGroupPageSize,
	}
	survey.AskOne(prompt, &chosenLogGroup)

	fmt.Printf("You selected log group: %s\n", chosenLogGroup)

	// Start interactive log viewer
	fmt.Println("\n📋 Starting log viewer...")
	fmt.Println("Controls: q=quit, /=search, n=next match, N=prev match, ↑↓=scroll")
	
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
type LogEntry struct {
	Timestamp time.Time
	Message   string
	Raw       string
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
}

type tickMsg time.Time
type logsMsg []LogEntry

func startLogViewer(profile, logGroupName string, uiConfig *UIConfig) error {
	ctx := context.Background()
	
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration for profile '%s': %v", profile, err)
	}

	client := cloudwatchlogs.NewFromConfig(cfg)
	
	model := logModel{
		profile:  profile,
		logGroup: logGroupName,
		client:   client,
		config:   uiConfig,
		logs:     []LogEntry{},
		height:   uiConfig.DefaultHeight,
		width:    uiConfig.DefaultWidth,
	}

	p := tea.NewProgram(&model, tea.WithAltScreen())
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
	return func() tea.Msg {
		ctx := context.Background()
		
		endTime := time.Now()
		startTime := endTime.Add(-time.Duration(m.config.LogTimeRange) * time.Hour)

		input := &cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(m.logGroup),
			StartTime:    aws.Int64(startTime.UnixMilli()),
			EndTime:      aws.Int64(endTime.UnixMilli()),
			Limit:        aws.Int32(m.config.LogsPerFetch),
		}

		if m.lastToken != nil {
			input.NextToken = m.lastToken
		}

		output, err := m.client.FilterLogEvents(ctx, input)
		if err != nil {
			return logsMsg{}
		}

		var logs []LogEntry
		for _, event := range output.Events {
			if event.Timestamp != nil && event.Message != nil {
				timestamp := time.UnixMilli(*event.Timestamp)
				raw := fmt.Sprintf("[%s] %s", 
					timestamp.Format("15:04:05"), 
					*event.Message)
				logs = append(logs, LogEntry{
					Timestamp: timestamp,
					Message:   *event.Message,
					Raw:       raw,
				})
			}
		}

		m.lastToken = output.NextToken
		return logsMsg(logs)
	}
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
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.matches = nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
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
			case "n":
				m.nextMatch()
			case "N":
				m.prevMatch()
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.logs)-1 {
					m.cursor++
				}
			case "g":
				m.cursor = 0
			case "G":
				m.cursor = len(m.logs) - 1
			}
		}

	case tickMsg:
		return m, tea.Batch(
			m.fetchLogs(),
			tea.Tick(time.Duration(m.config.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)

	case logsMsg:
		newLogs := []LogEntry(msg)
		if len(newLogs) > 0 {
			m.logs = append(m.logs, newLogs...)
			// Keep only last N logs based on config
			if len(m.logs) > m.config.MaxLogBuffer {
				m.logs = m.logs[len(m.logs)-m.config.MaxLogBuffer:]
			}
			// Auto-scroll to bottom for new logs
			m.cursor = len(m.logs) - 1
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

func (m *logModel) View() string {
	var b strings.Builder
	
	// Header
	b.WriteString(m.config.HeaderStyle().Render(fmt.Sprintf("CloudWatch Logs: %s", m.logGroup)))
	b.WriteString("\n")

	// Search bar
	if m.searchMode {
		b.WriteString(m.config.SearchStyle().Render(fmt.Sprintf("Search: %s_", m.searchQuery)))
	} else if len(m.matches) > 0 {
		b.WriteString(m.config.MatchStyle().Render(fmt.Sprintf("Matches: %d/%d", m.currentMatch+1, len(m.matches))))
	} else {
		b.WriteString("Press / to search, q to quit")
	}
	b.WriteString("\n\n")

	// Calculate viewport
	start := m.cursor - m.height/2
	if start < 0 {
		start = 0
	}
	end := start + m.height
	if end > len(m.logs) {
		end = len(m.logs)
		start = end - m.height
		if start < 0 {
			start = 0
		}
	}

	// Display logs with alternating colors
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

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
