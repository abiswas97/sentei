package tui

import (
	"testing"

	"github.com/abiswas97/sentei/internal/pipeline"
)

func makeSteps(statuses ...pipeline.StepStatus) []stepDisplay {
	steps := make([]stepDisplay, len(statuses))
	for i, st := range statuses {
		steps[i] = stepDisplay{name: stepName(i), status: st}
	}
	return steps
}

func stepName(i int) string {
	return string(rune('a' + i%26))
}

func countByStatus(steps []stepDisplay, status pipeline.StepStatus) int {
	n := 0
	for _, s := range steps {
		if s.status == status {
			n++
		}
	}
	return n
}

func TestWindowSteps_AllFit(t *testing.T) {
	steps := makeSteps(pipeline.StepDone, pipeline.StepRunning, pipeline.StepPending, pipeline.StepPending, pipeline.StepPending)
	r := WindowSteps(steps, 10)
	if r.Windowed {
		t.Error("expected no windowing when all items fit")
	}
	if len(r.Steps) != 5 || r.Stats.Showing != 5 {
		t.Errorf("expected all 5 steps, got %d (showing %d)", len(r.Steps), r.Stats.Showing)
	}
}

func TestWindowSteps_ExceedsBudget(t *testing.T) {
	var statuses []pipeline.StepStatus
	for range 10 {
		statuses = append(statuses, pipeline.StepDone)
	}
	statuses = append(statuses, pipeline.StepRunning, pipeline.StepRunning, pipeline.StepRunning)
	for range 17 {
		statuses = append(statuses, pipeline.StepPending)
	}
	r := WindowSteps(makeSteps(statuses...), 8)

	if !r.Windowed {
		t.Fatal("expected windowing with 30 items in 8 lines")
	}
	if r.Stats.Showing != len(r.Steps) {
		t.Errorf("Showing %d != visible %d", r.Stats.Showing, len(r.Steps))
	}
	// 3 active + WindowCompletedTrail done + WindowPendingLead pending = 6,
	// within the 7-line budget after the reserved stat line.
	want := 3 + WindowCompletedTrail + WindowPendingLead
	if len(r.Steps) != want {
		t.Errorf("expected %d visible steps, got %d", want, len(r.Steps))
	}
	if got := r.Stats.Done; got != 10 {
		t.Errorf("stats.Done = %d, want 10", got)
	}
	if got := r.Stats.Pending; got != 17 {
		t.Errorf("stats.Pending = %d, want 17", got)
	}
}

func TestWindowSteps_FailedAlwaysVisible(t *testing.T) {
	var statuses []pipeline.StepStatus
	for range 27 {
		statuses = append(statuses, pipeline.StepDone)
	}
	statuses = append(statuses, pipeline.StepFailed, pipeline.StepFailed, pipeline.StepFailed)
	r := WindowSteps(makeSteps(statuses...), 5)

	if got := countByStatus(r.Steps, pipeline.StepFailed); got != 3 {
		t.Errorf("expected all 3 failed steps visible under budget pressure, got %d", got)
	}
}

func TestWindowSteps_ActiveAlwaysVisible(t *testing.T) {
	var statuses []pipeline.StepStatus
	for range 25 {
		statuses = append(statuses, pipeline.StepPending)
	}
	for range 5 {
		statuses = append(statuses, pipeline.StepRunning)
	}
	r := WindowSteps(makeSteps(statuses...), 6)

	if got := countByStatus(r.Steps, pipeline.StepRunning); got != 5 {
		t.Errorf("expected all 5 active steps visible, got %d", got)
	}
}

func TestWindowSteps_BudgetZero_MinimumViable(t *testing.T) {
	steps := makeSteps(pipeline.StepDone, pipeline.StepFailed, pipeline.StepRunning, pipeline.StepPending)
	r := WindowSteps(steps, 0)

	if !r.Windowed {
		t.Fatal("expected windowing at zero budget")
	}
	if len(r.Steps) != 2 {
		t.Fatalf("expected only failed+active at zero budget, got %d steps", len(r.Steps))
	}
	for _, s := range r.Steps {
		if s.status != pipeline.StepFailed && s.status != pipeline.StepRunning {
			t.Errorf("unexpected status %v in minimum viable display", s.status)
		}
	}
}

func TestWindowSteps_RecentCompletedShown(t *testing.T) {
	steps := []stepDisplay{
		{name: "old-done", status: pipeline.StepDone},
		{name: "mid-done", status: pipeline.StepDone},
		{name: "new-done", status: pipeline.StepDone},
		{name: "running", status: pipeline.StepRunning},
		{name: "next-pending", status: pipeline.StepPending},
		{name: "far-pending-1", status: pipeline.StepPending},
		{name: "far-pending-2", status: pipeline.StepPending},
		{name: "far-pending-3", status: pipeline.StepPending},
	}
	r := WindowSteps(steps, 5) // 8 items in 5 lines -> windowed

	names := make(map[string]bool)
	for _, s := range r.Steps {
		names[s.name] = true
	}
	for _, want := range []string{"running", "new-done", "mid-done", "next-pending"} {
		if !names[want] {
			t.Errorf("expected %q visible, got %v", want, names)
		}
	}
	if names["old-done"] {
		t.Error("oldest completed step should be windowed out before recent ones")
	}
}

func TestWindowSteps_ResponsiveAcrossHeights(t *testing.T) {
	var statuses []pipeline.StepStatus
	for range 12 {
		statuses = append(statuses, pipeline.StepDone)
	}
	statuses = append(statuses, pipeline.StepRunning, pipeline.StepRunning)
	for range 16 {
		statuses = append(statuses, pipeline.StepPending)
	}
	steps := makeSteps(statuses...) // 30 items

	cases := []struct {
		height       int
		wantWindowed bool
	}{
		{60, false},
		{30, false},
		{20, true},
		{15, true},
	}
	for _, tc := range cases {
		r := WindowSteps(steps, tc.height)
		if r.Windowed != tc.wantWindowed {
			t.Errorf("height %d: windowed = %v, want %v", tc.height, r.Windowed, tc.wantWindowed)
		}
		if tc.wantWindowed && r.Stats.Showing >= 30 {
			t.Errorf("height %d: showing %d, expected a strict subset", tc.height, r.Stats.Showing)
		}
	}
}
