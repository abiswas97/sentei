package tui

// Layout constants shared by all progress views.
const (
	// WindowCompletedTrail is how many recently completed steps stay visible
	// when a step list is windowed.
	WindowCompletedTrail = 2

	// WindowPendingLead is how many upcoming pending steps stay visible when
	// a step list is windowed.
	WindowPendingLead = 1

	// minProgressBarWidth is the narrowest the overall progress bar may
	// render; below this the fill stops being readable.
	minProgressBarWidth = 20

	// progressBarElapsedReserve is the line budget held back from the bar
	// for the elapsed readout ("  elapsed 9999s"). A fixed reserve keeps the
	// bar from shrinking a cell every time the elapsed digits grow.
	progressBarElapsedReserve = 15

	// progressBarPercentReserve is the budget for the static fallback bar's
	// trailing percentage ("  100%"); the animated bar renders its own
	// percentage inside its width.
	progressBarPercentReserve = 6
)

// overallBarWidth is the single sizing rule for the overall bar: fill the
// content width, hold back the elapsed reserve, never go below the floor.
func overallBarWidth(viewWidth int) int {
	return max(minProgressBarWidth, viewWidth-2-progressBarElapsedReserve)
}

// viewChromeRows is the vertical chrome budget subtracted from the terminal
// height when sizing scrollable view bodies: title block, separators, and
// the footer area.
const viewChromeRows = 6
