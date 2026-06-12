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

type removalEventMsg struct{ event progress.Event }
type allDeletionsCompleteMsg struct{}
type pruneCompleteMsg struct{ Err error }

func waitForRemovalEvent(ch <-chan progress.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return allDeletionsCompleteMsg{}
		}
		return removalEventMsg{event: ev}
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

	case removalEventMsg:
		path := msg.event.Step
		switch msg.event.Status {
		case progress.StepRunning:
			m.remove.run.statuses[path] = statusRemoving
		case progress.StepDone:
			m.remove.run.statuses[path] = statusRemoved
			m.remove.run.result.SuccessCount++
			m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
				Path:    path,
				Success: true,
			})
		case progress.StepFailed:
			m.remove.run.statuses[path] = statusFailed
			m.remove.run.result.FailureCount++
			m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
				Path:    path,
				Success: false,
				Error:   fmt.Errorf("removing %s: %w", path, msg.event.Error),
			})
		}
		return m, tea.Batch(m.syncProgressBar(), waitForRemovalEvent(m.remove.run.progressCh))

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
		// The plan was scanned at confirm time: the phase shows its real
		// total from its first frame; steps resolve when results land.
		td := progress.PhaseState{Name: "Teardown", Total: len(run.teardownPlanned), Closed: true}
		for _, name := range run.teardownPlanned {
			td.Steps = append(td.Steps, progress.StepState{Name: name, Status: progress.StepPending, Declared: 1})
		}
		phases = append(phases, td)
	case len(run.teardownResults) > 0:
		td := progress.PhaseState{Name: "Teardown", Total: len(run.teardownResults), Closed: true}
		for _, r := range run.teardownResults {
			td.Steps = append(td.Steps, progress.StepState{Name: r.Name, Status: r.Status, Reached: 1, Declared: 1})
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

	// Each removal step declares two checkpoints (started, removed): the
	// deleter observes nothing finer inside `git worktree remove`, and the
	// start checkpoint is what moves the bar during parallel work.
	removing := progress.PhaseState{Name: worktree.RemovalPhaseName, Total: run.total(), Closed: true}
	for _, wt := range run.worktrees {
		label := worktreeLabel(wt)
		step := progress.StepState{Name: label, Declared: 2}
		switch run.statuses[wt.Path] {
		case statusRemoving:
			step.Status = progress.StepRunning
			step.Reached = 1
		case statusRemoved:
			step.Status = progress.StepDone
			step.Reached = 2
			removing.Done++
		case statusFailed:
			step.Status = progress.StepFailed
			step.Reached = 2
			removing.Failed++
			removing.Done++
		default:
			step.Status = progress.StepPending
		}
		removing.Steps = append(removing.Steps, step)
	}
	phases = append(phases, removing)

	cleanupPhase := progress.PhaseState{Name: "Prune & cleanup", Closed: true}
	if removing.Total > 0 && removing.Done == removing.Total {
		cleanupPhase.Total = 2
		cleanupPhase.Steps = []progress.StepState{
			{Name: "Prune worktree metadata", Status: progress.StepRunning, Declared: 1},
			{Name: "Repository cleanup", Status: progress.StepPending, Declared: 1},
		}
		if run.pruneErr != nil {
			if *run.pruneErr != nil {
				cleanupPhase.Steps[0].Status = progress.StepFailed
				cleanupPhase.Failed++
			} else {
				cleanupPhase.Steps[0].Status = progress.StepDone
			}
			cleanupPhase.Steps[0].Reached = 1
			cleanupPhase.Done++
			cleanupPhase.Steps[1].Status = progress.StepRunning
		}
		if run.cleanupResult != nil {
			cleanupPhase.Steps[1].Status = progress.StepDone
			cleanupPhase.Steps[1].Reached = 1
			cleanupPhase.Done++
		}
	}
	phases = append(phases, cleanupPhase)
	return phases
}
