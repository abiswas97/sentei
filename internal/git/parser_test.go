package git

import (
	"testing"
)

func TestParsePorcelain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Worktree
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []Worktree{},
		},
		{
			name: "single bare entry",
			input: `worktree /Users/dev/repo
bare`,
			expected: []Worktree{
				{Path: "/Users/dev/repo", IsBare: true},
			},
		},
		{
			name: "normal worktree",
			input: `worktree /Users/dev/repo/feature-x
HEAD abc123def456
branch refs/heads/feature-x`,
			expected: []Worktree{
				{
					Path:   "/Users/dev/repo/feature-x",
					HEAD:   "abc123def456",
					Branch: "refs/heads/feature-x",
				},
			},
		},
		{
			name: "detached HEAD",
			input: `worktree /Users/dev/repo/detached
HEAD abc123
detached`,
			expected: []Worktree{
				{
					Path:       "/Users/dev/repo/detached",
					HEAD:       "abc123",
					IsDetached: true,
				},
			},
		},
		{
			name: "locked without reason",
			input: `worktree /Users/dev/repo/locked-branch
HEAD 789abc123def
branch refs/heads/locked-branch
locked`,
			expected: []Worktree{
				{
					Path:     "/Users/dev/repo/locked-branch",
					HEAD:     "789abc123def",
					Branch:   "refs/heads/locked-branch",
					IsLocked: true,
				},
			},
		},
		{
			name: "locked with reason",
			input: `worktree /Users/dev/repo/locked-branch
HEAD 789abc123def
branch refs/heads/locked-branch
locked important work in progress`,
			expected: []Worktree{
				{
					Path:       "/Users/dev/repo/locked-branch",
					HEAD:       "789abc123def",
					Branch:     "refs/heads/locked-branch",
					IsLocked:   true,
					LockReason: "important work in progress",
				},
			},
		},
		{
			name: "prunable without reason",
			input: `worktree /Users/dev/repo/old-branch
HEAD def456
branch refs/heads/old-branch
prunable`,
			expected: []Worktree{
				{
					Path:       "/Users/dev/repo/old-branch",
					HEAD:       "def456",
					Branch:     "refs/heads/old-branch",
					IsPrunable: true,
				},
			},
		},
		{
			name: "prunable with reason",
			input: `worktree /Users/dev/repo/old-branch
HEAD def456
branch refs/heads/old-branch
prunable gitdir file points to non-existent location`,
			expected: []Worktree{
				{
					Path:        "/Users/dev/repo/old-branch",
					HEAD:        "def456",
					Branch:      "refs/heads/old-branch",
					IsPrunable:  true,
					PruneReason: "gitdir file points to non-existent location",
				},
			},
		},
		{
			name: "multiple worktrees",
			input: `worktree /Users/dev/repo
bare

worktree /Users/dev/repo/main
HEAD abc123def456
branch refs/heads/main

worktree /Users/dev/repo/feature-x
HEAD def456abc789
branch refs/heads/feature-x

worktree /Users/dev/repo/locked-branch
HEAD 789abc123def
branch refs/heads/locked-branch
locked`,
			expected: []Worktree{
				{Path: "/Users/dev/repo", IsBare: true},
				{Path: "/Users/dev/repo/main", HEAD: "abc123def456", Branch: "refs/heads/main"},
				{Path: "/Users/dev/repo/feature-x", HEAD: "def456abc789", Branch: "refs/heads/feature-x"},
				{Path: "/Users/dev/repo/locked-branch", HEAD: "789abc123def", Branch: "refs/heads/locked-branch", IsLocked: true},
			},
		},
		{
			name: "branch ref preserved exactly",
			input: `worktree /tmp/wt
HEAD aaa
branch refs/heads/feature/deep/nested`,
			expected: []Worktree{
				{Path: "/tmp/wt", HEAD: "aaa", Branch: "refs/heads/feature/deep/nested"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePorcelain(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d worktrees, want %d", len(got), len(tt.expected))
			}
			for i := range got {
				assertWorktreeEqual(t, got[i], tt.expected[i])
			}
		})
	}
}

func assertWorktreeEqual(t *testing.T, got, want Worktree) {
	t.Helper()
	if got.Path != want.Path {
		t.Errorf("Path = %q, want %q", got.Path, want.Path)
	}
	if got.HEAD != want.HEAD {
		t.Errorf("HEAD = %q, want %q", got.HEAD, want.HEAD)
	}
	if got.Branch != want.Branch {
		t.Errorf("Branch = %q, want %q", got.Branch, want.Branch)
	}
	if got.IsBare != want.IsBare {
		t.Errorf("IsBare = %v, want %v", got.IsBare, want.IsBare)
	}
	if got.IsLocked != want.IsLocked {
		t.Errorf("IsLocked = %v, want %v", got.IsLocked, want.IsLocked)
	}
	if got.LockReason != want.LockReason {
		t.Errorf("LockReason = %q, want %q", got.LockReason, want.LockReason)
	}
	if got.IsPrunable != want.IsPrunable {
		t.Errorf("IsPrunable = %v, want %v", got.IsPrunable, want.IsPrunable)
	}
	if got.PruneReason != want.PruneReason {
		t.Errorf("PruneReason = %q, want %q", got.PruneReason, want.PruneReason)
	}
	if got.IsDetached != want.IsDetached {
		t.Errorf("IsDetached = %v, want %v", got.IsDetached, want.IsDetached)
	}
}
