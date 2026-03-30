package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
)

func makeRemoveModel(worktrees []git.Worktree) Model {
	m := NewMenuModel(nil, nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.remove.worktrees = worktrees
	m.reindex()
	m.width = 80
	m.height = 24
	return m
}

func sampleWorktrees() []git.Worktree {
	return []git.Worktree{
		{
			Path:           "/repo/wt-alpha",
			Branch:         "refs/heads/feature/alpha",
			LastCommitDate: time.Now().Add(-48 * time.Hour),
		},
		{
			Path:           "/repo/wt-beta",
			Branch:         "refs/heads/feature/beta",
			LastCommitDate: time.Now().Add(-24 * time.Hour),
		},
		{
			Path:           "/repo/wt-main",
			Branch:         "refs/heads/main",
			LastCommitDate: time.Now(),
		},
	}
}

func TestSetRemoveOpts_PreSelectsMatchingWorktrees(t *testing.T) {
	wts := sampleWorktrees()
	m := makeRemoveModel(wts)

	m.SetRemoveOpts(RemovePreSelection{
		Paths:       []string{"/repo/wt-alpha", "/repo/wt-beta"},
		FilterLabel: "merged",
	})

	if m.view != listView {
		t.Errorf("expected view=listView, got %d", m.view)
	}
	if !m.remove.selected["/repo/wt-alpha"] {
		t.Error("expected wt-alpha to be pre-selected")
	}
	if !m.remove.selected["/repo/wt-beta"] {
		t.Error("expected wt-beta to be pre-selected")
	}
	if m.remove.selected["/repo/wt-main"] {
		t.Error("expected wt-main to NOT be pre-selected")
	}
	if len(m.remove.selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(m.remove.selected))
	}
}

func TestSetRemoveOpts_IgnoresPathsNotInWorktrees(t *testing.T) {
	wts := sampleWorktrees()
	m := makeRemoveModel(wts)

	m.SetRemoveOpts(RemovePreSelection{
		Paths:       []string{"/repo/wt-alpha", "/repo/nonexistent"},
		FilterLabel: "stale > 30d",
	})

	if !m.remove.selected["/repo/wt-alpha"] {
		t.Error("expected wt-alpha to be pre-selected")
	}
	if m.remove.selected["/repo/nonexistent"] {
		t.Error("expected nonexistent path to NOT be selected")
	}
	if len(m.remove.selected) != 1 {
		t.Errorf("expected 1 selected, got %d", len(m.remove.selected))
	}
}

func TestSetRemoveOpts_SetsFilterLabel(t *testing.T) {
	wts := sampleWorktrees()
	m := makeRemoveModel(wts)

	m.SetRemoveOpts(RemovePreSelection{
		Paths:       []string{"/repo/wt-alpha"},
		FilterLabel: "stale > 30d",
	})

	if m.remove.filterLabel != "stale > 30d" {
		t.Errorf("expected filterLabel='stale > 30d', got %q", m.remove.filterLabel)
	}
}

func TestSetRemoveOpts_EmptyPaths(t *testing.T) {
	wts := sampleWorktrees()
	m := makeRemoveModel(wts)

	m.SetRemoveOpts(RemovePreSelection{
		Paths:       nil,
		FilterLabel: "merged",
	})

	if len(m.remove.selected) != 0 {
		t.Errorf("expected 0 selected, got %d", len(m.remove.selected))
	}
	if m.view != listView {
		t.Errorf("expected view=listView, got %d", m.view)
	}
}

func TestFilterIndicator_AppearsInStatusBar(t *testing.T) {
	wts := sampleWorktrees()
	m := makeRemoveModel(wts)

	m.SetRemoveOpts(RemovePreSelection{
		Paths:       []string{"/repo/wt-alpha"},
		FilterLabel: "merged",
	})

	output := stripAnsi(m.viewStatusBar())
	if !strings.Contains(output, "pre-filter: merged") {
		t.Errorf("expected 'pre-filter: merged' in status bar, got:\n%s", output)
	}
}

func TestFilterIndicator_AbsentWhenNoFilterLabel(t *testing.T) {
	wts := sampleWorktrees()
	m := makeRemoveModel(wts)

	// Enter list view without filter.
	m.view = listView

	output := stripAnsi(m.viewStatusBar())
	if strings.Contains(output, "pre-filter:") {
		t.Errorf("expected no 'pre-filter:' in status bar, got:\n%s", output)
	}
}

func TestFilterIndicator_InFullListView(t *testing.T) {
	wts := sampleWorktrees()
	m := makeRemoveModel(wts)

	m.SetRemoveOpts(RemovePreSelection{
		Paths:       []string{"/repo/wt-alpha"},
		FilterLabel: "stale > 7d",
	})

	output := stripAnsi(m.viewList())
	if !strings.Contains(output, "pre-filter: stale > 7d") {
		t.Errorf("expected 'pre-filter: stale > 7d' in list view, got:\n%s", output)
	}
}
