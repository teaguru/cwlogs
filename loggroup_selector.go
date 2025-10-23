package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// logGroupSelectorModel represents the log group selection TUI
type logGroupSelectorModel struct {
	logGroups    []string
	cursor       int
	selected     string
	width        int
	height       int
	changeRegion bool
	quit         bool
	config       *UIConfig
}

// newLogGroupSelector creates a new log group selector
func newLogGroupSelector(logGroups []string, config *UIConfig) *logGroupSelectorModel {
	return &logGroupSelectorModel{
		logGroups: logGroups,
		cursor:    0,
		config:    config,
	}
}

// Init initializes the log group selector
func (m *logGroupSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the log group selector
func (m *logGroupSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "r":
			m.changeRegion = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.logGroups)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = m.logGroups[m.cursor]
			return m, tea.Quit

		case "esc":
			m.quit = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the log group selector
func (m *logGroupSelectorModel) View() string {
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("📋 CloudWatch Log Group Selection")
	
	b.WriteString(title)
	b.WriteString("\n\n")

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Use ↑↓ or j/k to navigate, Enter to select, r to change region, q to quit")
	
	b.WriteString(instructions)
	b.WriteString("\n\n")

	// Log groups list
	maxVisible := m.height - 8 // Reserve space for title, instructions, and controls
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	end := len(m.logGroups)

	// Calculate visible window
	if len(m.logGroups) > maxVisible {
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.logGroups) {
			end = len(m.logGroups)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	// Render visible log groups
	for i := start; i < end; i++ {
		logGroup := m.logGroups[i]
		
		// Truncate long log group names
		maxWidth := m.width - 4
		if maxWidth < 20 {
			maxWidth = 20
		}
		if len(logGroup) > maxWidth {
			logGroup = logGroup[:maxWidth-3] + "..."
		}

		if i == m.cursor {
			// Highlight selected item
			line := lipgloss.NewStyle().
				Background(lipgloss.Color("12")).
				Foreground(lipgloss.Color("0")).
				Render(fmt.Sprintf("> %s", logGroup))
			b.WriteString(line)
		} else {
			b.WriteString(fmt.Sprintf("  %s", logGroup))
		}
		b.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(m.logGroups) > maxVisible {
		scrollInfo := fmt.Sprintf("\n[%d/%d log groups]", m.cursor+1, len(m.logGroups))
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(scrollInfo))
	}

	// Controls
	b.WriteString("\n\n")
	controls := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("↑↓/j/k: navigate | Enter: select | r: change region | q: quit")
	b.WriteString(controls)

	return b.String()
}

// selectLogGroupInteractive shows an interactive log group selector
func selectLogGroupInteractive(logGroups []string, config *UIConfig) (string, bool, error) {
	model := newLogGroupSelector(logGroups, config)
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}

	if m, ok := finalModel.(*logGroupSelectorModel); ok {
		if m.changeRegion {
			return "", true, nil // Return true for region change
		}
		if m.quit || m.selected == "" {
			return "", false, fmt.Errorf("selection cancelled")
		}
		return m.selected, false, nil
	}

	return "", false, fmt.Errorf("unexpected model type")
}
