package cmd_test

import (
	"os/exec"
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
