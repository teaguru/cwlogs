package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Version is the application version, can be overridden at build time
var Version = "dev"

func main() {
	// Parse command-line flags
	flagVersion := flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use (skips profile selection)")
	flagRegion := flag.String("region", "", "AWS region to use (overrides profile default)")
	flagHelp := flag.Bool("help", false, "show help")
	
	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CloudWatch Log Viewer - Fast, terminal-based AWS CloudWatch log viewer\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [profile] [region]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  profile               AWS profile name (alternative to --profile flag)\n")
		fmt.Fprintf(os.Stderr, "  region                AWS region (alternative to --region flag)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                         # Interactive selection\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s dev                     # Use 'dev' profile\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s dev us-west-2           # Use 'dev' profile in us-west-2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --profile dev --region us-east-1  # Use flags\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --version               # Show version information\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFor more information, visit: https://github.com/teaguru/cwlogs\n")
	}
	
	flag.Parse()

	// Handle help flag
	if *flagHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Handle version flag
	if *flagVersion {
		fmt.Printf("cwlogs version %s\n", Version)
		os.Exit(0)
	}

	// Load configuration
	uiConfig := NewUIConfig()

	// Display welcome message
	displayWelcome()

	// Select or use provided AWS profile and region
	var profile, region string
	var err error
	
	// Parse profile from flag or positional argument
	if *flagProfile != "" {
		// Use profile from --profile flag
		profile = *flagProfile
		fmt.Printf("Using AWS profile: %s (from --profile flag)\n", profile)
	} else if len(flag.Args()) > 0 {
		// Use profile from positional argument
		profile = flag.Args()[0]
		fmt.Printf("Using AWS profile: %s (from argument)\n", profile)
	}
	
	// Parse region from flag or positional argument
	if *flagRegion != "" {
		// Use region from --region flag
		region = *flagRegion
		fmt.Printf("Using AWS region: %s (from --region flag)\n", region)
	} else if len(flag.Args()) > 1 {
		// Use region from second positional argument
		region = flag.Args()[1]
		fmt.Printf("Using AWS region: %s (from argument)\n", region)
	}
	
	// Check for too many arguments
	if len(flag.Args()) > 2 {
		fmt.Fprintf(os.Stderr, "Error: Too many arguments. Expected: %s [profile] [region]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Use --help for usage information.\n")
		os.Exit(1)
	}
	
	if profile != "" {
		// Validate the profile works by trying to create a client
		_, err = createCloudWatchClient(profile, region)
		if err != nil {
			handleError("validating AWS profile/region", err, profile)
		}
	} else {
		// Interactive profile selection
		profile, err = selectAWSProfile(uiConfig)
		if err != nil {
			handleError("selecting AWS profile", err, profile)
		}
		fmt.Printf("Selected AWS profile: %s\n", profile)
	}
	
	// Initialize region - use CLI override, profile default, or auto-detect
	var currentRegion string
	if region != "" {
		currentRegion = region
		fmt.Printf("Region override: %s\n", currentRegion)
	} else {
		// Try to get region from AWS configuration
		detectedRegion, err := getAWSRegion(profile)
		if err != nil || detectedRegion == "" {
			// If no region found and we're using "default" (likely EC2), try to auto-detect
			if profile == "default" {
				if ec2Region := getEC2Region(); ec2Region != "" {
					currentRegion = ec2Region
					fmt.Printf("Auto-detected EC2 region: %s\n", currentRegion)
				} else {
					// Fallback to us-east-1 if we can't detect
					currentRegion = "us-east-1"
					fmt.Printf("Using fallback region: %s (no region configured)\n", currentRegion)
				}
			} else {
				fmt.Printf("Using region from profile configuration\n")
				// currentRegion stays empty, AWS SDK will use profile's default region
			}
		} else {
			currentRegion = detectedRegion
			fmt.Printf("Using region from profile: %s\n", currentRegion)
		}
	}

	// Main loop to allow going back to log group selection and changing regions
	for {
		// List CloudWatch log groups for current region
		logGroups, err := listLogGroups(profile, currentRegion)
		if err != nil {
			handleError("listing CloudWatch log groups", err, profile)
		}

		if len(logGroups) == 0 {
			if currentRegion != "" {
				fmt.Printf("No log groups found in region %s\n", currentRegion)
			} else {
				fmt.Println("No log groups found in default region")
			}
			// Still show the selector so user can change region
		}

		if currentRegion != "" {
			fmt.Printf("Found %d log groups in region %s\n", len(logGroups), currentRegion)
		} else {
			fmt.Printf("Found %d log groups in default region\n", len(logGroups))
		}

		// Log group selection with region change support
		chosenLogGroup, changeRegion, err := selectLogGroupInteractive(logGroups, uiConfig)
		if err != nil {
			if err.Error() == "selection cancelled" {
				fmt.Println("Selection cancelled")
				return
			}
			fmt.Fprintf(os.Stderr, "Error selecting log group: %v\n", err)
			os.Exit(1)
		}

		// Handle region change request
		if changeRegion {
			fmt.Println("\nChanging region...")
			newRegion, err := selectAWSRegion(uiConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting region: %v\n", err)
				os.Exit(1)
			}

			currentRegion = newRegion
			fmt.Printf("Selected region: %s\n", currentRegion)
			continue // Go back to log group selection with new region
		}

		// Display success message and controls
		displayConnectionSuccess(profile, chosenLogGroup)

		// Inner loop for log viewer (allows going back to log group selection)
		for {
			// Start the log viewer
			exitCode, err := startLogViewer(profile, chosenLogGroup, currentRegion, uiConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error starting log viewer: %v\n", err)
				os.Exit(1)
			}

			if exitCode == 0 {
				// User quit normally
				return
			} else if exitCode == 2 {
				// User wants to go back to log group selection
				fmt.Println("\nReturning to log group selection...")
				break // Break inner loop to go back to log group selection
			}
		}
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
// Returns (exitCode, error) where exitCode: 0=quit, 2=back to log groups
func startLogViewer(profile, logGroupName, region string, uiConfig *UIConfig) (int, error) {
	client, err := createCloudWatchClient(profile, region)
	if err != nil {
		return 0, err
	}

	model := logModel{
		profile:          profile,
		logGroup:         logGroupName,
		client:           client,
		config:           uiConfig,
		store:            newLogStore(5000), // Fixed capacity ring buffer
		height:           uiConfig.DefaultHeight,
		width:            uiConfig.DefaultWidth,
		initialLoad:      true,
		followMode:       true,
		searchAttempt:    0,
		currentTimeRange: uiConfig.LogTimeRange,
		lastFormatState:  uiConfig.ParseAccessLogs, // Initialize with current config state
		highlighted:      make(map[int]string),     // Initialize highlighted cache
	}

	// Use alt-screen mode without mouse capture to allow normal text selection
	p := tea.NewProgram(&model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return 0, err
	}

	// Check if the user wants to go back to log group selection
	if logModel, ok := finalModel.(*logModel); ok && logModel.backToLogGroups {
		return 2, nil // Exit code 2 means go back
	}

	return 0, nil // Exit code 0 means quit normally
}
