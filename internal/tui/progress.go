package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
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
		return m, tea.Batch(m.syncProgressBar(), waitForDeletionEvent(m.remove.run.progressCh))

	case worktreeDeletedMsg:
		m.remove.run.statuses[msg.Path] = statusRemoved
		m.remove.run.result.SuccessCount++
		m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: true,
		})
		return m, tea.Batch(m.syncProgressBar(), waitForDeletionEvent(m.remove.run.progressCh))

	case worktreeDeleteFailedMsg:
		m.remove.run.statuses[msg.Path] = statusFailed
		m.remove.run.result.FailureCount++
		m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: false,
			Error:   fmt.Errorf("removing %s: %w", msg.Path, msg.Err),
		})
		return m, tea.Batch(m.syncProgressBar(), waitForDeletionEvent(m.remove.run.progressCh))

	case allDeletionsCompleteMsg:
		return m, tea.Batch(m.syncProgressBar(), runPrune(m.runner, m.repoPath))

	case pruneCompleteMsg:
		pruneErr := msg.Err
		m.remove.run.pruneErr = &pruneErr
		return m, tea.Batch(m.syncProgressBar(), runCleanup(m.runner, m.repoPath))

	case cleanupCompleteMsg:
		m.remove.run.cleanupResult = &msg.Result
		m.remove.selected = make(map[string]bool)
		m.worktreeGeneration++
		// Final spring target before the hold: all phases are complete, so
		// the bar settles at full while the completion frame holds.
		syncCmd := m.syncProgressBar()
		updated, holdCmd := m.holdOrAdvance(summaryView)
		return updated, tea.Batch(syncCmd, holdCmd,
			recordRemovals(m.repoPath, m.remove.run.result.SuccessCount),
			loadWorktreeContext(m.runner, m.repoPath, m.worktreeGeneration))

	}
	return m, nil
}

func (m Model) removalLayout() ProgressLayout {
	return ProgressLayout{
		Title:     titleRemoving,
		Completed: m.remove.run.cleanupResult != nil,
		Phases:    m.buildRemovalPhases(),
		Width:     m.width,
		Height:    m.height,
		Hints:     progressFooter,
	}
}

func (m Model) viewProgress() string {
	return m.renderProgressLayout(m.removalLayout())
}

// buildRemovalPhases maps the current removal run onto the shared phase
// shape consumed by ProgressLayout.
func (m Model) buildRemovalPhases() []progress.PhaseState {
	run := m.remove.run
	var phases []progress.PhaseState

	switch {
	case run.teardownRunning:
		phases = append(phases, progress.PhaseState{
			Name:  "Teardown",
			Total: 1,
			Steps: []progress.StepState{{Name: "Removing integration artifacts", Status: progress.StepRunning}},
		})
	case len(run.teardownResults) > 0:
		td := progress.PhaseState{Name: "Teardown", Total: len(run.teardownResults)}
		for _, r := range run.teardownResults {
			td.Steps = append(td.Steps, progress.StepState{Name: r.Name, Status: r.Status})
			switch r.Status {
			case progress.StepDone, progress.StepSkipped:
				td.Done++
			case progress.StepFailed:
				td.Failed++
				td.Done++
			}
		}
		phases = append(phases, td)
	}

	removing := progress.PhaseState{Name: "Removing worktrees", Total: run.total()}
	for _, wt := range run.worktrees {
		label := worktreeLabel(wt)
		var status progress.StepStatus
		switch run.statuses[wt.Path] {
		case statusRemoving:
			status = progress.StepRunning
		case statusRemoved:
			status = progress.StepDone
			removing.Done++
		case statusFailed:
			status = progress.StepFailed
			removing.Failed++
			removing.Done++
		default:
			status = progress.StepPending
		}
		removing.Steps = append(removing.Steps, progress.StepState{Name: label, Status: status})
	}
	phases = append(phases, removing)

	cleanupPhase := progress.PhaseState{Name: "Prune & cleanup"}
	if removing.Total > 0 && removing.Done == removing.Total {
		cleanupPhase.Total = 2
		cleanupPhase.Steps = []progress.StepState{
			{Name: "Prune worktree metadata", Status: progress.StepRunning},
			{Name: "Repository cleanup", Status: progress.StepPending},
		}
		if run.pruneErr != nil {
			if *run.pruneErr != nil {
				cleanupPhase.Steps[0].Status = progress.StepFailed
				cleanupPhase.Failed++
			} else {
				cleanupPhase.Steps[0].Status = progress.StepDone
			}
			cleanupPhase.Done++
			cleanupPhase.Steps[1].Status = progress.StepRunning
		}
		if run.cleanupResult != nil {
			cleanupPhase.Steps[1].Status = progress.StepDone
			cleanupPhase.Done++
		}
	}
	phases = append(phases, cleanupPhase)
	return phases
}
