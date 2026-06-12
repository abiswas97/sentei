package tui

import (
	"maps"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/state"
)

type integrationFinalizedMsg struct {
	err error
}

func (m Model) updateIntegrationProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		return m, nil

	case integrationEventMsg:
		m.integ.events = append(m.integ.events, msg.Event)
		return m, tea.Batch(m.syncProgressBar(), waitForIntegrationEvent(m.integ.eventCh, m.integ.doneCh))

	case integrationApplyDoneMsg:
		return m, m.finalizeIntegrationApply()

	case integrationFinalizedMsg:
		m.integ.finalized = true
		finalSync := m.syncProgressBar()
		// The migrate flow has its own summary; hand off unchanged.
		if m.integ.returnView == migrateNextView {
			updated, holdCmd := m.holdOrAdvance(migrateNextView)
			return updated, tea.Batch(finalSync, holdCmd)
		}
		// In-memory current/staged are never mutated here: dismissing the
		// summary reloads them from persisted state, so the list always
		// matches disk whether the save succeeded or failed.
		m.integ.saveErr = msg.err
		if msg.err == nil {
			m.worktreeGeneration++
			updated, holdCmd := m.holdOrAdvance(integrationSummaryView)
			return updated, tea.Batch(finalSync, holdCmd, loadWorktreeContext(m.runner, m.repoPath, m.worktreeGeneration))
		}
		updated, holdCmd := m.holdOrAdvance(integrationSummaryView)
		return updated, tea.Batch(finalSync, holdCmd)
	}
	return m, nil
}

func (m Model) finalizeIntegrationApply() tea.Cmd {
	runner := m.runner
	repoPath := m.repoPath
	returnView := m.integ.returnView
	integrations := m.integ.integrations
	staged := make(map[string]bool)
	maps.Copy(staged, m.integ.staged)
	repoResult := m.repo.result

	return func() tea.Msg {
		root := repoPath
		if returnView == migrateNextView {
			if result, ok := repoResult.(repo.MigrateResult); ok {
				root = result.BareRoot
			}
		}
		bareDir, err := git.CommonDir(runner, root)
		if err != nil {
			return integrationFinalizedMsg{err: err}
		}

		var enabled []string
		for _, integ := range integrations {
			if staged[integ.Name] {
				enabled = append(enabled, integ.Name)
			}
		}
		err = state.Save(bareDir, &state.State{Integrations: enabled})

		return integrationFinalizedMsg{err: err}
	}
}

func (m Model) integrationLayout() ProgressLayout {
	done, total := m.integrationOverallProgress()
	return ProgressLayout{
		Title:        titleApplyingChanges,
		Phases:       m.buildIntegrationPhases(),
		Width:        m.width,
		Height:       m.height,
		Hints:        progressFooter,
		OverallDone:  done,
		OverallTotal: total,
		Completed:    m.integ.finalized,
	}
}

func (m Model) viewIntegrationProgress() string {
	return m.renderProgressLayout(m.integrationLayout())
}

// buildIntegrationPhases maps apply events onto the shared phase shape, one
// phase per worktree, with every target worktree visible as pending before
// its events arrive. Failed step errors are baked into the step label.
func (m Model) buildIntegrationPhases() []progress.PhaseState {
	var phases []progress.PhaseState
	seen := make(map[string]bool)

	for _, g := range groupIntegrationEvents(m.integ.events) {
		pd := progress.PhaseState{Name: filepath.Base(g.worktree)}
		seen[g.worktree] = true
		for _, s := range g.steps {
			if s.ev.Status == integration.StatusSkipped {
				continue // skipped steps stay hidden, as before
			}
			label := s.step
			if s.ev.Error != nil {
				// One-line rows law: live progress shows only the error's
				// final line; the summary's peek and portal carry the rest.
				label += " " + errorPeekLast(s.ev.Error.Error(), max(m.width-10, 20))
			}
			var status progress.StepStatus
			switch s.ev.Status {
			case integration.StatusDone:
				status = progress.StepDone
				pd.Done++
			case integration.StatusRunning:
				status = progress.StepRunning
			case integration.StatusFailed:
				status = progress.StepFailed
				pd.Failed++
				pd.Done++
			}
			pd.Steps = append(pd.Steps, progress.StepState{Name: label, Status: status})
		}
		pd.Total = len(pd.Steps)
		phases = append(phases, pd)
	}

	for _, path := range m.integ.targetWorktrees {
		if !seen[path] {
			phases = append(phases, progress.PhaseState{Name: filepath.Base(path)})
		}
	}
	return phases
}

// integrationOverallProgress counts resolved unique steps against the
// upfront step total so the bar reflects the whole apply, not just the
// worktrees that have emitted events.
func (m Model) integrationOverallProgress() (done, total int) {
	total = m.integ.totalSteps
	resolved := make(map[string]bool)
	for _, ev := range m.integ.events {
		key := ev.Worktree + ":" + ev.Step
		if resolved[key] {
			continue
		}
		if ev.Status == integration.StatusDone || ev.Status == integration.StatusFailed || ev.Status == integration.StatusSkipped {
			resolved[key] = true
			done++
		}
	}
	if total == 0 {
		total = done
	}
	return done, total
}
