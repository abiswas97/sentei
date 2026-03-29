package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
)

type optionItem struct {
	label       string
	description string
	hint        string
	key         string
	section     string // "setup" or "integration"
}

func (m Model) buildOptionItems() []optionItem {
	var items []optionItem

	for _, eco := range m.create.ecosystems {
		items = append(items, optionItem{
			label:   fmt.Sprintf("Install dependencies (%s)", eco.Name),
			hint:    eco.Name + " detected",
			key:     "eco:" + eco.Name,
			section: "setup",
		})
	}

	items = append(items, optionItem{
		label:   "Merge default branch",
		hint:    fmt.Sprintf("%s \u2192 %s", m.create.baseInput.Value(), m.create.branchInput.Value()),
		key:     "merge",
		section: "setup",
	})

	hasEnvFiles := false
	var envFileNames []string
	for _, eco := range m.create.ecosystems {
		if len(eco.EnvFiles) > 0 {
			hasEnvFiles = true
			envFileNames = append(envFileNames, eco.EnvFiles...)
		}
	}
	if hasEnvFiles {
		items = append(items, optionItem{
			label:   "Copy environment files",
			hint:    strings.Join(envFileNames, ", "),
			key:     "envfiles",
			section: "setup",
		})
	}

	for _, integ := range m.create.integrations {
		items = append(items, optionItem{
			label:       integ.Name,
			description: integ.Description,
			hint:        integ.URL,
			key:         "int:" + integ.Name,
			section:     "integration",
		})
	}

	return items
}

func (m Model) isOptionEnabled(item optionItem) bool {
	switch {
	case strings.HasPrefix(item.key, "eco:"):
		name := strings.TrimPrefix(item.key, "eco:")
		return m.create.ecoEnabled[name]
	case item.key == "merge":
		return m.create.mergeBase
	case item.key == "envfiles":
		return m.create.copyEnvFiles
	case strings.HasPrefix(item.key, "int:"):
		name := strings.TrimPrefix(item.key, "int:")
		return m.create.intEnabled[name]
	}
	return false
}

func (m *Model) toggleOption(item optionItem) {
	switch {
	case strings.HasPrefix(item.key, "eco:"):
		name := strings.TrimPrefix(item.key, "eco:")
		m.create.ecoEnabled[name] = !m.create.ecoEnabled[name]
	case item.key == "merge":
		m.create.mergeBase = !m.create.mergeBase
	case item.key == "envfiles":
		m.create.copyEnvFiles = !m.create.copyEnvFiles
	case strings.HasPrefix(item.key, "int:"):
		name := strings.TrimPrefix(item.key, "int:")
		m.create.intEnabled[name] = !m.create.intEnabled[name]
	}
}

func (m Model) updateCreateOptions(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := m.buildOptionItems()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.view = createBranchView
			m.create.branchInput.Focus()
			return m, m.create.branchInput.Cursor.BlinkCmd()

		case key.Matches(msg, keys.Down):
			if m.create.optionsCursor < len(items)-1 {
				m.create.optionsCursor++
			}

		case key.Matches(msg, keys.Up):
			if m.create.optionsCursor > 0 {
				m.create.optionsCursor--
			}

		case key.Matches(msg, keys.Toggle):
			if m.create.optionsCursor < len(items) {
				m.toggleOption(items[m.create.optionsCursor])
			}

		case key.Matches(msg, keys.Confirm):
			m.startCreation()
			m.view = createProgressView
			return m, m.waitForCreateEvent()
		}
	}
	return m, nil
}

func (m *Model) startCreation() {
	var enabledEcos []config.EcosystemConfig
	for _, eco := range m.create.ecosystems {
		if m.create.ecoEnabled[eco.Name] {
			enabledEcos = append(enabledEcos, eco)
		}
	}

	var enabledInts []integration.Integration
	for _, integ := range m.create.integrations {
		if m.create.intEnabled[integ.Name] {
			enabledInts = append(enabledInts, integ)
		}
	}

	opts := creator.Options{
		BranchName:     m.create.branchInput.Value(),
		BaseBranch:     m.create.baseInput.Value(),
		RepoPath:       m.repoPath,
		SourceWorktree: m.findSourceWorktree(),
		MergeBase:      m.create.mergeBase,
		CopyEnvFiles:   m.create.copyEnvFiles,
		Ecosystems:     enabledEcos,
		Integrations:   enabledInts,
	}

	ch := make(chan creator.Event, 50)
	resultCh := make(chan creator.Result, 1)
	m.create.eventCh = ch
	m.create.resultCh = resultCh

	go func() {
		result := creator.Run(m.runner, &git.DefaultShellRunner{}, opts, func(e creator.Event) {
			ch <- e
		})
		close(ch)
		resultCh <- result
	}()
}

func (m Model) waitForCreateEvent() tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-m.create.eventCh
		if !ok {
			result := <-m.create.resultCh
			return createCompleteMsg{Result: result}
		}
		return createEventMsg{Event: ev}
	}
}

type createEventMsg struct {
	Event creator.Event
}
type createCompleteMsg struct {
	Result creator.Result
}

func (m Model) viewCreateOptions() string {
	var b strings.Builder
	items := m.buildOptionItems()

	branch := m.create.branchInput.Value()
	base := m.create.baseInput.Value()

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Create Worktree", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(styleAccent.Render(fmt.Sprintf("  %s \u2192 from %s", branch, base)))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	currentSection := ""
	for i, item := range items {
		if item.section != currentSection {
			currentSection = item.section
			sectionLabel := "Setup"
			if currentSection == "integration" {
				sectionLabel = "Integrations"
			}
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString("  " + styleTitle.Render(sectionLabel))
			b.WriteString("\n\n")
		}

		cursor := "  "
		if i == m.create.optionsCursor {
			cursor = "> "
		}

		var checkbox string
		if m.isOptionEnabled(item) {
			checkbox = styleCheckboxOn.Render("[x]")
		} else {
			checkbox = styleCheckboxOff.Render("[ ]")
		}

		hint := ""
		if item.hint != "" {
			hint = "  " + styleDim.Render(item.hint)
		}

		if i == m.create.optionsCursor {
			b.WriteString(styleAccent.Render(cursor) + checkbox + " " + item.label + hint)
		} else {
			b.WriteString("  " + checkbox + " " + item.label + hint)
		}
		b.WriteString("\n")

		if item.description != "" {
			b.WriteString("        " + styleDim.Render(item.description))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  space toggle \u00b7 enter create \u00b7 esc back"))
	b.WriteString("\n")

	return b.String()
}
