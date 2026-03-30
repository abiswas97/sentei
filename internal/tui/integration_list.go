package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/state"
)

type integrationStateLoadedMsg struct {
	integrations []integration.Integration
	current      map[string]bool
	enabled      []string
	err          error
}

type integrationEventMsg struct {
	Event integration.ManagerEvent
}

type integrationApplyDoneMsg struct{}

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

func waitForIntegrationEvent(ch <-chan integration.ManagerEvent, doneCh <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			<-doneCh
			return integrationApplyDoneMsg{}
		}
		return integrationEventMsg{Event: ev}
	}
}

func (m Model) integrationHasPendingChanges() bool {
	for _, integ := range m.integ.integrations {
		if m.integ.staged[integ.Name] != m.integ.current[integ.Name] {
			return true
		}
	}
	return false
}

func (m Model) pendingChangeCount() int {
	count := 0
	for _, integ := range m.integ.integrations {
		if m.integ.staged[integ.Name] != m.integ.current[integ.Name] {
			count++
		}
	}
	return count
}

func (m Model) startIntegrationApply() (Model, tea.Cmd) {
	var toEnable, toDisable []integration.Integration
	for _, integ := range m.integ.integrations {
		staged := m.integ.staged[integ.Name]
		current := m.integ.current[integ.Name]
		if staged && !current {
			toEnable = append(toEnable, integ)
		} else if !staged && current {
			toDisable = append(toDisable, integ)
		}
	}

	var wtPaths []string
	for _, wt := range m.remove.worktrees {
		wtPaths = append(wtPaths, wt.Path)
	}

	ch := make(chan integration.ManagerEvent, 50)
	doneCh := make(chan struct{}, 1)
	m.integ.eventCh = ch
	m.integ.doneCh = doneCh

	repoPath := m.repoPath
	shell := m.shell
	mainWT := m.findSourceWorktree()

	go func() {
		emit := func(e integration.ManagerEvent) { ch <- e }
		for _, integ := range toEnable {
			integration.EnableIntegration(shell, repoPath, mainWT, wtPaths, integ, emit)
		}
		for _, integ := range toDisable {
			integration.DisableIntegration(shell, wtPaths, integ, emit)
		}
		close(ch)
		doneCh <- struct{}{}
	}()

	return m, waitForIntegrationEvent(ch, doneCh)
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
		if m.integ.showInfo {
			return m.updateIntegrationInfo(msg)
		}

		switch {
		case key.Matches(msg, keys.IntDown):
			if len(m.integ.integrations) > 0 && m.integ.cursor < len(m.integ.integrations)-1 {
				m.integ.cursor++
			}

		case key.Matches(msg, keys.IntUp):
			if m.integ.cursor > 0 {
				m.integ.cursor--
			}

		case key.Matches(msg, keys.Toggle):
			if len(m.integ.integrations) > 0 {
				name := m.integ.integrations[m.integ.cursor].Name
				m.integ.staged[name] = !m.integ.staged[name]
			}

		case key.Matches(msg, keys.Info):
			if len(m.integ.integrations) > 0 {
				m.integ.showInfo = true
				m.integ.infoCursor = m.integ.cursor
			}

		case key.Matches(msg, keys.Confirm):
			if m.integrationHasPendingChanges() {
				m.integ.events = nil
				m.integ.returnView = integrationListView
				m.view = integrationProgressView
				updated, cmd := m.startIntegrationApply()
				return updated, cmd
			}

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

func (m Model) updateIntegrationInfo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.integ.showInfo = false
	case key.Matches(msg, keys.Left):
		if m.integ.infoCursor > 0 {
			m.integ.infoCursor--
		} else {
			m.integ.infoCursor = len(m.integ.integrations) - 1
		}
	case key.Matches(msg, keys.Right):
		if m.integ.infoCursor < len(m.integ.integrations)-1 {
			m.integ.infoCursor++
		} else {
			m.integ.infoCursor = 0
		}
	}
	return m, nil
}

func (m Model) viewIntegrationList() string {
	var b strings.Builder

	repoName := filepath.Base(m.repoPath)
	b.WriteString(styleTitle.Render("  sentei \u2500 Integrations"))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render(fmt.Sprintf("  %s (bare)", repoName)))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	for i, integ := range m.integ.integrations {
		cursor := "  "
		if i == m.integ.cursor {
			cursor = "> "
		}

		staged := m.integ.staged[integ.Name]
		current := m.integ.current[integ.Name]

		var checkbox string
		switch {
		case staged && !current:
			checkbox = styleStagedAdd.Render("[+]")
		case !staged && current:
			checkbox = styleStagedRemove.Render("[-]")
		case staged && current:
			checkbox = styleCheckboxOn.Render("[x]")
		default:
			checkbox = styleCheckboxOff.Render("[ ]")
		}

		if i == m.integ.cursor {
			b.WriteString(styleAccent.Render(cursor) + checkbox + " " + integ.Name)
		} else {
			b.WriteString("  " + checkbox + " " + integ.Name)
		}
		b.WriteString("\n")
		b.WriteString("       " + styleDim.Render(integ.Description))
		b.WriteString("\n")

		if i < len(m.integ.integrations)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	pending := m.pendingChangeCount()
	if pending > 0 {
		b.WriteString(styleAccent.Render(fmt.Sprintf("  %d %s pending",
			pending, pluralize(pending, "change", "changes"))))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	legend := fmt.Sprintf("  %s active  %s inactive  %s adding  %s removing",
		styleCheckboxOn.Render("[x]"),
		styleCheckboxOff.Render("[ ]"),
		styleStagedAdd.Render("[+]"),
		styleStagedRemove.Render("[-]"),
	)
	b.WriteString(legend)
	b.WriteString("\n\n")

	if pending > 0 {
		b.WriteString(styleDim.Render("  w/s navigate \u00b7 space toggle \u00b7 ? info \u00b7 enter apply \u00b7 esc back"))
	} else {
		b.WriteString(styleDim.Render("  w/s navigate \u00b7 space toggle \u00b7 ? info \u00b7 esc back"))
	}
	b.WriteString("\n")

	if m.integ.showInfo {
		overlay := m.renderIntegrationInfo()
		return lipgloss.Place(m.width, m.height+6, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars(" "))
	}

	return b.String()
}

func (m Model) renderIntegrationInfo() string {
	if len(m.integ.integrations) == 0 {
		return ""
	}

	integ := m.integ.integrations[m.integ.infoCursor]

	var content strings.Builder

	header := fmt.Sprintf("%s  %s",
		styleTitle.Render(integ.Name),
		styleDim.Render(fmt.Sprintf("%d / %d", m.integ.infoCursor+1, len(m.integ.integrations))),
	)
	content.WriteString(header)
	content.WriteString("\n\n")

	content.WriteString(integ.Description)
	content.WriteString("\n\n")

	if integ.URL != "" {
		content.WriteString(styleDim.Render(integ.URL))
		content.WriteString("\n\n")
	}

	if len(integ.Dependencies) > 0 {
		var depNames []string
		for _, dep := range integ.Dependencies {
			depNames = append(depNames, dep.Name)
		}
		content.WriteString(styleDim.Render("Requires " + strings.Join(depNames, ", ")))
		content.WriteString("\n\n")
	}

	content.WriteString(styleDim.Render("\u25c0 a/\u2190 prev \u00b7 d/\u2192 next \u25b6"))
	content.WriteString("\n")
	content.WriteString(styleDim.Render("esc to close"))

	dialog := styleDialogBox.Render(content.String())
	return dialog
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
