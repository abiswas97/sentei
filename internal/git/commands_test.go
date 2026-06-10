package git

import (
	"errors"
	"fmt"
	"testing"
)

type mockRunner struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	output string
	err    error
}

func (m *mockRunner) Run(dir string, args ...string) (string, error) {
	key := fmt.Sprintf("%s:%v", dir, args)
	if resp, ok := m.responses[key]; ok {
		return resp.output, resp.err
	}
	return "", fmt.Errorf("unexpected call: %s", key)
}

func TestListWorktrees_Success(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo:[rev-parse --git-dir]": {output: "."},
			"/repo:[worktree list --porcelain]": {
				output: "worktree /repo\nbare\n\nworktree /repo/main\nHEAD abc123\nbranch refs/heads/main",
			},
		},
	}

	wts, err := ListWorktrees(runner, "/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 2 {
		t.Fatalf("got %d worktrees, want 2", len(wts))
	}
	if !wts[0].IsBare {
		t.Error("first worktree should be bare")
	}
	if wts[1].Branch != "refs/heads/main" {
		t.Errorf("second worktree branch = %q, want refs/heads/main", wts[1].Branch)
	}
}

func TestListWorktrees_RepoValidationFailure(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/not-a-repo:[rev-parse --git-dir]": {err: fmt.Errorf("git rev-parse --git-dir: fatal: not a git repository")},
		},
	}

	wts, err := ListWorktrees(runner, "/not-a-repo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if wts != nil {
		t.Errorf("expected nil worktrees, got %v", wts)
	}
}

func TestListWorktrees_GitCommandFailure(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo:[rev-parse --git-dir]":       {output: "."},
			"/repo:[worktree list --porcelain]": {err: fmt.Errorf("git worktree list --porcelain: permission denied")},
		},
	}

	_, err := ListWorktrees(runner, "/repo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListWorktrees_EmptyOutput(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo:[rev-parse --git-dir]":       {output: "."},
			"/repo:[worktree list --porcelain]": {output: ""},
		},
	}

	wts, err := ListWorktrees(runner, "/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 0 {
		t.Fatalf("got %d worktrees, want 0", len(wts))
	}
}

func TestValidateRepository_NotARepo(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/bad:[rev-parse --git-dir]": {err: fmt.Errorf("fatal: not a git repository")},
		},
	}

	err := ValidateRepository(runner, "/bad")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateRepository_PreservesUnderlyingCause(t *testing.T) {
	cause := errors.New("exec: \"git\": executable file not found in $PATH")
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/x:[rev-parse --git-dir]": {err: cause},
		},
	}

	err := ValidateRepository(runner, "/x")
	if err == nil {
		t.Fatal("expected error")
	}
	// The real cause (git missing) must survive instead of being asserted away.
	if !errors.Is(err, cause) {
		t.Errorf("underlying cause not preserved: %v", err)
	}
}

func TestValidateRepository_PathDoesNotExist(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/nonexistent:[rev-parse --git-dir]": {err: fmt.Errorf("cannot change to '/nonexistent': No such file or directory")},
		},
	}

	err := ValidateRepository(runner, "/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"/plain/path": "'/plain/path'",
		"/with space": "'/with space'",
		"a&&rm -rf x": "'a&&rm -rf x'",
		"it's":        `'it'\''s'`,
	}
	for in, want := range cases {
		if got := ShellQuote(in); got != want {
			t.Errorf("ShellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}
