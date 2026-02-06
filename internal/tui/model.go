package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/worktree"
)

type viewState int

const (
	listView viewState = iota
	confirmView
	progressView
	summaryView
)

type SortField int

const (
	SortByAge SortField = iota
	SortByBranch
)

type Model struct {
	worktrees      []git.Worktree
	selected       map[string]bool
	visibleIndices []int
	cursor         int
	offset         int
	width          int
	height         int

	sortField     SortField
	sortAscending bool

	filterText   string
	filterActive bool
	filterInput  textinput.Model

	view viewState

	runner   git.CommandRunner
	repoPath string

	deletionStatuses map[string]string
	deletionResult   worktree.DeletionResult
	deletionTotal    int
	progressCh       <-chan worktree.DeletionEvent

	pruneErr *error
}

func NewModel(worktrees []git.Worktree, runner git.CommandRunner, repoPath string) Model {
	ti := textinput.New()
	ti.Prompt = "filter: "

	m := Model{
		worktrees:        worktrees,
		selected:         make(map[string]bool),
		sortField:        SortByAge,
		sortAscending:    true,
		filterInput:      ti,
		runner:           runner,
		repoPath:         repoPath,
		deletionStatuses: make(map[string]string),
		height:           20,
	}
	m.reindex()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.view {
	case listView:
		return m.updateList(msg)
	case confirmView:
		return m.updateConfirm(msg)
	case progressView:
		return m.updateProgress(msg)
	case summaryView:
		return m.updateSummary(msg)
	}
	return m, nil
}

func (m Model) View() string {
	switch m.view {
	case listView:
		return m.viewList()
	case confirmView:
		return m.viewConfirm()
	case progressView:
		return m.viewProgress()
	case summaryView:
		return m.viewSummary()
	}
	return ""
}

func (m Model) selectedWorktrees() []git.Worktree {
	var result []git.Worktree
	for _, wt := range m.worktrees {
		if m.selected[wt.Path] {
			result = append(result, wt)
		}
	}
	return result
}

func (m *Model) reindex() {
	filterLower := strings.ToLower(m.filterText)

	var indices []int
	for i, wt := range m.worktrees {
		if filterLower != "" {
			branch := strings.ToLower(stripBranchPrefix(wt.Branch))
			if !strings.Contains(branch, filterLower) {
				continue
			}
		}
		indices = append(indices, i)
	}

	sortAsc := m.sortAscending
	sortField := m.sortField
	wts := m.worktrees

	sort.SliceStable(indices, func(a, b int) bool {
		wa, wb := wts[indices[a]], wts[indices[b]]

		switch sortField {
		case SortByAge:
			aZero := wa.LastCommitDate.IsZero()
			bZero := wb.LastCommitDate.IsZero()
			if aZero != bZero {
				return !aZero
			}
			if aZero && bZero {
				return false
			}
			if sortAsc {
				return wa.LastCommitDate.Before(wb.LastCommitDate)
			}
			return wa.LastCommitDate.After(wb.LastCommitDate)

		case SortByBranch:
			ba := strings.ToLower(stripBranchPrefix(wa.Branch))
			bb := strings.ToLower(stripBranchPrefix(wb.Branch))
			if sortAsc {
				return ba < bb
			}
			return ba > bb

		default:
			return false
		}
	})

	m.visibleIndices = indices

	if m.cursor >= len(m.visibleIndices) {
		m.cursor = max(len(m.visibleIndices)-1, 0)
	}
	if m.offset > m.cursor {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height && m.height > 0 {
		m.offset = m.cursor - m.height + 1
	}
}
