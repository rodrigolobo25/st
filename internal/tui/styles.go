package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Green   = lipgloss.Color("#00ff00")
	Yellow  = lipgloss.Color("#ffff00")
	Cyan    = lipgloss.Color("#00ffff")
	Red     = lipgloss.Color("#ff5555")
	Magenta = lipgloss.Color("#ff79c6")
	Dim     = lipgloss.Color("#666666")
	White   = lipgloss.Color("#ffffff")

	// Branch styles
	CurrentBranchStyle = lipgloss.NewStyle().Bold(true).Foreground(Green)
	BranchStyle        = lipgloss.NewStyle().Foreground(White)
	TrunkStyle         = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	DimStyle           = lipgloss.NewStyle().Foreground(Dim)
	WarningStyle       = lipgloss.NewStyle().Foreground(Yellow)
	ErrorStyle         = lipgloss.NewStyle().Foreground(Red)
	SuccessStyle       = lipgloss.NewStyle().Foreground(Green)
	InfoStyle          = lipgloss.NewStyle().Foreground(Cyan)

	// TUI panel styles
	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Cyan)

	InactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Dim)

	SelectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Green)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(White)

	SearchStyle = lipgloss.NewStyle().
			Foreground(Yellow).
			Italic(true)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Cyan).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(Dim)

	HereMarker = lipgloss.NewStyle().Foreground(Green).Bold(true).Render(" ← you are here")
)
