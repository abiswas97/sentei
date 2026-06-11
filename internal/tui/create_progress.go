package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m Model) updateCreateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case createEventMsg:
		m.create.events = append(m.create.events, msg.Event)
		return m, m.waitForCreateEvent()

	case createCompleteMsg:
		m.create.result = &msg.Result
		m.worktreeGeneration++
		updated, holdCmd := m.holdOrAdvance(createSummaryView)
		return updated, tea.Batch(holdCmd, loadWorktreeContext(m.runner, m.repoPath, m.worktreeGeneration))
	}
	return m, nil
}

func (m Model) viewCreateProgress() string {
	return ProgressLayout{
		Title:    "Creating Worktree",
		Subtitle: fmt.Sprintf("%s \u2192 from %s", m.create.branchInput.Value(), m.create.baseInput.Value()),
		Phases:   withPendingPhases(buildPhaseDisplays(m.create.events), "Setup", "Dependencies", "Integrations"),
		Width:    m.width,
		Height:   m.height,
		Hints:    []KeyHint{{"q", "quit"}},
	}.View()
}
