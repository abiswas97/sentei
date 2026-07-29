package worktreefile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOutcomeMessage(t *testing.T) {
	tests := []struct {
		name    string
		outcome Outcome
		want    string
	}{
		{name: "empty", outcome: Outcome{}, want: ""},
		{name: "copied", outcome: Outcome{Copied: []string{"one", "two"}}, want: "copied: one, two"},
		{name: "skipped", outcome: Outcome{Skipped: []string{"one"}}, want: "preserved existing: one"},
		{
			name: "both",
			outcome: Outcome{
				Copied: []string{"one"}, Skipped: []string{"two"},
			},
			want: "copied: one; preserved existing: two",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.outcome.Message(); got != tc.want {
				t.Fatalf("Message() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCopyReportsInvalidRoots(t *testing.T) {
	fileRoot := filepath.Join(t.TempDir(), "file")
	writeTestFile(t, fileRoot, "not a directory", 0o600)
	tests := []struct {
		name         string
		repoRoot     string
		worktreeRoot string
		want         string
	}{
		{name: "missing repository", repoRoot: filepath.Join(t.TempDir(), "missing"), worktreeRoot: t.TempDir(), want: "repository root"},
		{name: "repository is file", repoRoot: fileRoot, worktreeRoot: t.TempDir(), want: "repository root is not a directory"},
		{name: "missing worktree", repoRoot: t.TempDir(), worktreeRoot: filepath.Join(t.TempDir(), "missing"), want: "worktree root"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Copy(tc.repoRoot, tc.worktreeRoot, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Copy() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestCopyRejectsDirectoryDestination(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "template"), "private", 0o600)
	if err := os.Mkdir(filepath.Join(worktreeRoot, "config"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "template", Destination: "config", Overwrite: true,
	}})
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("Copy() error = %v, want destination type failure", err)
	}
}

func TestCopyReportsSourceSymlinkResolutionFailure(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.Symlink("loop", filepath.Join(repoRoot, "loop")); err != nil {
		t.Fatal(err)
	}

	_, err := Copy(repoRoot, t.TempDir(), []Rule{{
		Source: "loop", Destination: "config",
	}})
	if err == nil || !strings.Contains(err.Error(), "resolving worktree file source") {
		t.Fatalf("Copy() error = %v, want symlink resolution failure", err)
	}
}

func TestCopyReportsInaccessibleDestinationParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission semantics required")
	}
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "template"), "private", 0o600)
	locked := filepath.Join(worktreeRoot, "locked")
	if err := os.Mkdir(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(locked, 0o700); err != nil {
			t.Errorf("restoring test directory permissions: %v", err)
		}
	})

	_, err := Copy(repoRoot, worktreeRoot, []Rule{{
		Source: "template", Destination: "locked/child/config",
	}})
	if err == nil || !strings.Contains(err.Error(), "inspecting parent") {
		t.Fatalf("Copy() error = %v, want inaccessible parent failure", err)
	}
}

func TestCopyReturnsCompletedOutcomeBeforeFailure(t *testing.T) {
	repoRoot := t.TempDir()
	worktreeRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "first"), "copied", 0o600)

	outcome, err := Copy(repoRoot, worktreeRoot, []Rule{
		{Source: "first", Destination: "first"},
		{Source: "missing", Destination: "second"},
	})
	if err == nil {
		t.Fatal("Copy() error = nil, want second-rule failure")
	}
	if len(outcome.Copied) != 1 || outcome.Copied[0] != "first" {
		t.Fatalf("Copied = %#v, want completed first rule", outcome.Copied)
	}
}

func TestCopyAtomicallyReportsWriteFailures(t *testing.T) {
	parent := t.TempDir()
	source := filepath.Join(parent, "source")
	writeTestFile(t, source, "content", 0o600)

	tests := []struct {
		name        string
		source      string
		destination string
		want        string
	}{
		{name: "missing parent", source: source, destination: filepath.Join(parent, "missing", "file"), want: "creating temporary destination"},
		{name: "missing source", source: filepath.Join(parent, "missing-source"), destination: filepath.Join(parent, "file"), want: "writing temporary destination"},
		{name: "destination directory", source: source, destination: t.TempDir(), want: "replacing destination"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := copyAtomically(tc.source, tc.destination, 0o600)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("copyAtomically() error = %v, want %q", err, tc.want)
			}
		})
	}
}
