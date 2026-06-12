package tui

import (
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/progress"
)

func TestProgressLayout_WithSubtitle(t *testing.T) {
	l := ProgressLayout{
		Title:    "Creating worktree",
		Subtitle: "feature/foo → from main",
		Width:    80,
		Height:   30,
		Phases: []phaseDisplay{
			{name: "Setup", total: 2, done: 1, steps: []stepDisplay{
				{name: "Create worktree", status: progress.StepDone},
				{name: "Merge base branch", status: progress.StepRunning},
			}},
			{name: "Dependencies"},
			{name: "Integrations"},
		},
	}
	view := stripANSI(l.View())

	for _, want := range []string{"sentei ─ Creating worktree", "feature/foo → from main", "┄", "Setup", "Dependencies", "pending"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected layout to contain %q, view:\n%s", want, view)
		}
	}
}

func TestProgressLayout_WithoutSubtitle(t *testing.T) {
	l := ProgressLayout{Title: "Removing worktrees", Width: 80, Height: 30}
	view := stripANSI(l.View())

	lines := strings.Split(view, "\n")
	if !strings.Contains(lines[0], "sentei ─ Removing worktrees") {
		t.Errorf("expected title first, got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "" {
		t.Errorf("expected blank line directly after title when no subtitle, got %q", lines[1])
	}
}

func TestProgressLayout_CompletedPhaseCollapses(t *testing.T) {
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Phases: []phaseDisplay{
			{name: "Setup", total: 2, done: 2, steps: []stepDisplay{
				{name: "Create worktree", status: progress.StepDone},
				{name: "Merge base branch", status: progress.StepDone},
			}},
		},
	}
	view := stripANSI(l.View())

	if !strings.Contains(view, "✦ Setup  2/2  100%") {
		t.Errorf("expected collapsed completed phase, view:\n%s", view)
	}
	if strings.Contains(view, "Create worktree") {
		t.Errorf("completed phase must not list steps, view:\n%s", view)
	}
}

func TestProgressLayout_CompletedPhaseWithFailuresKeepsSteps(t *testing.T) {
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Phases: []phaseDisplay{
			{name: "Setup", total: 2, done: 2, failed: 1, steps: []stepDisplay{
				{name: "Create worktree", status: progress.StepDone},
				{name: "Install hooks", status: progress.StepFailed},
			}},
		},
	}
	view := stripANSI(l.View())

	if !strings.Contains(view, "Install hooks") {
		t.Errorf("failed steps must stay visible after phase completion, view:\n%s", view)
	}
	if !strings.Contains(view, indicatorFailed) {
		t.Errorf("expected failed indicator, view:\n%s", view)
	}
}

func TestProgressLayout_ActivePhaseShowsSteps(t *testing.T) {
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Phases: []phaseDisplay{
			{name: "Removing worktrees", total: 30, done: 12, steps: []stepDisplay{
				{name: "done-step", status: progress.StepDone},
				{name: "active-step", status: progress.StepRunning},
				{name: "pending-step", status: progress.StepPending},
			}},
		},
	}
	view := stripANSI(l.View())

	if !strings.Contains(view, "✻ Removing worktrees  12/30  40%") {
		t.Errorf("expected active phase header with indicator left and count/pct, view:\n%s", view)
	}
	for _, want := range []string{"    ✦ done-step", "    ✻ active-step", "    · pending-step"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected 4-space indented step %q, view:\n%s", want, view)
		}
	}
}

func TestProgressLayout_ZeroTotalPhaseIsPending(t *testing.T) {
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Phases: []phaseDisplay{
			{name: "Integrations", total: 0, done: 0, steps: []stepDisplay{
				{name: "queued work", status: progress.StepPending},
			}},
		},
	}
	view := stripANSI(l.View())

	if strings.Contains(view, "100%") {
		t.Errorf("a 0/0 phase must never render complete, view:\n%s", view)
	}
	if !strings.Contains(view, "pending") {
		t.Errorf("expected pending phase, view:\n%s", view)
	}
}

func TestProgressLayout_OverallBarAggregatesPhases(t *testing.T) {
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Phases: []phaseDisplay{
			{name: "A", total: 12, done: 8},
			{name: "B", total: 8, done: 2},
		},
	}
	view := stripANSI(l.View())

	barWidth := overallBarWidth(80) - progressBarPercentReserve
	want := strings.Repeat("█", barWidth/2) + strings.Repeat("░", barWidth-barWidth/2) + " 50%"
	if !strings.Contains(view, want) {
		t.Errorf("expected aggregated 50%% bar, view:\n%s", view)
	}
}

func TestProgressLayout_DoneExceedsTotal_NoPanic(t *testing.T) {
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Phases: []phaseDisplay{
			{name: "A", total: 1, done: 3},
		},
	}
	view := stripANSI(l.View()) // must not panic

	if !strings.Contains(view, "100%") {
		t.Errorf("expected clamped 100%%, view:\n%s", view)
	}
	if strings.Contains(view, "300%") {
		t.Errorf("percentage must clamp, view:\n%s", view)
	}
}

func TestProgressLayout_WindowsStepsOnShortTerminal(t *testing.T) {
	steps := make([]stepDisplay, 30)
	for i := range steps {
		steps[i] = stepDisplay{name: stepName(i), status: progress.StepPending}
	}
	steps[0].status = progress.StepRunning
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 15,
		Phases: []phaseDisplay{{name: "Removing", total: 30, done: 0, steps: steps}},
	}
	view := stripANSI(l.View())

	if !strings.Contains(view, "showing") {
		t.Errorf("expected stat line when windowed at height 15, view:\n%s", view)
	}
}

func TestProgressLayout_HintsRendered(t *testing.T) {
	l := ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Hints: progressFooter,
	}
	view := stripANSI(l.View())

	if !strings.Contains(view, "  q quit") {
		t.Errorf("expected key hints, view:\n%s", view)
	}
}
