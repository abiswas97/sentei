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
	"github.com/abiswas97/sentei/internal/repo"
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
	repoNameView
	repoOptionsView
	repoProgressView
	repoSummaryView
	cloneInputView
	migrateConfirmView
	migrateProgressView
	migrateSummaryView
	migrateNextView
	integrationListView
	integrationProgressView
	migrateIntegrationsView
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

	teardownResults []creator.StepResult

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
	branchInput   textinput.Model
	baseInput     textinput.Model
	focusedField  int // 0 = branch, 1 = base
	validationErr string

	ecosystems             []config.EcosystemConfig
	ecoEnabled             map[string]bool
	activeIntegrationNames []string // loaded from state, displayed as info line
	mergeBase              bool
	copyEnvFiles           bool
	optionsCursor          int

	eventCh  chan creator.Event
	resultCh chan creator.Result
	events   []creator.Event
	result   *creator.Result
}

// MigrateInfo holds pre-loaded info about the repo being migrated.
type MigrateInfo struct {
	Branch  string
	IsDirty bool
}

// repoState holds all state for repo create/clone/migrate flows.
type repoState struct {
	// Create repo fields
	nameInput     textinput.Model
	locationInput textinput.Model
	focusedField  int // 0 = name, 1 = location
	validationErr string

	// Options
	createWorktree bool
	publishGitHub  bool
	visibility     string // "private" or "public"
	descInput      textinput.Model
	ghStatus       string // "authenticated", "not authenticated", "gh not found"
	optionsCursor  int

	// Clone fields
	urlInput           textinput.Model
	cloneNameInput     textinput.Model
	cloneFocusedField  int // 0 = url, 1 = name
	nameManuallyEdited bool

	// Migrate fields
	migrateInfo MigrateInfo

	// Shared progress/summary
	eventCh  chan repo.Event
	resultCh chan interface{} // receives CreateResult, CloneResult, or MigrateResult
	events   []repo.Event
	result   interface{}
	opType   string // "create", "clone", "migrate"
}

// integrationState holds all state for the integration management flow.
type integrationState struct {
	integrations []integration.Integration //nolint:unused
	current      map[string]bool           // what's on disk right now
	staged       map[string]bool           // desired state after apply
	detected     map[string]bool           // what was detected on disk at load time (for "detected" hints)
	depStatus    map[string]bool           // dep name → installed (for info dialog) //nolint:unused
	cursor       int                       //nolint:unused
	colCursor    int                       // 0-based column index for future expansion //nolint:unused

	// Info dialog
	showInfo   bool //nolint:unused
	infoCursor int  // which integration is shown in the carousel //nolint:unused

	// Progress
	events  []integration.ManagerEvent    //nolint:unused
	eventCh chan integration.ManagerEvent //nolint:unused
	doneCh  chan struct{}                 //nolint:unused

	// Context: where to return after progress completes
	returnView viewState //nolint:unused
}

type Model struct {
	view     viewState
	runner   git.CommandRunner
	shell    git.ShellRunner
	repoPath string
	cfg      *config.Config
	context  repo.RepoContext
	width    int
	height   int

	menuItems  []menuItem
	menuCursor int

	remove removeState
	create createState
	repo   repoState
	integ  integrationState
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

func NewMenuModel(runner git.CommandRunner, shell git.ShellRunner, repoPath string, cfg *config.Config, context repo.RepoContext) Model {
	branchInput := textinput.New()
	branchInput.Placeholder = "feature/my-branch"
	branchInput.Focus()

	baseInput := textinput.New()
	baseInput.Placeholder = "main"
	baseInput.SetValue("main")

	filterInput := textinput.New()
	filterInput.Prompt = "filter: "

	nameInput := textinput.New()
	nameInput.Placeholder = "my-project"

	locationInput := textinput.New()
	locationInput.SetValue(repoPath)
	locationInput.Placeholder = repoPath

	descInput := textinput.New()
	descInput.Placeholder = "optional description"

	urlInput := textinput.New()
	urlInput.Placeholder = "git@github.com:user/repo.git"

	cloneNameInput := textinput.New()
	cloneNameInput.Placeholder = "repo"

	var items []menuItem
	switch context {
	case repo.ContextBareRepo:
		items = []menuItem{
			{label: "Create new worktree", enabled: true},
			{label: "Manage integrations", enabled: true},
			{label: "Remove worktrees", hint: "loading\u2026", enabled: false},
			{label: "Cleanup & exit", hint: "prune refs, remove gone branches", enabled: true},
		}
	case repo.ContextNoRepo:
		items = []menuItem{
			{label: "Create new repository", enabled: true},
			{label: "Clone repository as bare", enabled: true},
		}
	case repo.ContextNonBareRepo:
		items = []menuItem{
			{label: "Migrate to bare repository", enabled: true},
			{label: "Clone repository as bare", enabled: true},
			{label: "Create new repository", enabled: true},
		}
	}

	m := Model{
		view:      menuView,
		runner:    runner,
		shell:     shell,
		repoPath:  repoPath,
		cfg:       cfg,
		context:   context,
		height:    20,
		menuItems: items,
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
			mergeBase:    true,
			copyEnvFiles: true,
		},
		repo: repoState{
			nameInput:      nameInput,
			locationInput:  locationInput,
			descInput:      descInput,
			urlInput:       urlInput,
			cloneNameInput: cloneNameInput,
			visibility:     "private",
		},
		integ: integrationState{
			current:  make(map[string]bool),
			staged:   make(map[string]bool),
			detected: make(map[string]bool),
		},
	}

	return m
}

func (m Model) Init() tea.Cmd {
	if m.view == menuView && m.context == repo.ContextBareRepo {
		return loadWorktreeContext(m.runner, m.repoPath)
	}
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
	case repoNameView:
		return m.updateRepoName(msg)
	case repoOptionsView:
		return m.updateRepoOptions(msg)
	case repoProgressView, migrateProgressView:
		return m.updateRepoProgress(msg)
	case repoSummaryView:
		return m.updateRepoSummary(msg)
	case cloneInputView:
		return m.updateCloneInput(msg)
	case migrateConfirmView:
		return m.updateMigrateConfirm(msg)
	case migrateSummaryView:
		return m.updateMigrateSummary(msg)
	case migrateNextView:
		return m.updateMigrateNext(msg)
	case integrationListView:
		return m.updateIntegrationList(msg)
	case integrationProgressView:
		return m.updateIntegrationProgress(msg)
	case migrateIntegrationsView:
		return m.updateMigrateIntegrations(msg)
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
	case repoNameView:
		return m.viewRepoName()
	case repoOptionsView:
		return m.viewRepoOptions()
	case repoProgressView, migrateProgressView:
		return m.viewRepoProgress()
	case repoSummaryView:
		return m.viewRepoSummary()
	case cloneInputView:
		return m.viewCloneInput()
	case migrateConfirmView:
		return m.viewMigrateConfirm()
	case migrateSummaryView:
		return m.viewMigrateSummary()
	case migrateNextView:
		return m.viewMigrateNext()
	case integrationListView:
		return m.viewIntegrationList()
	case integrationProgressView:
		return m.viewIntegrationProgress()
	case migrateIntegrationsView:
		return m.viewMigrateIntegrations()
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
