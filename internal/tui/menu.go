package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) viewMenu() string {
	return ""
}
