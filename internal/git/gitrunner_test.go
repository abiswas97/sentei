package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) < 3 || args[0] != "-C" {
		fmt.Fprintf(os.Stderr, "expected -C <dir> <cmd...>, got: %v", args)
		os.Exit(1)
	}

	dir := args[1]
	gitArgs := args[2:]

	switch {
	case len(gitArgs) >= 2 && gitArgs[0] == "rev-parse" && gitArgs[1] == "--git-dir":
		if strings.Contains(dir, "not-a-repo") {
			fmt.Fprintf(os.Stderr, "fatal: not a git repository (or any of the parent directories): .git")
			os.Exit(128)
		}
		fmt.Print(".")
		os.Exit(0)

	case len(gitArgs) >= 3 && gitArgs[0] == "worktree" && gitArgs[1] == "list" && gitArgs[2] == "--porcelain":
		fmt.Print("worktree /tmp/test-repo\nbare\n\nworktree /tmp/test-repo/main\nHEAD abc123\nbranch refs/heads/main")
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "unexpected git command: %v", gitArgs)
		os.Exit(1)
	}
}

func installFakeGit(t *testing.T) {
	t.Helper()

	testBin, err := os.Executable()
	if err != nil {
		t.Fatalf("could not find test executable: %v", err)
	}

	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("could not create bin dir: %v", err)
	}

	script := filepath.Join(binDir, "git")
	content := fmt.Sprintf("#!/bin/sh\nexec %q -test.run=^TestHelperProcess$ -- \"$@\"\n", testBin)
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatalf("could not write fake git: %v", err)
	}

	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestGitRunner_SuccessfulCommand(t *testing.T) {
	installFakeGit(t)
	runner := &GitRunner{}

	output, err := runner.Run("/tmp/some-repo", "rev-parse", "--git-dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "." {
		t.Errorf("output = %q, want %q", output, ".")
	}
}

func TestGitRunner_FailedCommand(t *testing.T) {
	installFakeGit(t)
	runner := &GitRunner{}

	_, err := runner.Run("/tmp/not-a-repo", "rev-parse", "--git-dir")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error = %q, want it to contain 'not a git repository'", err.Error())
	}
}

func TestGitRunner_StdoutTrimmed(t *testing.T) {
	installFakeGit(t)
	runner := &GitRunner{}

	output, err := runner.Run("/tmp/some-repo", "worktree", "list", "--porcelain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.HasPrefix(output, " ") || strings.HasPrefix(output, "\n") {
		t.Errorf("output has leading whitespace: %q", output)
	}
	if strings.HasSuffix(output, " ") || strings.HasSuffix(output, "\n") {
		t.Errorf("output has trailing whitespace: %q", output)
	}
}

func TestGitRunner_WorktreeListIntegration(t *testing.T) {
	installFakeGit(t)
	runner := &GitRunner{}

	output, err := runner.Run("/tmp/some-repo", "worktree", "list", "--porcelain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	worktrees, err := ParsePorcelain(output)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(worktrees) != 2 {
		t.Fatalf("got %d worktrees, want 2", len(worktrees))
	}
	if !worktrees[0].IsBare {
		t.Error("first worktree should be bare")
	}
	if worktrees[1].Branch != "refs/heads/main" {
		t.Errorf("second worktree branch = %q, want refs/heads/main", worktrees[1].Branch)
	}
}
