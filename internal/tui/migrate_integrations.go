package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateMigrateIntegrations(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) viewMigrateIntegrations() string {
	return "  Migrate integrations (todo)\n"
}
