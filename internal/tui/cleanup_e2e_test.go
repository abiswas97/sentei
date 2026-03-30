package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/worktree"
)

// setupBareRepoModel creates a Model connected to a real bare repo with worktrees loaded.
func setupBareRepoModel(t *testing.T) Model {
	t.Helper()

	tmpDir := t.TempDir()
	repoPath := createBareRepoWithWorktrees(t, tmpDir)

	runner := &git.GitRunner{}
	cfg, _ := config.LoadConfig(repoPath, config.WithRunner(runner))

	m := NewMenuModel(runner, nil, repoPath, cfg, repo.ContextBareRepo)

	// Pre-load worktrees so the menu is fully populated.
	wts, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		t.Fatalf("listing worktrees: %v", err)
	}
	wts = worktree.EnrichWorktrees(runner, wts, 10)
	var filtered []git.Worktree
	for _, wt := range wts {
		if !wt.IsBare {
			filtered = append(filtered, wt)
		}
	}
	m.remove.worktrees = filtered
	m.reindex()
	m.updateMenuHints()

	return m
}

// createBareRepoWithWorktrees creates a minimal bare repo with one worktree.
func createBareRepoWithWorktrees(t *testing.T, baseDir string) string {
	t.Helper()

	bareRepo := baseDir + "/test.git"

	runGitCmd(t, baseDir, "init", "--bare", bareRepo)

	cloneDir := baseDir + "/clone"
	runGitCmd(t, baseDir, "clone", bareRepo, cloneDir)
	runGitCmd(t, cloneDir, "config", "user.email", "test@test.com")
	runGitCmd(t, cloneDir, "config", "user.name", "Test")

	writeTestFile(t, cloneDir+"/README.md", "# Test\n")
	runGitCmd(t, cloneDir, "add", ".")
	runGitCmd(t, cloneDir, "commit", "-m", "initial commit")
	runGitCmd(t, cloneDir, "push", "origin", "main")

	// Create one worktree.
	wtPath := bareRepo + "/wt-0"
	runGitCmd(t, bareRepo, "worktree", "add", "-b", "feature/wt-0", wtPath, "main")
	runGitCmd(t, wtPath, "config", "user.email", "test@test.com")
	runGitCmd(t, wtPath, "config", "user.name", "Test")
	writeTestFile(t, wtPath+"/file.txt", "content\n")
	runGitCmd(t, wtPath, "add", ".")
	runGitCmd(t, wtPath, "commit", "-m", "commit on wt-0")

	return bareRepo
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed in %s: %v\n%s", args, dir, err, out)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func TestE2E_MenuToCleanupConfirmToResult(t *testing.T) {
	m := setupBareRepoModel(t)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// Give the TUI a moment to render.
	time.Sleep(200 * time.Millisecond)

	// Navigate to "Cleanup & exit" (index 3). Start at index 0.
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	// Press Enter to select Cleanup — should show confirmation.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(200 * time.Millisecond)

	// Press Enter to confirm — should start cleanup and show results.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for cleanup to complete, then press Enter to quit.
	time.Sleep(2 * time.Second)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	out := readFinalOutput(t, tm.FinalOutput(t))
	if len(out) == 0 {
		t.Error("expected non-empty final output from TUI")
	}

	// The final output should include cleanup result content.
	// teatest captures terminal frames which may have control sequences,
	// so we check for key substrings that appear in the cleanup result view.
	stripped := stripAnsi(string(out))
	if !strings.Contains(stripped, "Cleanup") {
		t.Errorf("expected 'Cleanup' in output, got:\n%s", stripped)
	}
}

func TestE2E_CleanupWithModeSafePrefilled(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := createBareRepoWithWorktrees(t, tmpDir)

	runner := &git.GitRunner{}
	cfg, _ := config.LoadConfig(repoPath, config.WithRunner(runner))

	m := NewMenuModel(runner, nil, repoPath, cfg, repo.ContextBareRepo)
	m.SetCleanupOpts(&cleanup.Options{Mode: cleanup.ModeSafe})

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// Give the TUI a moment to render the confirmation view.
	time.Sleep(200 * time.Millisecond)

	// Should start directly at the confirmation view with mode=safe.
	// Press Enter to confirm.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for cleanup to complete, then press Enter to quit.
	time.Sleep(2 * time.Second)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	out := readFinalOutput(t, tm.FinalOutput(t))
	stripped := stripAnsi(string(out))

	// Verify the flow reached the cleanup result view.
	if !strings.Contains(stripped, "Cleanup") {
		t.Errorf("expected 'Cleanup' in output, got:\n%s", stripped)
	}
}

func TestE2E_CleanupConfirmBackQuitsWhenDirectLaunch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := createBareRepoWithWorktrees(t, tmpDir)

	runner := &git.GitRunner{}
	cfg, _ := config.LoadConfig(repoPath, config.WithRunner(runner))

	m := NewMenuModel(runner, nil, repoPath, cfg, repo.ContextBareRepo)
	m.SetCleanupOpts(&cleanup.Options{Mode: cleanup.ModeSafe})

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	time.Sleep(200 * time.Millisecond)

	// Press Esc to go back — should quit since we launched directly.
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})

	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	out := readFinalOutput(t, tm.FinalOutput(t))
	if len(out) == 0 {
		t.Error("expected non-empty final output")
	}
}

func readFinalOutput(t *testing.T, r interface{ Read([]byte) (int, error) }) []byte {
	t.Helper()
	buf := make([]byte, 8192)
	var result []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return result
}
