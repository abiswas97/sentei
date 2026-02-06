package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	styleCursorRow = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	styleSelectedRow = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212"))

	styleNormalRow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	styleStatusClean = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))

	styleStatusDirty = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	styleStatusUntracked = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	styleStatusLocked = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	styleStatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0, 0, 0)

	styleDialogBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	styleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Fixed column widths for the worktree list table.
	// Cursor/checkbox/status include +1 char for inter-column gap (no Padding).
	// Age uses Padding(0,1) in StyleFunc instead, so no extra char here.
	colWidthCursor   = 3  // "> " (2) + 1 gap
	colWidthCheckbox = 5  // "[x]" (3) + 2 gap
	colWidthStatus   = 6  // "[ok]" (4) + 2 gap
	colWidthAge      = 16 // "12 hours ago" (12) + headroom
)
