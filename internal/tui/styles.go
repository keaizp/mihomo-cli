package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7c3aed")).
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#7c3aed")).
			Padding(0, 1)

	NormalStyle = lipgloss.NewStyle().
			Padding(0, 1)

	GroupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#a78bfa")).
				Padding(0, 1)

	InfoStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3f3f46"))

	LogStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717a"))

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a1a1aa")).
			Background(lipgloss.Color("#18181b")).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#52525b"))
)
