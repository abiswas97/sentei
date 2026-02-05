package playground

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSetup_CreatesExpectedWorktrees(t *testing.T) {
	repoPath, cleanup, err := Setup()
	if err != nil {
		t.Fatalf("Setup() error: %v", err)
	}
	defer cleanup()

	out, err := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain").Output()
	if err != nil {
		t.Fatalf("git worktree list error: %v", err)
	}

	porcelain := string(out)

	blocks := splitWorktreeBlocks(porcelain)
	// bare repo + 6 worktrees = 7 blocks
	if len(blocks) != 7 {
		t.Errorf("expected 7 worktree blocks, got %d\nporcelain:\n%s", len(blocks), porcelain)
	}

	expectedBranches := []string{
		"refs/heads/feature/active",
		"refs/heads/feature/wip",
		"refs/heads/experiment/abandoned",
		"refs/heads/hotfix/locked",
		"refs/heads/chore/old-deps",
	}
	for _, branch := range expectedBranches {
		if !strings.Contains(porcelain, "branch "+branch) {
			t.Errorf("expected branch %q in porcelain output", branch)
		}
	}

	if !strings.Contains(porcelain, "locked") {
		t.Error("expected a locked worktree in output")
	}

	if !strings.Contains(porcelain, "detached") {
		t.Error("expected a detached worktree in output")
	}
}

func TestSetup_Idempotent(t *testing.T) {
	_, cleanup1, err := Setup()
	if err != nil {
		t.Fatalf("first Setup() error: %v", err)
	}
	cleanup1()

	_, cleanup2, err := Setup()
	if err != nil {
		t.Fatalf("second Setup() error: %v", err)
	}
	defer cleanup2()
}

func TestSetup_CleanupRemovesDir(t *testing.T) {
	_, cleanup, err := Setup()
	if err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	if _, err := os.Stat(PlaygroundDir); os.IsNotExist(err) {
		t.Fatal("playground dir should exist before cleanup")
	}

	cleanup()

	if _, err := os.Stat(PlaygroundDir); !os.IsNotExist(err) {
		t.Fatal("playground dir should not exist after cleanup")
	}
}

func splitWorktreeBlocks(porcelain string) []string {
	var blocks []string
	var current []string
	for _, line := range strings.Split(porcelain, "\n") {
		if line == "" {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}
