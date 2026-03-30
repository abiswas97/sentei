package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/state"
)

type integrationFinalizedMsg struct {
	current map[string]bool
	err     error
}

func (m Model) updateIntegrationProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case integrationEventMsg:
		m.integ.events = append(m.integ.events, msg.Event)
		return m, waitForIntegrationEvent(m.integ.eventCh, m.integ.doneCh)

	case integrationApplyDoneMsg:
		return m, m.finalizeIntegrationApply()

	case integrationFinalizedMsg:
		if m.integ.returnView != migrateNextView {
			m.integ.current = msg.current
			for _, integ := range m.integ.integrations {
				m.integ.staged[integ.Name] = m.integ.current[integ.Name]
			}
		}
		m.view = m.integ.returnView
		return m, nil
	}
	return m, nil
}

func (m Model) finalizeIntegrationApply() tea.Cmd {
	repoPath := m.repoPath
	returnView := m.integ.returnView
	integrations := m.integ.integrations
	staged := make(map[string]bool)
	for k, v := range m.integ.staged {
		staged[k] = v
	}
	repoResult := m.repo.result

	return func() tea.Msg {
		bareDir := filepath.Join(repoPath, ".bare")
		if returnView == migrateNextView {
			if result, ok := repoResult.(repo.MigrateResult); ok {
				bareDir = filepath.Join(result.BareRoot, ".bare")
			}
		}

		var enabled []string
		for _, integ := range integrations {
			if staged[integ.Name] {
				enabled = append(enabled, integ.Name)
			}
		}
		err := state.Save(bareDir, &state.State{Integrations: enabled})

		current := make(map[string]bool)
		for _, integ := range integrations {
			current[integ.Name] = staged[integ.Name]
		}

		return integrationFinalizedMsg{current: current, err: err}
	}
}

func (m Model) viewIntegrationProgress() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("  sentei \u2500 Applying Integration Changes"))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	// Group events by worktree, preserving insertion order.
	type worktreeGroup struct {
		name   string
		events []integration.ManagerEvent
	}
	var groups []worktreeGroup
	indexByWorktree := make(map[string]int)

	for _, ev := range m.integ.events {
		if _, exists := indexByWorktree[ev.Worktree]; !exists {
			indexByWorktree[ev.Worktree] = len(groups)
			groups = append(groups, worktreeGroup{name: ev.Worktree})
		}
		idx := indexByWorktree[ev.Worktree]
		groups[idx].events = append(groups[idx].events, ev)
	}

	for _, g := range groups {
		b.WriteString(fmt.Sprintf("  %s\n", filepath.Base(g.name)))

		// Deduplicate steps: keep only the last event per step name.
		type stepEntry struct {
			step string
			ev   integration.ManagerEvent
		}
		var steps []stepEntry
		stepIndex := make(map[string]int)
		for _, ev := range g.events {
			if i, exists := stepIndex[ev.Step]; exists {
				steps[i].ev = ev
			} else {
				stepIndex[ev.Step] = len(steps)
				steps = append(steps, stepEntry{step: ev.Step, ev: ev})
			}
		}

		for _, s := range steps {
			var ind string
			switch s.ev.Status {
			case integration.StatusDone:
				ind = styleIndicatorDone.Render(indicatorDone)
			case integration.StatusRunning:
				ind = styleIndicatorActive.Render(indicatorActive)
			case integration.StatusFailed:
				ind = styleIndicatorFailed.Render(indicatorFailed)
			}

			line := fmt.Sprintf("    %s %s", ind, s.ev.Step)
			if s.ev.Error != nil {
				line += " " + styleError.Render(s.ev.Error.Error())
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	// Progress bar.
	total := len(m.integ.events)
	done := 0
	for _, ev := range m.integ.events {
		if ev.Status == integration.StatusDone || ev.Status == integration.StatusFailed {
			done++
		}
	}

	const barWidth = 20
	filled := 0
	if total > 0 {
		filled = (done * barWidth) / total
	}
	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", barWidth-filled)
	b.WriteString(fmt.Sprintf("  %s %d/%d\n", bar, done, total))

	return b.String()
}
