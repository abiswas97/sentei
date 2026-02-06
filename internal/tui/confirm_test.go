package tui

import (
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/git"
)

func TestViewConfirm_CleanWorktrees(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/feature-a"},
		{Path: "/work/b", Branch: "refs/heads/feature-b"},
	}, nil, "/repo")
	m.selected[0] = true
	m.selected[1] = true
	m.view = confirmView

	output := stripAnsi(m.viewConfirm())

	if !strings.Contains(output, "delete 2 worktree(s)") {
		t.Error("should mention count of worktrees")
	}
	if !strings.Contains(output, "feature-a") {
		t.Error("should list feature-a")
	}
	if !strings.Contains(output, "(clean)") {
		t.Error("should show (clean) label for clean worktrees")
	}
	if strings.Contains(output, "WARNING") {
		t.Error("should not show warnings for clean worktrees")
	}
}

func TestViewConfirm_DirtyWorktree(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/dirty", Branch: "refs/heads/dirty-branch", HasUncommittedChanges: true},
	}, nil, "/repo")
	m.selected[0] = true
	m.view = confirmView

	output := stripAnsi(m.viewConfirm())

	if !strings.Contains(output, "HAS UNCOMMITTED CHANGES") {
		t.Error("should warn about uncommitted changes")
	}
	if !strings.Contains(output, "WARNING") {
		t.Error("should show WARNING for dirty worktrees")
	}
}

func TestViewConfirm_LockedWorktree(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/locked", Branch: "refs/heads/locked-branch", IsLocked: true},
	}, nil, "/repo")
	m.selected[0] = true
	m.view = confirmView

	output := stripAnsi(m.viewConfirm())

	if !strings.Contains(output, "LOCKED") {
		t.Error("should warn about locked worktree")
	}
	if !strings.Contains(output, "force-remove") {
		t.Error("should mention force-removal")
	}
}

func TestViewConfirm_UntrackedFiles(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/untracked", Branch: "refs/heads/untracked-branch", HasUntrackedFiles: true},
	}, nil, "/repo")
	m.selected[0] = true
	m.view = confirmView

	output := stripAnsi(m.viewConfirm())

	if !strings.Contains(output, "untracked files") {
		t.Error("should warn about untracked files")
	}
}
