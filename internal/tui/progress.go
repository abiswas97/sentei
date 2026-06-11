package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/pipeline"
	"github.com/abiswas97/sentei/internal/worktree"
)

const (
	statusPending  = "pending"
	statusRemoving = "removing"
	statusRemoved  = "removed"
	statusFailed   = "failed"
)

type cleanupCompleteMsg struct {
	Result cleanup.Result
}

type worktreeDeleteStartedMsg struct{ Path string }
type worktreeDeletedMsg struct{ Path string }
type worktreeDeleteFailedMsg struct {
	Path string
	Err  error
}
type allDeletionsCompleteMsg struct{}
type pruneCompleteMsg struct{ Err error }

func waitForDeletionEvent(ch <-chan worktree.DeletionEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return allDeletionsCompleteMsg{}
		}
		switch ev.Type {
		case worktree.DeletionStarted:
			return worktreeDeleteStartedMsg{Path: ev.Path}
		case worktree.DeletionCompleted:
			return worktreeDeletedMsg{Path: ev.Path}
		case worktree.DeletionFailed:
			return worktreeDeleteFailedMsg{Path: ev.Path, Err: ev.Error}
		default:
			return waitForDeletionEvent(ch)()
		}
	}
}

func runPrune(runner git.CommandRunner, repoPath string) tea.Cmd {
	return func() tea.Msg {
		err := worktree.PruneWorktrees(runner, repoPath)
		return pruneCompleteMsg{Err: err}
	}
}

func runCleanup(runner git.CommandRunner, repoPath string) tea.Cmd {
	return func() tea.Msg {
		result := cleanup.Run(runner, repoPath, cleanup.Options{Mode: cleanup.ModeSafe}, func(cleanup.Event) {})
		return cleanupCompleteMsg{Result: result}
	}
}

func (m Model) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		return m, nil

	case teardownCompleteMsg:
		// Yes switches to progressView before teardown runs, so completion
		// lands here, not in updateConfirm.
		m.remove.run.teardownRunning = false
		m.remove.run.teardownResults = msg.results
		return m.startDeletions()

	case worktreeDeleteStartedMsg:
		m.remove.run.statuses[msg.Path] = statusRemoving
		return m, waitForDeletionEvent(m.remove.run.progressCh)

	case worktreeDeletedMsg:
		m.remove.run.statuses[msg.Path] = statusRemoved
		m.remove.run.result.SuccessCount++
		m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: true,
		})
		return m, waitForDeletionEvent(m.remove.run.progressCh)

	case worktreeDeleteFailedMsg:
		m.remove.run.statuses[msg.Path] = statusFailed
		m.remove.run.result.FailureCount++
		m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: false,
			Error:   fmt.Errorf("removing %s: %w", msg.Path, msg.Err),
		})
		return m, waitForDeletionEvent(m.remove.run.progressCh)

	case allDeletionsCompleteMsg:
		return m, runPrune(m.runner, m.repoPath)

	case pruneCompleteMsg:
		pruneErr := msg.Err
		m.remove.run.pruneErr = &pruneErr
		return m, runCleanup(m.runner, m.repoPath)

	case cleanupCompleteMsg:
		m.remove.run.cleanupResult = &msg.Result
		m.remove.selected = make(map[string]bool)
		m.worktreeGeneration++
		updated, holdCmd := m.holdOrAdvance(summaryView)
		return updated, tea.Batch(holdCmd, loadWorktreeContext(m.runner, m.repoPath, m.worktreeGeneration))
	}
	return m, nil
}

func (m Model) viewProgress() string {
	return ProgressLayout{
		Title:  "Removing Worktrees",
		Phases: m.buildRemovalPhases(),
		Width:  m.width,
		Height: m.height,
		Hints:  progressFooter,
	}.View()
}

// buildRemovalPhases maps the current removal run onto the shared phase
// shape consumed by ProgressLayout.
func (m Model) buildRemovalPhases() []phaseDisplay {
	run := m.remove.run
	var phases []phaseDisplay

	switch {
	case run.teardownRunning:
		phases = append(phases, phaseDisplay{
			name:  "Teardown",
			total: 1,
			steps: []stepDisplay{{name: "Removing integration artifacts", status: pipeline.StepRunning}},
		})
	case len(run.teardownResults) > 0:
		td := phaseDisplay{name: "Teardown", total: len(run.teardownResults)}
		for _, r := range run.teardownResults {
			td.steps = append(td.steps, stepDisplay{name: r.Name, status: r.Status})
			switch r.Status {
			case pipeline.StepDone, pipeline.StepSkipped:
				td.done++
			case pipeline.StepFailed:
				td.failed++
				td.done++
			}
		}
		phases = append(phases, td)
	}

	removing := phaseDisplay{name: "Removing worktrees", total: run.total()}
	for _, wt := range run.worktrees {
		label := worktreeLabel(wt)
		var status pipeline.StepStatus
		switch run.statuses[wt.Path] {
		case statusRemoving:
			status = pipeline.StepRunning
		case statusRemoved:
			status = pipeline.StepDone
			removing.done++
		case statusFailed:
			status = pipeline.StepFailed
			removing.failed++
			removing.done++
		default:
			status = pipeline.StepPending
		}
		removing.steps = append(removing.steps, stepDisplay{name: label, status: status})
	}
	phases = append(phases, removing)

	cleanupPhase := phaseDisplay{name: "Prune & cleanup"}
	if removing.total > 0 && removing.done == removing.total {
		cleanupPhase.total = 2
		cleanupPhase.steps = []stepDisplay{
			{name: "Prune worktree metadata", status: pipeline.StepRunning},
			{name: "Repository cleanup", status: pipeline.StepPending},
		}
		if run.pruneErr != nil {
			if *run.pruneErr != nil {
				cleanupPhase.steps[0].status = pipeline.StepFailed
				cleanupPhase.failed++
			} else {
				cleanupPhase.steps[0].status = pipeline.StepDone
			}
			cleanupPhase.done++
			cleanupPhase.steps[1].status = pipeline.StepRunning
		}
		if run.cleanupResult != nil {
			cleanupPhase.steps[1].status = pipeline.StepDone
			cleanupPhase.done++
		}
	}
	phases = append(phases, cleanupPhase)
	return phases
}
