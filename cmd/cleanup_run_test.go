package cmd

import (
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/cleanup"
)

func TestParseCleanupRepoPath(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no args", nil, "."},
		{"flags only", []string{"--mode", "safe", "--dry-run"}, "."},
		{"positional only", []string{"/some/repo"}, "/some/repo"},
		{"flags then positional", []string{"--mode", "safe", "/some/repo"}, "/some/repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCleanupRepoPath(tt.args); got != tt.want {
				t.Errorf("ParseCleanupRepoPath(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestRunCleanup_InvalidMode(t *testing.T) {
	err := RunCleanup([]string{"--mode", "bogus"})
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "invalid value for --mode") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunCleanup_UnknownFlag(t *testing.T) {
	err := RunCleanup([]string{"--no-such-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestRunCleanup_MissingMode(t *testing.T) {
	err := RunCleanup(nil)
	if err == nil {
		t.Fatal("expected error for missing mode")
	}
	if !strings.Contains(err.Error(), "missing required flag: --mode") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunCleanup_SafeDryRun(t *testing.T) {
	bareRepo := setupBareRepo(t)

	var err error
	out := captureStdout(t, func() {
		err = RunCleanup([]string{"--mode", "safe", "--dry-run", bareRepo})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "(dry run)") {
		t.Errorf("expected '(dry run)' marker, got:\n%s", out)
	}
}

func TestRunCleanupWithOpts_ReportsStepErrors(t *testing.T) {
	notARepo := t.TempDir()

	var err error
	out := captureStdout(t, func() {
		err = RunCleanupWithOpts(&cleanup.Options{Mode: cleanup.ModeSafe}, notARepo)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "⚠") || !strings.Contains(out, "resolve-config") {
		t.Errorf("expected step error report, got:\n%s", out)
	}
}

func TestRunCleanupWithOpts_TipForNonWorktreeBranches(t *testing.T) {
	bareRepo := setupBareRepo(t)
	// A local branch not checked out in any worktree: safe mode leaves it and
	// should print the aggressive-mode tip.
	mustGit(t, bareRepo, "branch", "extra", "main")

	var err error
	out := captureStdout(t, func() {
		err = RunCleanupWithOpts(&cleanup.Options{Mode: cleanup.ModeSafe}, bareRepo)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sentei cleanup --mode=aggressive") {
		t.Errorf("expected aggressive-mode tip, got:\n%s", out)
	}
}
