package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmpBin := t.TempDir() + "/sentei"
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	build.Dir = ".."
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return tmpBin
}

func TestEcosystemsCLI(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "ecosystems")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sentei ecosystems failed: %v\n%s", err, out)
	}

	output := string(out)
	for _, want := range []string{"Ecosystems (", "pnpm", "go", "SOURCE"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
}

func TestUnknownCommandCLI(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "foobar")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for unknown command")
	}

	output := string(out)
	if !strings.Contains(output, "unknown command: foobar") {
		t.Errorf("expected 'unknown command' error, got:\n%s", output)
	}
}

func TestCleanupNonInteractive_MissingMode(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "cleanup", "--non-interactive", "--force")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for missing --mode")
	}

	output := string(out)
	if !strings.Contains(output, "missing required flag: --mode") {
		t.Errorf("expected 'missing required flag' error, got:\n%s", output)
	}
}

func TestCleanupNonInteractive_DestructiveWithoutForce(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "cleanup", "--non-interactive")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for destructive without --force")
	}

	output := string(out)
	if !strings.Contains(output, "destructive operation requires --force") {
		t.Errorf("expected '--force required' error, got:\n%s", output)
	}
}

func TestCleanupNonInteractive_SafeMode(t *testing.T) {
	bin := buildBinary(t)

	// Create a bare repo for the test.
	repoDir := t.TempDir()
	setupGitRepo(t, repoDir)

	cmd := exec.Command(bin, "cleanup", "--mode", "safe", "--non-interactive", "--force", repoDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sentei cleanup --mode safe --non-interactive --force failed: %v\n%s", err, out)
	}
	// Should produce some output (the cleanup ran).
	if len(out) == 0 {
		t.Error("expected non-empty output from cleanup")
	}
}

func TestCleanupNonInteractive_InvalidMode(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "cleanup", "--mode", "invalid", "--non-interactive", "--force")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for invalid mode")
	}

	output := string(out)
	if !strings.Contains(output, "invalid value for --mode") {
		t.Errorf("expected 'invalid value for --mode' error, got:\n%s", output)
	}
}

func setupGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", dir},
		{"-C", dir, "config", "user.email", "test@test.com"},
		{"-C", dir, "config", "user.name", "Test"},
	} {
		c := exec.Command("git", args...)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
}

func TestCloneNonInteractive_MissingURL(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "clone", "--non-interactive")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for missing --url")
	}

	output := string(out)
	if !strings.Contains(output, "missing required flag: --url") {
		t.Errorf("expected 'missing required flag: --url' error, got:\n%s", output)
	}
}

func setupBareRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "test.git")

	runGitCmd(t, tmpDir, "init", "--bare", bareRepo)

	cloneDir := filepath.Join(tmpDir, "clone")
	runGitCmd(t, tmpDir, "clone", bareRepo, cloneDir)
	runGitCmd(t, cloneDir, "config", "user.email", "test@test.com")
	runGitCmd(t, cloneDir, "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(cloneDir, "README.md"), []byte("# test\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	runGitCmd(t, cloneDir, "add", ".")
	runGitCmd(t, cloneDir, "commit", "-m", "initial commit")
	runGitCmd(t, cloneDir, "push", "origin", "main")

	return bareRepo
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed in %s: %v\n%s", args, dir, err, out)
	}
}

func TestCreateNonInteractive_Success(t *testing.T) {
	bin := buildBinary(t)
	bareRepo := setupBareRepo(t)

	cmd := exec.Command(bin, "create", "--branch", "feature/e2e-test", "--base", "main", "--non-interactive", bareRepo)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sentei create --non-interactive failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Worktree created") {
		t.Errorf("expected 'Worktree created' in output, got:\n%s", output)
	}

	// Verify the worktree directory actually exists.
	wtPath := filepath.Join(bareRepo, "feature-e2e-test")
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("expected worktree directory to exist at %s", wtPath)
	}
}

func TestCreateNonInteractive_MissingBranch(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "create", "--non-interactive")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for missing --branch")
	}

	output := string(out)
	if !strings.Contains(output, "missing required flag: --branch") {
		t.Errorf("expected 'missing required flag: --branch' error, got:\n%s", output)
	}
}

func TestIntegrationsCLI(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "integrations")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sentei integrations failed: %v\n%s", err, out)
	}

	output := string(out)
	for _, want := range []string{"Integrations (2", "code-review-graph", "cocoindex-code", "https://github.com/"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
}
