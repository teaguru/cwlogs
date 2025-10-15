package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load configuration
	uiConfig := NewUIConfig()

	// Display welcome message
	displayWelcome()

	// Select AWS profile
	profile, err := selectAWSProfile(uiConfig)
	if err != nil {
		handleError("selecting AWS profile", err, profile)
	}
	fmt.Printf("Selected AWS profile: %s\n", profile)

	// List CloudWatch log groups
	logGroups, err := listLogGroups(profile)
	if err != nil {
		handleError("listing CloudWatch log groups", err, profile)
	}

	if len(logGroups) == 0 {
		fmt.Println("No log groups found")
		return
	}

	// Let user select a log group
	chosenLogGroup, err := selectLogGroup(logGroups, uiConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting log group: %v\n", err)
		os.Exit(1)
	}

	// Display success message and controls
	displayConnectionSuccess(profile, chosenLogGroup)

	// Start the log viewer
	err = startLogViewer(profile, chosenLogGroup, uiConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting log viewer: %v\n", err)
		os.Exit(1)
	}
}

// handleError handles common AWS errors with helpful messages
func handleError(operation string, err error, profile string) {
	if strings.Contains(err.Error(), "SSO session has expired") || strings.Contains(err.Error(), "sso session") {
		fmt.Fprintf(os.Stderr, "\n❌ AWS SSO session expired\n")
		fmt.Fprintf(os.Stderr, "Please run: aws sso login --profile <root-account>\n")
		fmt.Fprintf(os.Stderr, "Then try again with profile '%s'\n", profile)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Error %s: %v\n", operation, err)
	os.Exit(1)
}

// startLogViewer creates and runs the TUI log viewer
func startLogViewer(profile, logGroupName string, uiConfig *UIConfig) error {
	client, err := createCloudWatchClient(profile)
	if err != nil {
		return err
	}

	model := logModel{
		profile:          profile,
		logGroup:         logGroupName,
		client:           client,
		config:           uiConfig,
		store:            NewLogStore(5000), // Fixed capacity ring buffer
		logs:             []LogEntry{},      // Deprecated, kept for compatibility during migration
		height:           uiConfig.DefaultHeight,
		width:            uiConfig.DefaultWidth,
		initialLoad:      true,
		followMode:       true,
		searchAttempt:    0,
		currentTimeRange: uiConfig.LogTimeRange,
		lastFormatState:  uiConfig.ParseAccessLogs, // Initialize with current config state
		highlighted:      make(map[int]string),     // Initialize highlighted cache
	}

	// Use alt-screen mode with mouse support for better display and copy/paste
	p := tea.NewProgram(&model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}
