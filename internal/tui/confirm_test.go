package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/playground"
	"github.com/abiswas97/sentei/internal/worktree"
)

func TestViewConfirm_CleanWorktrees(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/feature-a"},
		{Path: "/work/b", Branch: "refs/heads/feature-b"},
	}, nil, "/repo")
	m.remove.selected["/work/a"] = true
	m.remove.selected["/work/b"] = true
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
	m.remove.selected["/work/dirty"] = true
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
	m.remove.selected["/work/locked"] = true
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
	m.remove.selected["/work/untracked"] = true
	m.view = confirmView

	output := stripAnsi(m.viewConfirm())

	if !strings.Contains(output, "untracked files") {
		t.Error("should warn about untracked files")
	}
}

func TestConfirmDeletion_UnlocksLockedWorktrees(t *testing.T) {
	tmp := t.TempDir()

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}

	repoPath := filepath.Join(tmp, "repo")
	run(tmp, "init", "--bare", "--initial-branch=main", repoPath)
	seed := filepath.Join(tmp, "_seed")
	run(tmp, "clone", repoPath, seed)
	run(seed, "commit", "--allow-empty", "-m", "init")
	run(seed, "push", "origin", "main")

	wtPath := filepath.Join(tmp, "locked-wt")
	run(repoPath, "worktree", "add", "-b", "locked-branch", wtPath)
	run(repoPath, "worktree", "lock", wtPath)

	runner := &git.GitRunner{}
	worktrees, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}
	worktrees = worktree.EnrichWorktrees(runner, worktrees, 1)

	m := NewModel(worktrees, runner, repoPath)
	// Select only the locked worktree
	for _, wt := range worktrees {
		if wt.IsLocked {
			m.remove.selected[wt.Path] = true
		}
	}
	m.view = confirmView

	// Send 'y' to confirm
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// Pump all commands until no more
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		model, cmd = model.Update(msg)
	}

	// Verify: directory should be gone
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("locked worktree directory should have been removed")
	}

	// Verify: git worktree list should not show the locked worktree
	out, _ := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain").CombinedOutput()
	if strings.Contains(string(out), "locked-branch") {
		t.Error("locked worktree should not appear in git worktree list after deletion and prune")
	}

	_ = model // silence unused variable warning
}

func TestPlayground_DeleteAll_IncludesLockedWorktree(t *testing.T) {
	repoPath, cleanup, err := playground.Setup()
	if err != nil {
		t.Fatalf("playground setup: %v", err)
	}
	defer cleanup()

	runner := &git.GitRunner{}
	worktrees, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}
	worktrees = worktree.EnrichWorktrees(runner, worktrees, 5)

	// Verify there's exactly one locked worktree
	var lockedCount int
	var lockedPath string
	for _, wt := range worktrees {
		if wt.IsLocked {
			lockedCount++
			lockedPath = wt.Path
		}
	}
	if lockedCount != 1 {
		t.Fatalf("expected 1 locked worktree, got %d", lockedCount)
	}

	m := NewModel(worktrees, runner, repoPath)
	// Select all non-bare, non-protected worktrees (including locked)
	for _, wt := range worktrees {
		if !wt.IsBare && !git.IsProtectedBranch(wt.Branch) {
			m.remove.selected[wt.Path] = true
		}
	}
	m.view = confirmView

	// Confirm deletion
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		model, cmd = model.Update(msg)
	}

	// After deletion + prune, re-list worktrees
	remaining, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		t.Fatalf("ListWorktrees after delete: %v", err)
	}

	for _, wt := range remaining {
		if wt.Path == lockedPath {
			t.Errorf("locked worktree %s should have been removed but still appears in git worktree list", lockedPath)
		}
		if wt.IsLocked {
			t.Errorf("no locked worktrees should remain, found: %s", wt.Path)
		}
	}

	_ = model // silence unused variable warning
}
