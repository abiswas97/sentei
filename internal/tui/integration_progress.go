package tui

import (
	"maps"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/state"
)

type integrationFinalizedMsg struct {
	err error
}

func (m Model) updateIntegrationProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case integrationPreparedMsg:
		if msg.err != nil {
			m.integ.prepareErr = msg.err
			m.integ.lifecycle = integrationSettling
			updated, holdCmd := m.holdOrAdvance(integrationSummaryView)
			return updated, holdCmd
		}
		return m.startPreparedIntegrationApply(msg.prepared)

	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		return m, nil

	case integrationEventMsg:
		m.integ.events = append(m.integ.events, msg.Event)
		return m, tea.Batch(m.syncProgressBar(), waitForIntegrationEvent(m.integ.eventCh, m.integ.resultCh))

	case integrationApplyDoneMsg:
		if msg.result.err != nil {
			m.integ.executionErr = msg.result.err
			m.integ.lifecycle = integrationSettling
			finalSync := m.syncProgressBar()
			updated, holdCmd := m.holdOrAdvance(integrationSummaryView)
			return updated, tea.Batch(finalSync, holdCmd)
		}
		if !integrationExecutionCanSave(msg.result.phases, msg.result.empty) {
			m.integ.lifecycle = integrationSettling
			updated, holdCmd := m.holdOrAdvance(integrationSummaryView)
			return updated, tea.Batch(m.syncProgressBar(), holdCmd)
		}
		m.integ.lifecycle = integrationSaving
		return m, m.finalizeIntegrationApply()

	case integrationFinalizedMsg:
		m.integ.lifecycle = integrationSettling
		m.integ.saveErr = msg.err
		finalSync := m.syncProgressBar()
		// Clean migration applies hand off directly. Errors use the integration
		// summary first so their failure cannot disappear between flows.
		if m.integ.returnView == migrateNextView && msg.err == nil {
			updated, holdCmd := m.holdOrAdvance(migrateNextView)
			return updated, tea.Batch(finalSync, holdCmd)
		}
		// In-memory current/staged are never mutated here: dismissing the
		// summary reloads them from persisted state, so the list always
		// matches disk whether the save succeeded or failed.
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
		persisted, err := state.Load(bareDir)
		if err != nil {
			return integrationFinalizedMsg{err: err}
		}
		persisted.Integrations = enabled
		err = state.Save(bareDir, persisted)

		return integrationFinalizedMsg{err: err}
	}
}

func (m Model) integrationLayout() ProgressLayout {
	return ProgressLayout{
		Title:     titleApplyingChanges,
		Phases:    m.buildIntegrationPhases(),
		Width:     m.width,
		Height:    m.height,
		Hints:     progressFooter,
		Completed: m.integ.lifecycle == integrationSettling,
	}
}

func (m Model) viewIntegrationProgress() string {
	if m.integ.lifecycle == integrationPreparing {
		var b strings.Builder
		b.WriteString(viewTitle(titleApplyingChanges))
		b.WriteString("\n\n")
		b.WriteString(viewSeparator(m.width))
		b.WriteString("\n\n  ")
		b.WriteString(shimmerLine(starFrame(m.motionTick)+" Preparing plan...", rampAccent, m.motionTick))
		b.WriteString("\n")
		return b.String()
	}
	return m.renderProgressLayout(m.integrationLayout())
}

func integrationExecutionCanSave(phases []progress.Phase, empty bool) bool {
	if empty {
		return len(phases) == 0
	}
	sawStep := false
	for _, phase := range phases {
		for _, step := range phase.Steps {
			sawStep = true
			if step.Status != progress.StepDone {
				return false
			}
		}
	}
	return sawStep
}

func (m Model) buildIntegrationPhases() []progress.PhaseState {
	states := progress.Snapshot(m.integ.events)
	for phaseIndex := range states {
		states[phaseIndex].Name = filepath.Base(states[phaseIndex].Name)
		for stepIndex := range states[phaseIndex].Steps {
			step := &states[phaseIndex].Steps[stepIndex]
			if step.Error != nil {
				step.Name += " " + errorPeekLast(step.Error.Error(), max(m.width-10, 20))
			}
		}
	}
	return states
}
