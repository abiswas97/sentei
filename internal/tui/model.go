package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas/wt-sweep/internal/git"
	"github.com/abiswas/wt-sweep/internal/worktree"
)

type viewState int

const (
	listView viewState = iota
	confirmView
	progressView
	summaryView
)

type Model struct {
	worktrees []git.Worktree
	selected  map[int]bool
	cursor    int
	offset    int
	height    int

	view viewState

	runner   git.CommandRunner
	repoPath string

	deletionStatuses map[string]string
	deletionResult   worktree.DeletionResult
	deletionTotal    int
	deletionDone     int
	progressCh       <-chan worktree.DeletionEvent
}

func NewModel(worktrees []git.Worktree, runner git.CommandRunner, repoPath string) Model {
	return Model{
		worktrees:        worktrees,
		selected:         make(map[int]bool),
		runner:           runner,
		repoPath:         repoPath,
		deletionStatuses: make(map[string]string),
		height:           20,
	}
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
	for i, wt := range m.worktrees {
		if m.selected[i] {
			result = append(result, wt)
		}
	}
	return result
}
