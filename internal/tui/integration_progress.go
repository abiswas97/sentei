package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateIntegrationProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) viewIntegrationProgress() string {
	return "  Integration progress (todo)\n"
}
