package ui

import "github.com/charmbracelet/lipgloss"

var (
	appPadding = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#111827")).
			Background(lipgloss.Color("#A7F3D0")).
			Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1)

	focusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))

	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
)
