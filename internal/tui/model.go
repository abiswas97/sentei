package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/worktree"
)

type viewState int

const (
	menuView viewState = iota
	listView
	confirmView
	progressView
	summaryView
	createBranchView
	createOptionsView
	createProgressView
	createSummaryView
)

type SortField int

const (
	SortByAge SortField = iota
	SortByBranch
)

// removeState holds all state for the worktree removal flow.
type removeState struct {
	worktrees      []git.Worktree
	selected       map[string]bool
	visibleIndices []int
	cursor         int
	offset         int

	sortField     SortField
	sortAscending bool

	filterText   string
	filterActive bool
	filterInput  textinput.Model

	deletionStatuses map[string]string
	deletionResult   worktree.DeletionResult
	deletionTotal    int
	progressCh       <-chan worktree.DeletionEvent

	pruneErr      *error
	cleanupResult *cleanup.Result
}

// menuItem represents a selectable menu entry.
type menuItem struct {
	label   string
	hint    string
	enabled bool
}

// createState holds all state for the worktree creation flow.
type createState struct {
	branchInput  textinput.Model
	baseInput    textinput.Model
	focusedField int // 0 = branch, 1 = base

	ecosystems    []config.EcosystemConfig
	integrations  []integration.Integration
	ecoEnabled    map[string]bool
	intEnabled    map[string]bool
	mergeBase     bool
	copyEnvFiles  bool
	optionsCursor int

	eventCh chan creator.Event
}

type Model struct {
	view     viewState
	runner   git.CommandRunner
	repoPath string
	cfg      *config.Config
	width    int
	height   int

	menuItems  []menuItem
	menuCursor int

	remove removeState
	create createState
}

func NewModel(worktrees []git.Worktree, runner git.CommandRunner, repoPath string) Model {
	ti := textinput.New()
	ti.Prompt = "filter: "

	m := Model{
		view:     listView,
		runner:   runner,
		repoPath: repoPath,
		remove: removeState{
			worktrees:        worktrees,
			selected:         make(map[string]bool),
			sortField:        SortByAge,
			sortAscending:    true,
			filterInput:      ti,
			deletionStatuses: make(map[string]string),
		},
		height: 20,
	}
	m.reindex()
	return m
}

func NewMenuModel(runner git.CommandRunner, repoPath string, cfg *config.Config) Model {
	branchInput := textinput.New()
	branchInput.Placeholder = "feature/my-branch"
	branchInput.Focus()

	baseInput := textinput.New()
	baseInput.Placeholder = "main"
	baseInput.SetValue("main")

	filterInput := textinput.New()
	filterInput.Prompt = "filter: "

	m := Model{
		view:     menuView,
		runner:   runner,
		repoPath: repoPath,
		cfg:      cfg,
		height:   20,
		menuItems: []menuItem{
			{label: "Create new worktree", enabled: true},
			{label: "Remove worktrees", enabled: true},
			{label: "Cleanup", hint: "safe mode", enabled: true},
		},
		remove: removeState{
			selected:         make(map[string]bool),
			sortField:        SortByAge,
			sortAscending:    true,
			filterInput:      filterInput,
			deletionStatuses: make(map[string]string),
		},
		create: createState{
			branchInput:  branchInput,
			baseInput:    baseInput,
			ecoEnabled:   make(map[string]bool),
			intEnabled:   make(map[string]bool),
			mergeBase:    true,
			copyEnvFiles: true,
		},
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.view {
	case menuView:
		return m.updateMenu(msg)
	case listView:
		return m.updateList(msg)
	case confirmView:
		return m.updateConfirm(msg)
	case progressView:
		return m.updateProgress(msg)
	case summaryView:
		return m.updateSummary(msg)
	case createBranchView:
		return m.updateCreateBranch(msg)
	case createOptionsView:
		return m.updateCreateOptions(msg)
	case createProgressView:
		return m.updateCreateProgress(msg)
	case createSummaryView:
		return m.updateCreateSummary(msg)
	}
	return m, nil
}

func (m Model) View() string {
	switch m.view {
	case menuView:
		return m.viewMenu()
	case listView:
		return m.viewList()
	case confirmView:
		return m.viewConfirm()
	case progressView:
		return m.viewProgress()
	case summaryView:
		return m.viewSummary()
	case createBranchView:
		return m.viewCreateBranch()
	case createOptionsView:
		return m.viewCreateOptions()
	case createProgressView:
		return m.viewCreateProgress()
	case createSummaryView:
		return m.viewCreateSummary()
	}
	return ""
}

func (m Model) selectedWorktrees() []git.Worktree {
	var result []git.Worktree
	for _, wt := range m.remove.worktrees {
		if m.remove.selected[wt.Path] {
			result = append(result, wt)
		}
	}
	return result
}

func (m *Model) reindex() {
	filterLower := strings.ToLower(m.remove.filterText)

	var indices []int
	for i, wt := range m.remove.worktrees {
		if filterLower != "" {
			branch := strings.ToLower(stripBranchPrefix(wt.Branch))
			if !strings.Contains(branch, filterLower) {
				continue
			}
		}
		indices = append(indices, i)
	}

	sortAsc := m.remove.sortAscending
	sortField := m.remove.sortField
	wts := m.remove.worktrees

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

	m.remove.visibleIndices = indices

	if m.remove.cursor >= len(m.remove.visibleIndices) {
		m.remove.cursor = max(len(m.remove.visibleIndices)-1, 0)
	}
	if m.remove.offset > m.remove.cursor {
		m.remove.offset = m.remove.cursor
	}
	if m.remove.cursor >= m.remove.offset+m.height && m.height > 0 {
		m.remove.offset = m.remove.cursor - m.height + 1
	}
}
