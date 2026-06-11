package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/abiswas97/sentei/internal/integration"
)

type integrationStateLoadedMsg struct {
	integrations []integration.Integration
	current      map[string]bool
	enabled      []string
	depStatus    map[string]bool // dep name → installed
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

		depStatus := integration.DetectDeps(m.shell, all)

		st, err := m.loadRepoState()
		if err != nil {
			return integrationStateLoadedMsg{
				integrations: all,
				current:      current,
				depStatus:    depStatus,
				err:          err,
			}
		}

		return integrationStateLoadedMsg{
			integrations: all,
			current:      current,
			depStatus:    depStatus,
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

	// Calculate total steps upfront for accurate progress bar.
	// Enable: setup is always 1 per worktree. Deps/install are conditional
	// but we count them as maximum so the bar doesn't exceed total.
	totalSteps := 0
	for _, integ := range toEnable {
		stepsPerWT := 1 // setup (always runs)
		stepsPerWT += len(integ.Dependencies)
		stepsPerWT++ // install
		totalSteps += stepsPerWT * len(wtPaths)
	}
	for _, integ := range toDisable {
		stepsPerWT := 0
		if integ.Teardown.Command != "" {
			stepsPerWT++ // teardown
		}
		stepsPerWT += len(integ.Teardown.Dirs) // dir removals
		totalSteps += stepsPerWT * len(wtPaths)
	}
	m.integ.totalSteps = totalSteps
	m.integ.targetWorktrees = wtPaths

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
	case integrationStateLoadedMsg:
		m.integ.integrations = msg.integrations
		m.integ.current = msg.current
		m.integ.depStatus = msg.depStatus
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

	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelDown:
			m = m.integrationCursorDown()
		case tea.MouseWheelUp:
			m = m.integrationCursorUp()
		}
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Down):
			m = m.integrationCursorDown()

		case key.Matches(msg, keys.Up):
			m = m.integrationCursorUp()

		case key.Matches(msg, keys.Toggle):
			if len(m.integ.integrations) > 0 {
				name := m.integ.integrations[m.integ.cursor].Name
				m.integ.staged[name] = !m.integ.staged[name]
			}

		case key.Matches(msg, keys.Confirm):
			if m.integrationHasPendingChanges() {
				m.integ.events = nil
				m.integ.returnView = integrationListView
				m.progressStartedAt = time.Now()
				m.progressToken++
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

func (m Model) integrationCursorDown() Model {
	if len(m.integ.integrations) > 0 && m.integ.cursor < len(m.integ.integrations)-1 {
		m.integ.cursor++
	}
	return m
}

func (m Model) integrationCursorUp() Model {
	if m.integ.cursor > 0 {
		m.integ.cursor--
	}
	return m
}

func (m Model) viewIntegrationList() string {
	var b strings.Builder

	repoName := filepath.Base(m.repoPath)
	b.WriteString(viewTitle("Integrations"))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render(fmt.Sprintf("  %s (bare)", repoName)))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

	for i, integ := range m.integ.integrations {
		cursor := "  "
		if i == m.integ.cursor {
			cursor = "▸ "
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
			// Settled state stays neutral; green is reserved for staged adds.
			checkbox = "[x]"
		default:
			checkbox = styleCheckboxOff.Render("[ ]")
		}

		if i == m.integ.cursor {
			b.WriteString(styleAccent.Render(cursor) + checkbox + " " + styleAccent.Render(integ.Name))
		} else {
			b.WriteString("  " + checkbox + " " + integ.Name)
		}
		b.WriteString("\n")
		b.WriteString("       " + styleDim.Render(integ.ShortDescription))
		b.WriteString("\n")

		if i < len(m.integ.integrations)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// The pending line is always reserved so the chrome below never shifts
	// while the user toggles.
	pending := m.pendingChangeCount()
	if pending > 0 {
		b.WriteString(styleAccent.Render(fmt.Sprintf("  %d %s pending",
			pending, pluralize(pending, "change", "changes"))))
	}
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

	legend := fmt.Sprintf("  %s active  %s inactive  %s adding  %s removing",
		"[x]",
		styleCheckboxOff.Render("[ ]"),
		styleStagedAdd.Render("[+]"),
		styleStagedRemove.Render("[-]"),
	)
	b.WriteString(legend)
	b.WriteString("\n\n")

	if pending > 0 {
		b.WriteString(viewFooter(m.width, integrationPendingFooter))
	} else {
		b.WriteString(viewFooter(m.width, integrationFooter))
	}
	b.WriteString("\n")

	return b.String()
}

// renderIntegrationsDetail builds the `?` portal page: one section per
// integration with its description, dependency install status, and URL.
// Content is wrapped/truncated to the portal width so the viewport never
// holds over-wide lines.
func (m Model) renderIntegrationsDetail() string {
	width := m.portal.contentWidth()
	wrap := lipgloss.NewStyle().Width(width)

	var b strings.Builder
	for i, integ := range m.integ.integrations {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(styleTitle.Render(truncateWithEllipsis(integ.Name, width)))
		b.WriteString("\n")
		b.WriteString(wrap.Render(integ.Description))
		b.WriteString("\n")

		if len(integ.Dependencies) > 0 {
			b.WriteString(styleDim.Render("Dependencies"))
			b.WriteString("\n")
			for _, dep := range integ.Dependencies {
				installed := m.integ.depStatus[dep.Name]
				var indicator, status string
				if installed {
					indicator = styleStatusClean.Render(indicatorDone)
					status = styleStatusClean.Render("installed")
				} else {
					indicator = styleIndicatorPending.Render(indicatorPending)
					status = styleDim.Render("will be installed")
				}
				fmt.Fprintf(&b, "  %s %-20s %s\n", indicator, dep.Name, status)
			}
		}

		if integ.URL != "" {
			b.WriteString(styleDim.Render(truncateWithEllipsis(integ.URL, width)))
			b.WriteString("\n")
		}
	}
	return b.String()
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
