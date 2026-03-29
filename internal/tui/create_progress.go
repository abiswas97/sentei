package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateCreateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) viewCreateProgress() string {
	return ""
}
