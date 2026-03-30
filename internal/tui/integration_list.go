package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/state"
)

type integrationStateLoadedMsg struct {
	integrations []integration.Integration
	current      map[string]bool
	enabled      []string
	err          error
}

func (m Model) loadIntegrationState() tea.Cmd {
	return func() tea.Msg {
		all := integration.All()
		mainWT := m.findSourceWorktree()
		current := make(map[string]bool)
		if mainWT != "" {
			current = integration.DetectAllPresent(mainWT, all)
		}

		bareDir := filepath.Join(m.repoPath, ".bare")
		st, err := state.Load(bareDir)
		if err != nil {
			return integrationStateLoadedMsg{
				integrations: all,
				current:      current,
				err:          err,
			}
		}

		return integrationStateLoadedMsg{
			integrations: all,
			current:      current,
			enabled:      st.Integrations,
		}
	}
}

func (m Model) updateIntegrationList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case integrationStateLoadedMsg:
		m.integ.integrations = msg.integrations
		m.integ.current = msg.current
		m.integ.staged = make(map[string]bool)
		for _, integ := range msg.integrations {
			m.integ.staged[integ.Name] = msg.current[integ.Name]
		}
		// If state file loaded successfully and had entries, overlay those as source of truth
		if msg.err == nil && len(msg.enabled) > 0 {
			for _, integ := range msg.integrations {
				m.integ.staged[integ.Name] = false
			}
			for _, name := range msg.enabled {
				m.integ.staged[name] = true
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			for _, integ := range m.integ.integrations {
				m.integ.staged[integ.Name] = m.integ.current[integ.Name]
			}
			m.view = menuView
			return m, nil

		case key.Matches(msg, keys.Quit):
			m.view = menuView
			return m, nil
		}
	}
	return m, nil
}

func (m Model) viewIntegrationList() string {
	return "  Integration list (loading...)\n"
}
