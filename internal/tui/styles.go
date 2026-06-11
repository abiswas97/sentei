package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// palette is the full set of color tokens the TUI draws from. Themes are
// values of this struct: behavior never branches on darkness, it only reads
// tokens, so adding a theme means adding a table below and nothing else.
type palette struct {
	accent    color.Color // active, accent, borders
	success   color.Color // clean, done, success
	warning   color.Color // dirty, warnings
	errorc    color.Color // untracked, failed, errors
	dim       color.Color // secondary text, pending
	emphasis  color.Color // titles, cursor row
	body      color.Color // normal rows
	selected  color.Color // selected rows
	protected color.Color // protected worktrees
	muted     color.Color // locked worktrees
	barStart  color.Color // overall bar gradient, trailing edge
	barEnd    color.Color // overall bar gradient, leading edge
}

// The two palettes, selected by terminal background detection (model.go).
// Dark is the default and the documented baseline in .impeccable.md.
var (
	darkPalette = palette{
		accent:    lipgloss.Color("62"),
		success:   lipgloss.Color("42"),
		warning:   lipgloss.Color("214"),
		errorc:    lipgloss.Color("196"),
		dim:       lipgloss.Color("241"),
		emphasis:  lipgloss.Color("15"),
		body:      lipgloss.Color("252"),
		selected:  lipgloss.Color("212"),
		protected: lipgloss.Color("63"),
		muted:     lipgloss.Color("245"),
		// Hex twins of accent (62) and selected (212): blends interpolate
		// in RGB, which indexed ANSI codes cannot do smoothly.
		barStart: lipgloss.Color("#5f5fd7"),
		barEnd:   lipgloss.Color("#ff87d7"),
	}

	lightPalette = palette{
		accent:    lipgloss.Color("56"),
		success:   lipgloss.Color("29"),
		warning:   lipgloss.Color("166"),
		errorc:    lipgloss.Color("160"),
		dim:       lipgloss.Color("245"),
		emphasis:  lipgloss.Color("235"),
		body:      lipgloss.Color("238"),
		selected:  lipgloss.Color("168"),
		protected: lipgloss.Color("26"),
		muted:     lipgloss.Color("243"),
		// Hex twins of accent (56) and selected (168).
		barStart: lipgloss.Color("#5f00d7"),
		barEnd:   lipgloss.Color("#d75f87"),
	}
)

// Active color tokens and the styles derived from them. All are assigned by
// applyPalette: nothing in this package constructs a color or style outside
// that one path.
var (
	colorAccent    color.Color
	colorSuccess   color.Color
	colorWarning   color.Color
	colorError     color.Color
	colorDim       color.Color
	colorEmphasis  color.Color
	colorBody      color.Color
	colorSelected  color.Color
	colorProtected color.Color
	colorMuted     color.Color
	colorBarStart  color.Color
	colorBarEnd    color.Color

	// UI chrome
	styleStatusBar lipgloss.Style
	styleDim       lipgloss.Style

	// Row styles
	styleCursorRow          lipgloss.Style
	styleSelectedRow        lipgloss.Style
	styleNormalRow          lipgloss.Style
	styleColumnHeader       lipgloss.Style
	styleColumnHeaderSorted lipgloss.Style

	// Status indicators
	styleStatusClean     lipgloss.Style
	styleStatusDirty     lipgloss.Style
	styleStatusUntracked lipgloss.Style
	styleStatusLocked    lipgloss.Style
	styleStatusProtected lipgloss.Style

	// Semantic styles
	styleWarning lipgloss.Style
	styleSuccess lipgloss.Style
	styleError   lipgloss.Style

	// Phase header styles
	stylePhaseDone    lipgloss.Style
	stylePhaseActive  lipgloss.Style
	stylePhasePending lipgloss.Style

	// Progress indicators
	styleIndicatorDone    lipgloss.Style
	styleIndicatorActive  lipgloss.Style
	styleIndicatorPending lipgloss.Style
	styleIndicatorFailed  lipgloss.Style
	styleIndicatorWarning lipgloss.Style

	// Layout elements
	styleSeparator    lipgloss.Style
	styleTitle        lipgloss.Style
	styleAccent       lipgloss.Style
	styleCheckboxOn   lipgloss.Style
	styleCheckboxOff  lipgloss.Style
	styleStagedAdd    lipgloss.Style
	styleStagedRemove lipgloss.Style

	// Bordered overlays
	stylePortalBox lipgloss.Style
)

func init() {
	applyPalette(darkPalette)
}

