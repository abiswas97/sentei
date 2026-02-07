package tui

import (
	"testing"
	"time"

	"github.com/abiswas97/sentei/internal/git"
)

func makeWorktrees() []git.Worktree {
	now := time.Now()
	return []git.Worktree{
		{Path: "/work/b-feature", Branch: "refs/heads/feature/auth", LastCommitDate: now.Add(-48 * time.Hour)},
		{Path: "/work/a-bugfix", Branch: "refs/heads/bugfix/nav", LastCommitDate: now.Add(-1 * time.Hour)},
		{Path: "/work/c-chore", Branch: "refs/heads/chore/deps", LastCommitDate: now.Add(-720 * time.Hour)},
	}
}

func TestReindex_SortByAgeAscending(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.sortField = SortByAge
	m.sortAscending = true
	m.reindex()

	if len(m.visibleIndices) != 3 {
		t.Fatalf("expected 3 visible, got %d", len(m.visibleIndices))
	}
	if m.worktrees[m.visibleIndices[0]].Path != "/work/c-chore" {
		t.Errorf("expected oldest first, got %s", m.worktrees[m.visibleIndices[0]].Path)
	}
	if m.worktrees[m.visibleIndices[2]].Path != "/work/a-bugfix" {
		t.Errorf("expected newest last, got %s", m.worktrees[m.visibleIndices[2]].Path)
	}
}

func TestReindex_SortByAgeDescending(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.sortField = SortByAge
	m.sortAscending = false
	m.reindex()

	if m.worktrees[m.visibleIndices[0]].Path != "/work/a-bugfix" {
		t.Errorf("expected newest first, got %s", m.worktrees[m.visibleIndices[0]].Path)
	}
}

func TestReindex_SortByBranch(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.sortField = SortByBranch
	m.sortAscending = true
	m.reindex()

	branches := make([]string, len(m.visibleIndices))
	for i, idx := range m.visibleIndices {
		branches[i] = stripBranchPrefix(m.worktrees[idx].Branch)
	}
	if branches[0] != "bugfix/nav" || branches[1] != "chore/deps" || branches[2] != "feature/auth" {
		t.Errorf("expected alphabetical order, got %v", branches)
	}
}

func TestReindex_FilterSubstring(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.filterText = "feature"
	m.reindex()

	if len(m.visibleIndices) != 1 {
		t.Fatalf("expected 1 match, got %d", len(m.visibleIndices))
	}
	if m.worktrees[m.visibleIndices[0]].Path != "/work/b-feature" {
		t.Errorf("expected feature/auth, got %s", m.worktrees[m.visibleIndices[0]].Path)
	}
}

func TestReindex_FilterCaseInsensitive(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.filterText = "BUGFIX"
	m.reindex()

	if len(m.visibleIndices) != 1 {
		t.Fatalf("expected 1 match, got %d", len(m.visibleIndices))
	}
	if m.worktrees[m.visibleIndices[0]].Branch != "refs/heads/bugfix/nav" {
		t.Errorf("expected bugfix/nav, got %s", m.worktrees[m.visibleIndices[0]].Branch)
	}
}

func TestReindex_FilterEmpty(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.filterText = ""
	m.reindex()

	if len(m.visibleIndices) != 3 {
		t.Errorf("expected all 3 visible with empty filter, got %d", len(m.visibleIndices))
	}
}

func TestReindex_FilterNoMatches(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.filterText = "zzz-nonexistent"
	m.reindex()

	if len(m.visibleIndices) != 0 {
		t.Errorf("expected 0 matches, got %d", len(m.visibleIndices))
	}
}

func TestReindex_CombinedSortAndFilter(t *testing.T) {
	now := time.Now()
	wts := []git.Worktree{
		{Path: "/work/1", Branch: "refs/heads/feature/a", LastCommitDate: now.Add(-10 * time.Hour)},
		{Path: "/work/2", Branch: "refs/heads/bugfix/b", LastCommitDate: now.Add(-5 * time.Hour)},
		{Path: "/work/3", Branch: "refs/heads/feature/c", LastCommitDate: now.Add(-20 * time.Hour)},
	}
	m := NewModel(wts, nil, "/repo")
	m.filterText = "feature"
	m.sortField = SortByAge
	m.sortAscending = true
	m.reindex()

	if len(m.visibleIndices) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(m.visibleIndices))
	}
	if m.worktrees[m.visibleIndices[0]].Path != "/work/3" {
		t.Errorf("expected oldest feature first, got %s", m.worktrees[m.visibleIndices[0]].Path)
	}
	if m.worktrees[m.visibleIndices[1]].Path != "/work/1" {
		t.Errorf("expected newer feature second, got %s", m.worktrees[m.visibleIndices[1]].Path)
	}
}

func TestReindex_SelectionPersistsAcrossSortChange(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.selected["/work/b-feature"] = true
	m.selected["/work/c-chore"] = true

	m.sortField = SortByBranch
	m.reindex()

	if !m.selected["/work/b-feature"] {
		t.Error("expected /work/b-feature to stay selected")
	}
	if !m.selected["/work/c-chore"] {
		t.Error("expected /work/c-chore to stay selected")
	}
	if len(m.selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(m.selected))
	}
}

func TestReindex_SelectAllWithFilter(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.filterText = "feature"
	m.reindex()

	for _, idx := range m.visibleIndices {
		m.selected[m.worktrees[idx].Path] = true
	}

	if len(m.selected) != 1 {
		t.Errorf("expected 1 selected (visible only), got %d", len(m.selected))
	}
	if !m.selected["/work/b-feature"] {
		t.Error("expected feature/auth to be selected")
	}
}

func TestReindex_CursorClampedOnFilter(t *testing.T) {
	m := NewModel(makeWorktrees(), nil, "/repo")
	m.cursor = 2

	m.filterText = "feature"
	m.reindex()

	if m.cursor >= len(m.visibleIndices) {
		t.Errorf("cursor %d should be < visible count %d", m.cursor, len(m.visibleIndices))
	}
}

func TestReindex_ZeroDateSortsToEnd(t *testing.T) {
	now := time.Now()
	wts := []git.Worktree{
		{Path: "/work/no-date", Branch: "refs/heads/x", LastCommitDate: time.Time{}},
		{Path: "/work/old", Branch: "refs/heads/y", LastCommitDate: now.Add(-100 * time.Hour)},
		{Path: "/work/new", Branch: "refs/heads/z", LastCommitDate: now.Add(-1 * time.Hour)},
	}

	m := NewModel(wts, nil, "/repo")
	m.sortField = SortByAge
	m.sortAscending = true
	m.reindex()

	lastIdx := m.visibleIndices[len(m.visibleIndices)-1]
	if m.worktrees[lastIdx].Path != "/work/no-date" {
		t.Errorf("expected zero-date at end ascending, got %s", m.worktrees[lastIdx].Path)
	}

	m.sortAscending = false
	m.reindex()

	lastIdx = m.visibleIndices[len(m.visibleIndices)-1]
	if m.worktrees[lastIdx].Path != "/work/no-date" {
		t.Errorf("expected zero-date at end descending, got %s", m.worktrees[lastIdx].Path)
	}
}
