package tui

import (
	"testing"
	"time"

	"github.com/abiswas97/sentei/internal/git"
)

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{
			name: "zero time",
			t:    time.Time{},
			want: "unknown",
		},
		{
			name: "seconds ago",
			t:    time.Now().Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "1 minute ago",
			t:    time.Now().Add(-1 * time.Minute),
			want: "1 minute ago",
		},
		{
			name: "5 minutes ago",
			t:    time.Now().Add(-5 * time.Minute),
			want: "5 minutes ago",
		},
		{
			name: "1 hour ago",
			t:    time.Now().Add(-1 * time.Hour),
			want: "1 hour ago",
		},
		{
			name: "2 hours ago",
			t:    time.Now().Add(-2 * time.Hour),
			want: "2 hours ago",
		},
		{
			name: "1 day ago",
			t:    time.Now().Add(-24 * time.Hour),
			want: "1 day ago",
		},
		{
			name: "15 days ago",
			t:    time.Now().Add(-15 * 24 * time.Hour),
			want: "15 days ago",
		},
		{
			name: "3 months ago",
			t:    time.Now().Add(-90 * 24 * time.Hour),
			want: "3 months ago",
		},
		{
			name: "1 year ago",
			t:    time.Now().Add(-365 * 24 * time.Hour),
			want: "1 year ago",
		},
		{
			name: "2 years ago",
			t:    time.Now().Add(-730 * 24 * time.Hour),
			want: "2 years ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := relativeTime(tt.t)
			if got != tt.want {
				t.Errorf("relativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripBranchPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"refs/heads/main", "main"},
		{"refs/heads/feature/auth", "feature/auth"},
		{"main", "main"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripBranchPrefix(tt.input)
			if got != tt.want {
				t.Errorf("stripBranchPrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStatusIndicator(t *testing.T) {
	tests := []struct {
		name string
		wt   git.Worktree
		want string
	}{
		{
			name: "clean worktree",
			wt:   git.Worktree{},
			want: "[ok]",
		},
		{
			name: "dirty worktree",
			wt:   git.Worktree{HasUncommittedChanges: true},
			want: "[~]",
		},
		{
			name: "untracked files",
			wt:   git.Worktree{HasUntrackedFiles: true},
			want: "[!]",
		},
		{
			name: "locked takes priority",
			wt:   git.Worktree{IsLocked: true, HasUncommittedChanges: true},
			want: "[L]",
		},
		{
			name: "dirty takes priority over untracked",
			wt:   git.Worktree{HasUncommittedChanges: true, HasUntrackedFiles: true},
			want: "[~]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusIndicator(tt.wt)
			if !containsText(got, tt.want) {
				t.Errorf("statusIndicator() rendered %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

func containsText(rendered, text string) bool {
	// Strip ANSI escape sequences for comparison
	clean := stripAnsi(rendered)
	return clean == text
}

func stripAnsi(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
			continue
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}
