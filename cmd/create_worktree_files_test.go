package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreateCopiesConfiguredWorktreeFilesWithoutFlags(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	bareRepo := setupSenteiBareRepo(t)
	templatePath := filepath.Join(bareRepo, ".sentei", "private", "codex.toml")
	if err := os.MkdirAll(filepath.Dir(templatePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(templatePath, []byte("private = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(bareRepo, ".sentei.yaml"), `
worktree_files:
  - source: .sentei/private/codex.toml
    destination: .codex/config.toml
`)

	var createErr error
	output := captureStdout(t, func() {
		createErr = RunCreate([]string{"--branch", "feature/private-config", "--base", "main", bareRepo})
	})
	if createErr != nil {
		t.Fatalf("RunCreate() error: %v", createErr)
	}
	if !strings.Contains(output, "copied: .codex/config.toml") {
		t.Fatalf("output missing copy diagnostic:\n%s", output)
	}
	destination := filepath.Join(bareRepo, "feature-private-config", ".codex", "config.toml")
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "private = true\n" {
		t.Errorf("destination content = %q", data)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("destination mode = %o, want 600", info.Mode().Perm())
	}
}

func setupSenteiBareRepo(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	repoRoot := filepath.Join(parent, "repo")
	bareDir := filepath.Join(repoRoot, ".bare")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	mustGit(t, parent, "init", "--bare", "--initial-branch=main", bareDir)
	mustWriteFile(t, filepath.Join(repoRoot, ".git"), "gitdir: .bare\n")

	cloneDir := filepath.Join(parent, "clone")
	mustGit(t, parent, "clone", bareDir, cloneDir)
	mustGit(t, cloneDir, "config", "user.email", "test@test.com")
	mustGit(t, cloneDir, "config", "user.name", "Test")
	mustWriteFile(t, filepath.Join(cloneDir, "README.md"), "# test\n")
	mustGit(t, cloneDir, "add", ".")
	mustGit(t, cloneDir, "commit", "-m", "initial commit")
	mustGit(t, cloneDir, "push", "origin", "main")
	mustGit(t, repoRoot, "worktree", "add", filepath.Join(repoRoot, "main"), "main")
	return repoRoot
}
