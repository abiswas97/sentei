package tui

import (
	"charm.land/lipgloss/v2"
)

// Palette: the single source for every color in the TUI. Styles below and
// the bordered overlays reference these tokens, never raw color values.
var (
	colorAccent    = lipgloss.Color("62")  // purple: active, accent, borders
	colorSuccess   = lipgloss.Color("42")  // green: clean, done, success
	colorWarning   = lipgloss.Color("214") // orange: dirty, warnings
	colorError     = lipgloss.Color("196") // red: untracked, failed, errors
	colorDim       = lipgloss.Color("241") // gray: secondary text, pending
	colorEmphasis  = lipgloss.Color("15")  // white: titles, cursor row
	colorBody      = lipgloss.Color("252") // light gray: normal rows
	colorSelected  = lipgloss.Color("212") // pink: selected rows
	colorProtected = lipgloss.Color("63")  // blue-purple: protected worktrees
	colorMuted     = lipgloss.Color("245") // mid gray: locked worktrees
)

// UI chrome
var (
	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(1, 0, 0, 0)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDim)
)

// Row styles
var (
	styleCursorRow = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorEmphasis)

	styleSelectedRow = lipgloss.NewStyle().
				Foreground(colorSelected)

	styleNormalRow = lipgloss.NewStyle().
			Foreground(colorBody)

	styleColumnHeader = lipgloss.NewStyle().
				Foreground(colorDim)

	styleColumnHeaderSorted = lipgloss.NewStyle().
				Foreground(colorEmphasis).
				Bold(true)
)

// Status indicators
var (
	styleStatusClean = lipgloss.NewStyle().
				Foreground(colorSuccess)

	styleStatusDirty = lipgloss.NewStyle().
				Foreground(colorWarning)

	styleStatusUntracked = lipgloss.NewStyle().
				Foreground(colorError)

	styleStatusLocked = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleStatusProtected = lipgloss.NewStyle().
				Foreground(colorProtected)
)

// Semantic styles
var (
	styleWarning = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorSuccess)

	styleError = lipgloss.NewStyle().
			Foreground(colorError)
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
			Foreground(colorSuccess).
			Bold(true)

	stylePhaseActive = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	stylePhasePending = lipgloss.NewStyle().
				Foreground(colorDim)
)

// Progress indicators
var (
	styleIndicatorDone = lipgloss.NewStyle().
				Foreground(colorSuccess)

	styleIndicatorActive = lipgloss.NewStyle().
				Foreground(colorAccent)

	styleIndicatorPending = lipgloss.NewStyle().
				Foreground(colorDim)

	styleIndicatorFailed = lipgloss.NewStyle().
				Foreground(colorError)

	styleIndicatorWarning = lipgloss.NewStyle().
				Foreground(colorWarning)
)

// Layout elements
var (
	styleSeparator = lipgloss.NewStyle().
			Foreground(colorDim)

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorEmphasis)

	styleAccent = lipgloss.NewStyle().
			Foreground(colorAccent)

	styleCheckboxOn = lipgloss.NewStyle().
			Foreground(colorSuccess)

	styleCheckboxOff = lipgloss.NewStyle().
				Foreground(colorDim)

	styleStagedAdd = lipgloss.NewStyle().
			Foreground(colorSuccess)

	styleStagedRemove = lipgloss.NewStyle().
				Foreground(colorWarning)
)

// Indicator characters
const (
	indicatorDone    = "●"
	indicatorActive  = "◐"
	indicatorPending = "·"
	indicatorFailed  = "✗"
	indicatorWarning = "⚠"
)
