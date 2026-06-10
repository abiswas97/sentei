package tui

// Layout constants shared by all progress views.
const (
	// WindowCompletedTrail is how many recently completed steps stay visible
	// when a step list is windowed.
	WindowCompletedTrail = 2

	// WindowPendingLead is how many upcoming pending steps stay visible when
	// a step list is windowed.
	WindowPendingLead = 1

	// progressBarWidth is the cell width of the overall progress bar.
	progressBarWidth = 20
)
