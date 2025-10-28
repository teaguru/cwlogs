package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"gopkg.in/ini.v1"
)

// selectAWSProfile reads AWS config and credentials, extracts profiles, and lets user select one
func selectAWSProfile(uiConfig *UIConfig) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	profiles := make(map[string]bool) // Use map to avoid duplicates

	// Check ~/.aws/config for profiles
	configPath := filepath.Join(home, ".aws", "config")
	if cfg, err := ini.Load(configPath); err == nil {
		for _, section := range cfg.Sections() {
			name := section.Name()
			if name == "DEFAULT" {
				profiles["default"] = true
			} else {
				profiles[trimProfilePrefix(name)] = true
			}
		}
	}

	// Check ~/.aws/credentials for profiles (common with `aws configure`)
	credentialsPath := filepath.Join(home, ".aws", "credentials")
	if creds, err := ini.Load(credentialsPath); err == nil {
		for _, section := range creds.Sections() {
			name := section.Name()
			if name != "DEFAULT" && name != "" {
				profiles[name] = true
			} else if name == "DEFAULT" {
				profiles["default"] = true
			}
		}
	}

	// Convert map to slice
	var profileList []string
	for profile := range profiles {
		profileList = append(profileList, profile)
	}

	// If no profiles found, try to use default
	if len(profileList) == 0 {
		// Check if we can load default AWS config (environment variables, etc.)
		ctx := context.Background()
		if _, err := config.LoadDefaultConfig(ctx); err == nil {
			profileList = append(profileList, "default")
		} else {
			return "", fmt.Errorf("no AWS profiles found and default configuration failed: %w\n\nOptions to fix this:\n1. Run 'aws configure' to set up credentials\n2. Set environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY\n3. On EC2: attach an IAM role with CloudWatch permissions\n4. Use AWS SSO: 'aws configure sso'", err)
		}
	}

	// If only one profile, use it automatically
	if len(profileList) == 1 {
		fmt.Printf("Using AWS profile: %s\n\n", profileList[0])
		return profileList[0], nil
	}

	// Display profile selection title
	printStyled("🔐 AWS Profile Selection", "12", true)
	fmt.Println()

	var chosen string
	prompt := &survey.Select{
		Message:  "Select AWS profile:",
		Options:  profileList,
		PageSize: uiConfig.ProfilePageSize,
	}
	err = survey.AskOne(prompt, &chosen)
	if err != nil {
		return "", fmt.Errorf("failed to select profile: %w", err)
	}

	return chosen, nil
}

// listLogGroups lists CloudWatch log groups for the selected profile and optional region
func listLogGroups(profile string, region ...string) ([]string, error) {
	// Use the updated createCloudWatchClient function
	client, err := createCloudWatchClient(profile, region...)
	if err != nil {
		return nil, err
	}

	// List log groups
	ctx := context.Background()
	var logGroups []string
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, &cloudwatchlogs.DescribeLogGroupsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list CloudWatch log groups: %w", err)
		}

		for _, logGroup := range output.LogGroups {
			if logGroup.LogGroupName != nil {
				logGroups = append(logGroups, *logGroup.LogGroupName)
			}
		}
	}

	return logGroups, nil
}

// createCloudWatchClient creates a CloudWatch Logs client for the given profile and optional region
func createCloudWatchClient(profile string, region ...string) (*cloudwatchlogs.Client, error) {
	ctx := context.Background()
	
	var configOptions []func(*config.LoadOptions) error
	
	// Only use shared config profile if it's not the fallback "default"
	// This allows IAM roles on EC2 to work without requiring AWS CLI profiles
	if profile != "default" {
		configOptions = append(configOptions, config.WithSharedConfigProfile(profile))
	}
	// For "default", let AWS SDK use its default credential chain:
	// 1. Environment variables
	// 2. IAM roles (EC2, ECS, Lambda)
	// 3. Shared credentials file
	
	// Override region if provided
	if len(region) > 0 && region[0] != "" {
		configOptions = append(configOptions, config.WithRegion(region[0]))
	}
	
	cfg, err := config.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration for profile '%s': %w", profile, err)
	}
	
	return cloudwatchlogs.NewFromConfig(cfg), nil
}

// getAWSRegion tries to get the region from AWS configuration
func getAWSRegion(profile string) (string, error) {
	ctx := context.Background()
	
	var configOptions []func(*config.LoadOptions) error
	if profile != "default" {
		configOptions = append(configOptions, config.WithSharedConfigProfile(profile))
	}
	
	cfg, err := config.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		return "", err
	}
	
	return cfg.Region, nil
}

// getEC2Region attempts to get the current EC2 instance's region from metadata
func getEC2Region() string {
	// Try EC2 instance metadata service (IMDSv2)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// First get the token for IMDSv2
	tokenReq, err := http.NewRequestWithContext(ctx, "PUT", "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return ""
	}
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")
	
	client := &http.Client{Timeout: 2 * time.Second}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return ""
	}
	defer tokenResp.Body.Close()
	
	if tokenResp.StatusCode != 200 {
		return ""
	}
	
	tokenBytes, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return ""
	}
	token := string(tokenBytes)
	
	// Now get the region using the token
	regionReq, err := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/latest/meta-data/placement/region", nil)
	if err != nil {
		return ""
	}
	regionReq.Header.Set("X-aws-ec2-metadata-token", token)
	
	regionResp, err := client.Do(regionReq)
	if err != nil {
		return ""
	}
	defer regionResp.Body.Close()
	
	if regionResp.StatusCode != 200 {
		return ""
	}
	
	regionBytes, err := io.ReadAll(regionResp.Body)
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(regionBytes))
}

// selectAWSRegion lets user select an AWS region
func selectAWSRegion(uiConfig *UIConfig) (string, error) {
	// Common AWS regions
	regions := []string{
		"us-east-1",      // N. Virginia
		"us-east-2",      // Ohio
		"us-west-1",      // N. California
		"us-west-2",      // Oregon
		"eu-west-1",      // Ireland
		"eu-west-2",      // London
		"eu-west-3",      // Paris
		"eu-central-1",   // Frankfurt
		"eu-north-1",     // Stockholm
		"ap-southeast-1", // Singapore
		"ap-southeast-2", // Sydney
		"ap-northeast-1", // Tokyo
		"ap-northeast-2", // Seoul
		"ap-south-1",     // Mumbai
		"ca-central-1",   // Canada
		"sa-east-1",      // São Paulo
	}

	// Display region selection title
	printStyled("🌍 AWS Region Selection", "12", true)
	fmt.Println()

	var chosen string
	prompt := &survey.Select{
		Message:  "Select AWS region:",
		Options:  regions,
		PageSize: uiConfig.ProfilePageSize,
	}
	err := survey.AskOne(prompt, &chosen)
	if err != nil {
		return "", fmt.Errorf("failed to select region: %w", err)
	}

	return chosen, nil
}

// trimProfilePrefix removes "profile " prefix from AWS config section names
func trimProfilePrefix(name string) string {
	if strings.HasPrefix(name, "profile ") {
		return strings.TrimPrefix(name, "profile ")
	}
	return name
}
