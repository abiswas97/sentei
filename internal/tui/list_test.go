package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
			got := stripAnsi(statusIndicator(tt.wt))
			if got != tt.want {
				t.Errorf("statusIndicator() rendered %q, want %q", got, tt.want)
			}
		})
	}
}

func TestViewLegend(t *testing.T) {
	m := Model{}
	got := stripAnsi(m.viewLegend())

	for _, want := range []string{"[ok] clean", "[~] dirty", "[!] untracked", "[L] locked", "[P] protected"} {
		if !strings.Contains(got, want) {
			t.Errorf("viewLegend() = %q, want it to contain %q", got, want)
		}
	}
}

func keyMsg(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func TestToggle_SkipsProtectedWorktree(t *testing.T) {
	wts := []git.Worktree{
		{Path: "/work/main", Branch: "refs/heads/main"},
		{Path: "/work/feature", Branch: "refs/heads/feature/x"},
	}
	m := NewModel(wts, nil, "/repo")
	m.cursor = 0

	updated, _ := m.Update(keyMsg(" "))
	m = updated.(Model)

	if m.selected["/work/main"] {
		t.Error("protected worktree should not be selectable via spacebar")
	}

	m.cursor = 1
	updated, _ = m.Update(keyMsg(" "))
	m = updated.(Model)

	if !m.selected["/work/feature"] {
		t.Error("non-protected worktree should be selectable via spacebar")
	}
}

func TestSelectAll_SkipsProtectedWorktrees(t *testing.T) {
	wts := []git.Worktree{
		{Path: "/work/main", Branch: "refs/heads/main"},
		{Path: "/work/feature", Branch: "refs/heads/feature/x"},
		{Path: "/work/bugfix", Branch: "refs/heads/bugfix/y"},
	}
	m := NewModel(wts, nil, "/repo")

	updated, _ := m.Update(keyMsg("a"))
	m = updated.(Model)

	if m.selected["/work/main"] {
		t.Error("protected worktree should not be selected by select-all")
	}
	if !m.selected["/work/feature"] {
		t.Error("non-protected worktree should be selected by select-all")
	}
	if !m.selected["/work/bugfix"] {
		t.Error("non-protected worktree should be selected by select-all")
	}
	if len(m.selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(m.selected))
	}
}

func TestSelectAll_WithFilter_SkipsProtected(t *testing.T) {
	now := time.Now()
	wts := []git.Worktree{
		{Path: "/work/main", Branch: "refs/heads/main", LastCommitDate: now},
		{Path: "/work/feat-a", Branch: "refs/heads/feature/a", LastCommitDate: now},
		{Path: "/work/feat-b", Branch: "refs/heads/feature/b", LastCommitDate: now},
		{Path: "/work/dev", Branch: "refs/heads/dev", LastCommitDate: now},
		{Path: "/work/bugfix", Branch: "refs/heads/bugfix/z", LastCommitDate: now},
	}
	m := NewModel(wts, nil, "/repo")

	m.filterText = "feat"
	m.reindex()

	if len(m.visibleIndices) != 2 {
		t.Fatalf("expected 2 visible after filter, got %d", len(m.visibleIndices))
	}

	updated, _ := m.Update(keyMsg("a"))
	m = updated.(Model)

	if !m.selected["/work/feat-a"] || !m.selected["/work/feat-b"] {
		t.Error("visible non-protected worktrees should be selected")
	}
	if m.selected["/work/main"] || m.selected["/work/dev"] {
		t.Error("protected worktrees should not be selected even if hidden by filter")
	}
	if len(m.selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(m.selected))
	}
}

func TestDeselectAll_WithProtected(t *testing.T) {
	wts := []git.Worktree{
		{Path: "/work/main", Branch: "refs/heads/main"},
		{Path: "/work/feature", Branch: "refs/heads/feature/x"},
		{Path: "/work/bugfix", Branch: "refs/heads/bugfix/y"},
	}
	m := NewModel(wts, nil, "/repo")
	m.selected["/work/feature"] = true
	m.selected["/work/bugfix"] = true

	updated, _ := m.Update(keyMsg("a"))
	m = updated.(Model)

	if len(m.selected) != 0 {
		t.Errorf("expected 0 selected after deselect-all, got %d", len(m.selected))
	}
}
