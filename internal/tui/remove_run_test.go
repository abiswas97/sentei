package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/worktree"
)

func keyRune(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

// pollutedRun returns run state as a completed previous deletion left it:
// two outcomes, done statuses, prune and cleanup results all set.
func pollutedRun() removalRun {
	pruneErr := error(nil)
	return removalRun{
		worktrees: []git.Worktree{
			{Path: "/work/a", Branch: "refs/heads/a"},
			{Path: "/work/b", Branch: "refs/heads/b"},
		},
		statuses: map[string]string{"/work/a": statusRemoved, "/work/b": statusRemoved},
		result: worktree.DeletionResult{
			SuccessCount: 2,
			Outcomes: []worktree.WorktreeOutcome{
				{Path: "/work/a", Success: true},
				{Path: "/work/b", Success: true},
			},
		},
		teardownResults: []progress.StepResult{{Name: "old", Status: progress.StepDone}},
		pruneErr:        &pruneErr,
		cleanupResult:   &cleanup.Result{},
	}
}

func TestConfirmYes_SecondRunStartsFresh(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/a"},
		{Path: "/work/b", Branch: "refs/heads/b"},
		{Path: "/work/c", Branch: "refs/heads/c"},
	}, nil, "/repo")
	m.remove.run = pollutedRun()
	m.remove.selected = map[string]bool{"/work/c": true}
	m.view = confirmView

	updated, _ := m.updateConfirm(keyRune('y'))
	model := updated.(Model)

	run := model.remove.run
	if got := len(run.result.Outcomes); got != 0 {
		t.Errorf("expected 0 outcomes in fresh run, got %d", got)
	}
	if run.result.SuccessCount != 0 || run.result.FailureCount != 0 {
		t.Errorf("expected zero counts in fresh run, got %+v", run.result)
	}
	if got := run.total(); got != 1 {
		t.Errorf("expected run total 1, got %d", got)
	}
	if got := len(run.statuses); got != 1 {
		t.Errorf("expected 1 seeded status, got %d: %v", got, run.statuses)
	}
	if run.statuses["/work/c"] != statusPending {
		t.Errorf("expected /work/c pending, got %q", run.statuses["/work/c"])
	}
	if run.pruneErr != nil {
		t.Error("expected pruneErr cleared in fresh run")
	}
	if run.cleanupResult != nil {
		t.Error("expected cleanupResult cleared in fresh run")
	}
	if len(run.teardownResults) != 0 {
		t.Error("expected teardownResults cleared in fresh run")
	}

	view := model.viewProgress()
	if strings.Contains(view, "200%") {
		t.Error("progress view shows 200% from stale outcomes")
	}
	if !strings.Contains(view, "0%") {
		t.Errorf("expected fresh run to start at 0%%, view:\n%s", view)
	}
	if !strings.Contains(view, "pending") {
		t.Errorf("expected prune phase pending on fresh run, view:\n%s", view)
	}
}

func TestViewProgress_SecondRunCompletesAtHundredPercent(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/c", Branch: "refs/heads/c"},
	}, nil, "/repo")
	m.remove.run = newRemovalRun([]git.Worktree{{Path: "/work/c", Branch: "refs/heads/c"}})
	m.view = progressView

	updated, _ := m.updateProgress(worktreeDeletedMsg{Path: "/work/c"})
	model := updated.(Model)

	view := model.viewProgress()
	if !strings.Contains(view, "100%") {
		t.Errorf("expected 100%% after the run's single deletion, view:\n%s", view)
	}
}

func TestUpdateProgress_CleanupComplete_ClearsSelection(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/a"},
	}, nil, "/repo")
	m.remove.selected = map[string]bool{"/work/a": true}
	m.view = progressView

	updated, _ := m.updateProgress(cleanupCompleteMsg{})
	model := updated.(Model)

	if got := len(model.remove.selected); got != 0 {
		t.Errorf("expected selection cleared after completed run, got %d selected", got)
	}
}

func TestMenuEntry_RemoveWorktrees_ClearsSelection(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/a"},
	}, nil, "/repo")
	m.menuItems = []menuItem{{label: "Remove worktrees", enabled: true}}
	m.menuCursor = 0
	m.remove.selected = map[string]bool{"/work/a": true}
	m.view = menuView

	updated, _ := m.updateMenu(tea.KeyPressMsg{Code: tea.KeyEnter})
	model := updated.(Model)

	if model.view != listView {
		t.Fatalf("expected listView, got %d", model.view)
	}
	if got := len(model.remove.selected); got != 0 {
		t.Errorf("expected selection cleared on menu entry, got %d selected", got)
	}
}

func TestViewProgress_TeardownRunning_ShowsActivePhase(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/a"},
	}, nil, "/repo")
	m.remove.run = newRemovalRun([]git.Worktree{{Path: "/work/a", Branch: "refs/heads/a"}})
	m.remove.run.teardownRunning = true
	m.view = progressView

	view := stripAnsi(m.viewProgress())
	if !strings.Contains(view, "Teardown") {
		t.Errorf("expected Teardown phase visible while running, view:\n%s", view)
	}
	if !strings.Contains(view, starFrames[0]+" Teardown") {
		t.Errorf("expected star frame on running teardown, view:\n%s", view)
	}
}

func TestViewSummary_FailedOutcome_ReadsFromRun(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.remove.run = newRemovalRun(nil)
	m.remove.run.result = worktree.DeletionResult{
		SuccessCount: 1,
		FailureCount: 1,
		Outcomes: []worktree.WorktreeOutcome{
			{Path: "/work/a", Success: true},
			{Path: "/work/b", Success: false, Error: errors.New("boom")},
		},
	}
	m.view = summaryView

	view := m.viewSummary()
	if !strings.Contains(view, "1 removed") || !strings.Contains(view, "1 failed") {
		t.Errorf("expected counts from run result, view:\n%s", view)
	}
	if !strings.Contains(view, "boom") {
		t.Errorf("expected failure detail from run result, view:\n%s", view)
	}
}

// Regression: teardownCompleteMsg arrives while the view is progressView
// (Yes switches views before teardown runs), so updateProgress must handle
// it — updateConfirm never sees it. Dropping it left removal hanging at
// "Teardown 0/1" forever.
func TestUpdateProgress_TeardownComplete_StartsDeletions(t *testing.T) {
	wts := []git.Worktree{{Path: "/work/a", Branch: "refs/heads/a"}}
	m := NewModel(wts, nil, "/repo")
	m.remove.run = newRemovalRun(wts)
	m.remove.run.teardownRunning = true
	m.view = progressView

	updated, cmd := m.updateProgress(teardownCompleteMsg{
		results: []progress.StepResult{{Name: "Teardown code-review-graph", Status: progress.StepDone}},
	})
	model := updated.(Model)

	if model.remove.run.teardownRunning {
		t.Error("teardown must be marked finished")
	}
	if len(model.remove.run.teardownResults) != 1 {
		t.Errorf("expected teardown results stored, got %d", len(model.remove.run.teardownResults))
	}
	if model.remove.run.progressCh == nil {
		t.Error("expected deletion channel to be created")
	}
	if cmd == nil {
		t.Fatal("expected a Cmd that consumes deletion events")
	}
}
