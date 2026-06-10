package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateCreateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
