package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/abiswas/wt-sweep/internal/git"
	"github.com/abiswas/wt-sweep/internal/worktree"
)

type stubRunner struct {
	responses map[string]stubResponse
}

type stubResponse struct {
	output string
	err    error
}

func (s *stubRunner) Run(dir string, args ...string) (string, error) {
	key := dir
	for _, a := range args {
		key += " " + a
	}
	resp, ok := s.responses[key]
	if !ok {
		return "", fmt.Errorf("unexpected command: %s", key)
	}
	return resp.output, resp.err
}

func TestUpdateProgress_AllDeletionsComplete_TriggersPrune(t *testing.T) {
	runner := &stubRunner{
		responses: map[string]stubResponse{
			"/repo worktree prune": {output: ""},
		},
	}

	m := NewModel([]git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/a"},
	}, runner, "/repo")
	m.view = progressView

	updated, cmd := m.updateProgress(allDeletionsCompleteMsg{})

	if cmd == nil {
		t.Fatal("expected a Cmd from allDeletionsCompleteMsg, got nil")
	}

	model := updated.(Model)
	if model.view == summaryView {
		t.Error("should not transition to summaryView yet — prune should run first")
	}

	msg := cmd()
	pruneMsg, ok := msg.(pruneCompleteMsg)
	if !ok {
		t.Fatalf("expected pruneCompleteMsg, got %T", msg)
	}
	if pruneMsg.Err != nil {
		t.Errorf("expected prune success, got error: %v", pruneMsg.Err)
	}
}

func TestUpdateProgress_PruneComplete_TransitionsToSummary(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView

	updated, _ := m.updateProgress(pruneCompleteMsg{Err: nil})
	model := updated.(Model)

	if model.view != summaryView {
		t.Errorf("expected summaryView, got %d", model.view)
	}
	if model.pruneErr == nil {
		t.Fatal("expected pruneErr to be set (non-nil pointer)")
	}
	if *model.pruneErr != nil {
		t.Errorf("expected nil prune error, got %v", *model.pruneErr)
	}
}

func TestUpdateProgress_PruneFailed_TransitionsToSummary(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView

	pruneError := fmt.Errorf("permission denied")
	updated, _ := m.updateProgress(pruneCompleteMsg{Err: pruneError})
	model := updated.(Model)

	if model.view != summaryView {
		t.Errorf("expected summaryView, got %d", model.view)
	}
	if model.pruneErr == nil {
		t.Fatal("expected pruneErr to be set")
	}
	if *model.pruneErr == nil {
		t.Error("expected non-nil prune error")
	}
}

func TestViewSummary_PruneSuccess(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = summaryView
	m.deletionResult = worktree.DeletionResult{
		SuccessCount: 2,
		FailureCount: 0,
		Outcomes: []worktree.WorktreeOutcome{
			{Path: "/a", Success: true},
			{Path: "/b", Success: true},
		},
	}
	noErr := error(nil)
	m.pruneErr = &noErr

	output := stripAnsi(m.viewSummary())

	if !strings.Contains(output, "Pruned orphaned worktree metadata") {
		t.Error("should show prune success message")
	}
	if strings.Contains(output, "Warning") {
		t.Error("should not show warning on prune success")
	}
}

func TestViewSummary_PruneFailure(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = summaryView
	m.deletionResult = worktree.DeletionResult{
		SuccessCount: 1,
		FailureCount: 0,
		Outcomes: []worktree.WorktreeOutcome{
			{Path: "/a", Success: true},
		},
	}
	pruneError := fmt.Errorf("permission denied")
	m.pruneErr = &pruneError

	output := stripAnsi(m.viewSummary())

	if !strings.Contains(output, "Warning: failed to prune worktree metadata") {
		t.Error("should show prune failure warning")
	}
	if !strings.Contains(output, "permission denied") {
		t.Error("should include the prune error message")
	}
}
