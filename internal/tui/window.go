package tui

import "github.com/abiswas97/sentei/internal/progress"

// WindowStats summarizes a windowed step list for the stat line.
type WindowStats struct {
	Done    int
	Active  int
	Pending int
	Failed  int
	Showing int
	Total   int
}

// WindowResult is the outcome of windowing a step list.
type WindowResult struct {
	Steps    []progress.StepState
	Windowed bool
	Stats    WindowStats
}

// WindowSteps selects which steps to display when the list may exceed the
// available terminal lines. Priority: failed and active steps are always
// visible; remaining room shows the most recently completed steps (up to
// WindowCompletedTrail) and the next pending steps (up to WindowPendingLead).
// One line is reserved for the stat line when windowing engages.
func WindowSteps(steps []progress.StepState, availableLines int) WindowResult {
	stats := WindowStats{Total: len(steps)}
	for _, s := range steps {
		switch s.Status {
		case progress.StepDone, progress.StepSkipped:
			stats.Done++
		case progress.StepRunning:
			stats.Active++
		case progress.StepFailed:
			stats.Failed++
		default:
			stats.Pending++
		}
	}

	if len(steps) <= availableLines {
		stats.Showing = len(steps)
		return WindowResult{Steps: steps, Windowed: false, Stats: stats}
	}

	visible := make([]bool, len(steps))
	for i, s := range steps {
		if s.Status == progress.StepFailed || s.Status == progress.StepRunning {
			visible[i] = true
		}
	}

	budget := max(availableLines-1, 0) // reserve the stat line
	remaining := budget - stats.Failed - stats.Active

	completedTrail := min(WindowCompletedTrail, max(remaining, 0))
	for i := len(steps) - 1; i >= 0 && completedTrail > 0; i-- {
		if steps[i].Status == progress.StepDone || steps[i].Status == progress.StepSkipped {
			visible[i] = true
			completedTrail--
			remaining--
		}
	}

	pendingLead := min(WindowPendingLead, max(remaining, 0))
	for i := 0; i < len(steps) && pendingLead > 0; i++ {
		if steps[i].Status == progress.StepPending {
			visible[i] = true
			pendingLead--
		}
	}

	var windowed []progress.StepState
	for i, s := range steps {
		if visible[i] {
			windowed = append(windowed, s)
		}
	}
	stats.Showing = len(windowed)
	return WindowResult{Steps: windowed, Windowed: true, Stats: stats}
}
