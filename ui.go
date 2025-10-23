package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/lipgloss"
)

// printStyled is a helper function for styled output
func printStyled(text, color string, bold bool) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Padding(1, 0)
	if bold {
		style = style.Bold(true)
	}
	fmt.Println(style.Render(text))
}

// selectLogGroup displays log group selection and returns the chosen group
func selectLogGroup(logGroups []string, uiConfig *UIConfig) (string, error) {
	// Display log group selection title
	printStyled("📋 CloudWatch Log Group Selection", "12", true)
	fmt.Println()

	// Let user select a log group
	var chosenLogGroup string
	prompt := &survey.Select{
		Message:  "Select CloudWatch log group:",
		Options:  logGroups,
		PageSize: uiConfig.LogGroupPageSize,
	}
	err := survey.AskOne(prompt, &chosenLogGroup)
	return chosenLogGroup, err
}

// displayWelcome shows the application welcome message
func displayWelcome() {
	printStyled("☁️  AWS CloudWatch Logs Viewer", "10", true)
}

// displayConnectionSuccess shows successful connection message and controls
func displayConnectionSuccess(profile, logGroup string) {
	printStyled(fmt.Sprintf("✅ Connected to: %s → %s", profile, logGroup), "10", false)
	fmt.Println()
	printStyled("Controls: ↑↓/j/k=scroll, PgUp/PgDn=fast scroll, g=top, G/End=latest\n"+
		"          /=search, Esc=clear search, n/N=next/prev match\n"+
		"          J=format toggle, F=follow toggle, H=load history, b=back, q=quit\n\n"+
		"💡 Auto-follow turns OFF when you scroll up, ON when you reach bottom\n"+
		"💡 Use mouse to select text for copy/paste (Ctrl+C in most terminals)", "15", false)
}
