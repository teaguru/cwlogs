package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"gopkg.in/ini.v1"
)

// selectAWSProfile reads AWS config, extracts profiles, and lets user select one
func selectAWSProfile(uiConfig *UIConfig) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".aws", "config")
	cfg, err := ini.Load(path)
	if err != nil {
		return "", fmt.Errorf("failed to read AWS config file at %s: %w", path, err)
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
	printStyled("🔐 AWS Profile Selection", "12", true)
	fmt.Println()

	var chosen string
	prompt := &survey.Select{
		Message:  "Select AWS profile:",
		Options:  profiles,
		PageSize: uiConfig.ProfilePageSize,
	}
	err = survey.AskOne(prompt, &chosen)
	if err != nil {
		return "", fmt.Errorf("failed to select profile: %w", err)
	}

	return chosen, nil
}

// listLogGroups lists CloudWatch log groups for the selected profile
func listLogGroups(profile string) ([]string, error) {
	ctx := context.Background()

	// Load AWS config with the selected profile
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration for profile '%s': %w", profile, err)
	}

	// Create CloudWatch Logs client
	client := cloudwatchlogs.NewFromConfig(cfg)

	// List log groups
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

// createCloudWatchClient creates a CloudWatch Logs client for the given profile
func createCloudWatchClient(profile string) (*cloudwatchlogs.Client, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration for profile '%s': %w", profile, err)
	}
	return cloudwatchlogs.NewFromConfig(cfg), nil
}

// trimProfilePrefix removes "profile " prefix from AWS config section names
func trimProfilePrefix(name string) string {
	if strings.HasPrefix(name, "profile ") {
		return strings.TrimPrefix(name, "profile ")
	}
	return name
}
