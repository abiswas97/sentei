package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// UI chrome
var (
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0, 0, 0)

	styleDialogBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// Row styles
var (
	styleCursorRow = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	styleSelectedRow = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212"))

	styleNormalRow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	styleColumnHeader = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	styleColumnHeaderSorted = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Bold(true)
)

// Status indicators
var (
	styleStatusClean = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))

	styleStatusDirty = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	styleStatusUntracked = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	styleStatusLocked = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	styleStatusProtected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("63"))
)

// Semantic styles
var (
	styleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// Fixed column widths for the worktree list table.
// Cursor/checkbox/status include +1 char for inter-column gap (no Padding).
// Age uses Padding(0,1) in StyleFunc instead, so no extra char here.
var (
	colWidthCursor   = 3  // "> " (2) + 1 gap
	colWidthCheckbox = 5  // "[x]" (3) + 2 gap
	colWidthStatus   = 6  // "[ok]" (4) + 2 gap
	colWidthAge      = 16 // "12 hours ago" (12) + headroom
)

// Phase header styles
var (
	stylePhaseDone = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	stylePhaseActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("62")).
				Bold(true)

	stylePhasePending = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

// Progress indicators
var (
	styleIndicatorDone = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))

	styleIndicatorActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("62"))

	styleIndicatorPending = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	styleIndicatorFailed = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	styleIndicatorWarning = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))
)

// Layout elements
var (
	styleSeparator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	styleAccent = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62"))

	styleCheckboxOn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	styleCheckboxOff = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	styleStagedAdd = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")) // green — same as clean/success

	styleStagedRemove = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // orange — same as dirty/warning

	styleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// Indicator characters
const (
	indicatorDone    = "●"
	indicatorActive  = "◐"
	indicatorPending = "·"
	indicatorFailed  = "✗"
	indicatorWarning = "⚠"
)

// separator renders a dotted separator line at the given width.
func separator(width int) string {
	if width <= 4 {
		width = 40
	}
	line := strings.Repeat("┄", width-4)
	return styleSeparator.Render("  " + line)
}
