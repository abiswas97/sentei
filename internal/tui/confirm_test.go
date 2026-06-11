package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/pipeline"
	"github.com/abiswas97/sentei/internal/playground"
	"github.com/abiswas97/sentei/internal/testtmp"
	"github.com/abiswas97/sentei/internal/testutil/mock"
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

	if !strings.Contains(output, "delete 2 worktrees") {
		t.Error("should mention count of worktrees")
	}
	if !strings.Contains(output, "feature-a") {
		t.Error("should list feature-a")
	}
	if !strings.Contains(output, "(clean)") {
		t.Error("should show (clean) label for clean worktrees")
	}
	if strings.Contains(output, "⚠") {
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

	if !strings.Contains(output, "uncommitted changes — will be lost") {
		t.Error("should warn about uncommitted changes")
	}
	if !strings.Contains(output, "⚠") {
		t.Error("should show a warning line for dirty worktrees")
	}
}

func TestViewConfirm_LockedWorktree(t *testing.T) {
	m := NewModel([]git.Worktree{
		{Path: "/work/locked", Branch: "refs/heads/locked-branch", IsLocked: true},
	}, nil, "/repo")
	m.remove.selected["/work/locked"] = true
	m.view = confirmView

	output := stripAnsi(m.viewConfirm())

	if !strings.Contains(output, "locked — will force-remove") {
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
	tmp := testtmp.RobustTempDir(t)

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
	run(seed, "config", "user.email", "test@test.com")
	run(seed, "config", "user.name", "Test")
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
	model, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})

	model = pumpCmds(model, cmd)

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
	model, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	model = pumpCmds(model, cmd)

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

func TestRunTeardownPhase_FallsBackToRemovingArtifactDirs(t *testing.T) {
	withArtifacts := t.TempDir()
	artifactDir := filepath.Join(withArtifacts, ".fake-artifact")
	if err := os.Mkdir(artifactDir, 0o755); err != nil {
		t.Fatal(err)
	}
	clean := t.TempDir()

	m := NewModel(nil, nil, "/repo")
	integrations := []integration.Integration{{
		Name:     "fake",
		Teardown: integration.TeardownSpec{Dirs: []string{".fake-artifact/"}},
	}}
	worktrees := []git.Worktree{{Path: withArtifacts}, {Path: clean}}

	msg := m.runTeardownPhase(worktrees, integrations)()

	done, ok := msg.(teardownCompleteMsg)
	if !ok {
		t.Fatalf("expected teardownCompleteMsg, got %T", msg)
	}
	if len(done.results) != 1 {
		t.Fatalf("results = %d, want 1 (only the worktree with artifacts)", len(done.results))
	}
	if done.results[0].Status != pipeline.StepDone {
		t.Errorf("teardown status = %v, want StepDone", done.results[0].Status)
	}
	if _, err := os.Stat(artifactDir); !os.IsNotExist(err) {
		t.Error("artifact dir should have been removed")
	}
}

func TestRunTeardownPhase_TeardownCommandHandlesRemoval(t *testing.T) {
	wtPath := t.TempDir()
	artifactDir := filepath.Join(wtPath, ".fake-artifact")
	if err := os.Mkdir(artifactDir, 0o755); err != nil {
		t.Fatal(err)
	}

	m := NewModel(nil, nil, "/repo")
	m.shell = &mock.Runner{Responses: map[string]mock.Response{
		wtPath + ":shell[fake clean]": {Output: "cleaned"},
	}}
	integrations := []integration.Integration{{
		Name:     "fake",
		Teardown: integration.TeardownSpec{Command: "fake clean", Dirs: []string{".fake-artifact/"}},
	}}

	msg := m.runTeardownPhase([]git.Worktree{{Path: wtPath}}, integrations)()

	done := msg.(teardownCompleteMsg)
	if len(done.results) != 1 || done.results[0].Status != pipeline.StepDone {
		t.Fatalf("results = %+v, want one successful teardown", done.results)
	}
	if _, err := os.Stat(artifactDir); err != nil {
		t.Error("a successful teardown command must not trigger the dir-removal fallback")
	}
}

func gateModel(wts []git.Worktree) Model {
	m := NewModel(wts, nil, "/repo")
	m.width, m.height = 90, 28
	m.view = listView
	m.remove.selected = map[string]bool{}
	for _, wt := range wts {
		m.remove.selected[wt.Path] = true
	}
	m.reindex()
	return m
}

func TestRemovalGate_CleanPushedSkipsConfirmation(t *testing.T) {
	m := gateModel([]git.Worktree{
		{Path: "/w/a", Branch: "refs/heads/a"},
		{Path: "/w/b", Branch: "refs/heads/b"},
	})

	updated, cmd := m.updateList(tea.KeyPressMsg{Code: tea.KeyEnter})
	model := updated.(Model)

	if model.view != progressView {
		t.Errorf("clean+pushed selection must skip the gate, got view %d", model.view)
	}
	if cmd == nil {
		t.Error("expected the deletion pipeline to start")
	}
	if model.remove.run.total() != 2 {
		t.Errorf("run total = %d, want 2", model.remove.run.total())
	}
}

func TestRemovalGate_AtRiskTriggersConfirmation(t *testing.T) {
	cases := []struct {
		name string
		wt   git.Worktree
	}{
		{"dirty", git.Worktree{Path: "/w/x", Branch: "refs/heads/x", HasUncommittedChanges: true}},
		{"untracked", git.Worktree{Path: "/w/x", Branch: "refs/heads/x", HasUntrackedFiles: true}},
		{"unpushed", git.Worktree{Path: "/w/x", Branch: "refs/heads/x", HasUnpushedCommits: true}},
		{"locked", git.Worktree{Path: "/w/x", Branch: "refs/heads/x", IsLocked: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := gateModel([]git.Worktree{
				{Path: "/w/clean", Branch: "refs/heads/clean"},
				tc.wt,
			})

			updated, _ := m.updateList(tea.KeyPressMsg{Code: tea.KeyEnter})
			model := updated.(Model)

			if model.view != confirmView {
				t.Errorf("%s selection must stop at the gate, got view %d", tc.name, model.view)
			}
		})
	}
}

func TestViewConfirm_UnpushedWarning(t *testing.T) {
	m := gateModel([]git.Worktree{
		{Path: "/w/x", Branch: "refs/heads/feature/x", HasUnpushedCommits: true},
	})
	m.view = confirmView

	view := stripANSI(m.viewConfirm())
	for _, want := range []string{
		"[^] commits not on any remote",
		"⚠ 1 worktree with commits not pushed to any remote",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
}