// applyPalette installs p as the active theme: tokens first, then every
// derived style. Called at init with the dark palette and again from Update
// when background detection reports a light terminal; the Elm loop is
// single-goroutine, so reassignment here is never concurrent with a render.
func applyPalette(p palette) {
	colorAccent = p.accent
	colorSuccess = p.success
	colorWarning = p.warning
	colorError = p.errorc
	colorDim = p.dim
	colorEmphasis = p.emphasis
	colorBody = p.body
	colorSelected = p.selected
	colorProtected = p.protected
	colorMuted = p.muted
	colorBarStart = p.barStart
	colorBarEnd = p.barEnd

	styleStatusBar = lipgloss.NewStyle().Foreground(colorDim).Padding(1, 0, 0, 0)
	styleDim = lipgloss.NewStyle().Foreground(colorDim)

	styleCursorRow = lipgloss.NewStyle().Bold(true).Foreground(colorEmphasis)
	styleSelectedRow = lipgloss.NewStyle().Foreground(colorSelected)
	styleNormalRow = lipgloss.NewStyle().Foreground(colorBody)
	styleColumnHeader = lipgloss.NewStyle().Foreground(colorDim)
	styleColumnHeaderSorted = lipgloss.NewStyle().Foreground(colorEmphasis).Bold(true)

	styleStatusClean = lipgloss.NewStyle().Foreground(colorSuccess)
	styleStatusDirty = lipgloss.NewStyle().Foreground(colorWarning)
	styleStatusUntracked = lipgloss.NewStyle().Foreground(colorError)
	styleStatusLocked = lipgloss.NewStyle().Foreground(colorMuted)
	styleStatusProtected = lipgloss.NewStyle().Foreground(colorProtected)

	styleWarning = lipgloss.NewStyle().Foreground(colorWarning).Bold(true)
	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess)
	styleError = lipgloss.NewStyle().Foreground(colorError)

	stylePhaseDone = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	stylePhaseActive = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	stylePhasePending = lipgloss.NewStyle().Foreground(colorDim)

	styleIndicatorDone = lipgloss.NewStyle().Foreground(colorSuccess)
	styleIndicatorActive = lipgloss.NewStyle().Foreground(colorAccent)
	styleIndicatorPending = lipgloss.NewStyle().Foreground(colorDim)
	styleIndicatorFailed = lipgloss.NewStyle().Foreground(colorError)
	styleIndicatorWarning = lipgloss.NewStyle().Foreground(colorWarning)

	styleSeparator = lipgloss.NewStyle().Foreground(colorDim)
	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(colorEmphasis)
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleCheckboxOn = lipgloss.NewStyle().Foreground(colorSuccess)
	styleCheckboxOff = lipgloss.NewStyle().Foreground(colorDim)
	styleStagedAdd = lipgloss.NewStyle().Foreground(colorSuccess)
	styleStagedRemove = lipgloss.NewStyle().Foreground(colorWarning)

	stylePortalBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent).Padding(0, 1)

	footerHelp.Styles.ShortKey = styleDim
	footerHelp.Styles.ShortDesc = styleDim
	footerHelp.Styles.ShortSeparator = styleDim
	footerHelp.Styles.Ellipsis = styleDim
}

// Fixed column widths for the worktree list table.
// Cursor/checkbox/status include +1 char for inter-column gap (no Padding).
// Age uses Padding(0,1) in StyleFunc instead, so no extra char here.
var (
	colWidthCursor   = 3  // "> " (2) + 1 gap
	colWidthCheckbox = 5  // "[x]" (3) + 2 gap
	colWidthStatus   = 6  // "[ok]" (4) + 2 gap
	colWidthAge      = 16 // "12 hours ago" (12) + headroom
)

// Indicator characters. ✓/✗ are verdicts about a whole operation (summary
// headlines); ●/·/the work animation are states of items within it, always
// rendered among peers. ● never appears alone on a screen.
const (
	indicatorSuccess = "✓"
	indicatorDone    = "●"
	indicatorPending = "·"
	indicatorFailed  = "✗"
	indicatorWarning = "⚠"

	// indicatorActiveFallback marks live work in pure layouts with no
	// animation frame injected.
	indicatorActiveFallback = "∙"
)

// workFrames is the one working animation: the breathing dot, the pending
// dot inflating toward the done dot and back. The dot family is the only
// spinner that is optically centered and weight-matched beside bold text;
// braille frames hang into the descender zone and sit flush-left, which is
// font geometry the app cannot fix.
var workFrames = []string{"·", "∙", "•", "●", "•", "∙"}
