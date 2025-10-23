package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// logGroupSelectorModel represents the log group selection TUI
type logGroupSelectorModel struct {
	logGroups       []string
	filteredGroups  []string
	cursor          int
	selected        string
	width           int
	height          int
	changeRegion    bool
	quit            bool
	config          *UIConfig
	searchQuery     string
}

// newLogGroupSelector creates a new log group selector
func newLogGroupSelector(logGroups []string, config *UIConfig) *logGroupSelectorModel {
	return &logGroupSelectorModel{
		logGroups:      logGroups,
		filteredGroups: logGroups, // Initially show all groups
		cursor:         0,
		config:         config,
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

		case "R":
			m.changeRegion = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filteredGroups)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.filteredGroups) > 0 {
				m.selected = m.filteredGroups[m.cursor]
				return m, tea.Quit
			}

		case "esc":
			if m.searchQuery != "" {
				// Clear search
				m.searchQuery = ""
				m.filteredGroups = m.logGroups
				m.cursor = 0
			} else {
				m.quit = true
				return m, tea.Quit
			}

		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterLogGroups()
			}

		default:
			// Auto-start search when typing alphanumeric characters
			if len(msg.Runes) > 0 {
				char := msg.Runes[0]
				// Check if it's a printable character (letters, numbers, common symbols)
				if (char >= 'a' && char <= 'z') || 
				   (char >= 'A' && char <= 'Z') || 
				   (char >= '0' && char <= '9') || 
				   char == '-' || char == '_' || char == '/' || char == '.' {
					m.searchQuery += string(msg.Runes)
					m.filterLogGroups()
				}
			}
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

	// Instructions and search status
	var instructions string
	if m.searchQuery != "" {
		instructions = fmt.Sprintf("Filter: %s_ | Esc to clear, Enter to select", m.searchQuery)
		instructions = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(instructions)
	} else {
		instructions = "Type to filter, ↑↓/j/k to navigate, Enter to select, R to change region, q to quit"
		instructions = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(instructions)
	}
	
	b.WriteString(instructions)
	b.WriteString("\n\n")

	// Log groups list
	maxVisible := m.height - 8 // Reserve space for title, instructions, and controls
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	end := len(m.filteredGroups)

	// Calculate visible window
	if len(m.filteredGroups) > maxVisible {
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.filteredGroups) {
			end = len(m.filteredGroups)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	// Show "no results" message if filtered list is empty
	if len(m.filteredGroups) == 0 {
		noResults := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render("No log groups match your search")
		b.WriteString(noResults)
		b.WriteString("\n")
	} else {
		// Render visible log groups
		for i := start; i < end; i++ {
			logGroup := m.filteredGroups[i]
		
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
	}

	// Show scroll indicator if needed
	if len(m.filteredGroups) > maxVisible {
		scrollInfo := fmt.Sprintf("\n[%d/%d log groups", m.cursor+1, len(m.filteredGroups))
		if m.searchQuery != "" {
			scrollInfo += fmt.Sprintf(" of %d total]", len(m.logGroups))
		} else {
			scrollInfo += "]"
		}
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(scrollInfo))
	}

	// Controls
	b.WriteString("\n\n")
	var controls string
	if m.searchQuery != "" {
		controls = "Type to filter | Backspace: delete | Esc: clear | Enter: select | q: quit"
	} else {
		controls = "Type to filter | ↑↓/j/k: navigate | Enter: select | R: change region | q: quit"
	}
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(controls))

	return b.String()
}

// filterLogGroups filters the log groups based on the search query
func (m *logGroupSelectorModel) filterLogGroups() {
	if m.searchQuery == "" {
		m.filteredGroups = m.logGroups
		m.cursor = 0
		return
	}

	m.filteredGroups = []string{}
	query := strings.ToLower(m.searchQuery)
	
	for _, group := range m.logGroups {
		if strings.Contains(strings.ToLower(group), query) {
			m.filteredGroups = append(m.filteredGroups, group)
		}
	}
	
	// Reset cursor to top of filtered results
	m.cursor = 0
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
