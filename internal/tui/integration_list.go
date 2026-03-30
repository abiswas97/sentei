package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateIntegrationList(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) viewIntegrationList() string {
	return "  Integration list (todo)\n"
}
