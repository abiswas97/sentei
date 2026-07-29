package worktreefile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func TestCopyCreatesNestedDestinationAndPreservesPermissions(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "private", "codex.toml"), "model = \"private\"\n", 0o600)

	outcome, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "private/codex.toml", Destination: ".codex/config.toml",
	}})
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}
	if len(outcome.Copied) != 1 || outcome.Copied[0] != ".codex/config.toml" {
		t.Fatalf("Copied = %#v", outcome.Copied)
	}
	destination := filepath.Join(worktreeRoot, ".codex", "config.toml")
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "model = \"private\"\n" {
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

func TestCopyPreservesExistingDestinationByDefault(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "template"), "new", 0o600)
	writeTestFile(t, filepath.Join(worktreeRoot, "config"), "existing", 0o644)

	outcome, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "template", Destination: "config",
	}})
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}
	if len(outcome.Skipped) != 1 || outcome.Skipped[0] != "config" {
		t.Fatalf("Skipped = %#v", outcome.Skipped)
	}
	data, err := os.ReadFile(filepath.Join(worktreeRoot, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing" {
		t.Errorf("existing destination was overwritten: %q", data)
	}
}

func TestCopyOverwritesExplicitly(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "template"), "new", 0o600)
	writeTestFile(t, filepath.Join(worktreeRoot, "config"), "existing", 0o644)

	_, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "template", Destination: "config", Overwrite: true,
	}})
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}
	info, err := os.Stat(filepath.Join(worktreeRoot, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("destination mode = %o, want 600", info.Mode().Perm())
	}
}

func TestCopyRejectsUnsafeRuntimePaths(t *testing.T) {
	tests := []Rule{
		{Source: "", Destination: "config"},
		{Source: "/absolute", Destination: "config"},
		{Source: "../escape", Destination: "config"},
		{Source: "template", Destination: "../escape"},
	}
	for _, rule := range tests {
		t.Run(rule.Source+"-"+rule.Destination, func(t *testing.T) {
			_, err := Copy(t.TempDir(), t.TempDir(), []Rule{rule})
			if err == nil {
				t.Fatal("Copy() error = nil, want unsafe path error")
			}
		})
	}
}

func TestCopyReportsMissingAndNonRegularSources(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	if _, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "missing", Destination: "config",
	}}); err == nil || !strings.Contains(err.Error(), `source "missing" does not exist`) {
		t.Fatalf("missing source error = %v", err)
	}

	if err := os.Mkdir(filepath.Join(repoRoot, "directory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "directory", Destination: "config",
	}}); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("directory source error = %v", err)
	}
}

func TestCopyRejectsSourceSymlinkEscape(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret")
	writeTestFile(t, outside, "private", 0o600)
	if err := os.Symlink(outside, filepath.Join(repoRoot, "template")); err != nil {
		t.Fatal(err)
	}

	_, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "template", Destination: "config",
	}})
	if err == nil || !strings.Contains(err.Error(), "outside repository root") {
		t.Fatalf("Copy() error = %v, want source escape", err)
	}
}

func TestCopyAllowsContainedSourceSymlink(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "private", "template"), "private", 0o600)
	if err := os.Symlink(filepath.Join("private", "template"), filepath.Join(repoRoot, "link")); err != nil {
		t.Fatal(err)
	}

	if _, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "link", Destination: "config",
	}}); err != nil {
		t.Fatalf("Copy() error: %v", err)
	}
}

func TestCopyRejectsDestinationSymlinks(t *testing.T) {
	tests := []struct {
		name        string
		linkTarget  func(worktreeRoot, outside string) string
		destination string
	}{
		{
			name: "ancestor",
			linkTarget: func(worktreeRoot, outside string) string {
				return filepath.Join(worktreeRoot, ".codex")
			},
			destination: ".codex/config.toml",
		},
		{
			name: "destination",
			linkTarget: func(worktreeRoot, outside string) string {
				return filepath.Join(worktreeRoot, "config")
			},
			destination: "config",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			worktreeRoot := t.TempDir()
			outside := t.TempDir()
			writeTestFile(t, filepath.Join(repoRoot, "template"), "private", 0o600)
			if err := os.Symlink(outside, tc.linkTarget(worktreeRoot, outside)); err != nil {
				t.Fatal(err)
			}
			_, err := Copy(repoRoot, worktreeRoot, []Rule{{
				Source: "template", Destination: tc.destination, Overwrite: true,
			}})
			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Fatalf("Copy() error = %v, want destination symlink rejection", err)
			}
		})
	}
}

func TestCopyReportsDestinationParentFailure(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "template"), "private", 0o600)
	writeTestFile(t, filepath.Join(worktreeRoot, ".codex"), "not a directory", 0o600)

	_, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "template", Destination: ".codex/config.toml",
	}})
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("Copy() error = %v, want destination parent failure", err)
	}
}
