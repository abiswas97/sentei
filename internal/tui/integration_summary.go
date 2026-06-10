package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
)

type integrationStepOutcome struct {
	step string
	ev   integration.ManagerEvent
}

type integrationWorktreeOutcomes struct {
	worktree string
	steps    []integrationStepOutcome
}

// groupIntegrationEvents folds an apply's event stream into per-worktree step
// outcomes: groups in first-seen order, one entry per step holding its latest
// event. Shared by the progress and summary views.
func groupIntegrationEvents(events []integration.ManagerEvent) []integrationWorktreeOutcomes {
	var groups []integrationWorktreeOutcomes
	groupIndex := make(map[string]int)
	stepIndex := make(map[string]map[string]int)

	for _, ev := range events {
		gi, exists := groupIndex[ev.Worktree]
		if !exists {
			gi = len(groups)
			groupIndex[ev.Worktree] = gi
			groups = append(groups, integrationWorktreeOutcomes{worktree: ev.Worktree})
			stepIndex[ev.Worktree] = make(map[string]int)
		}
		if si, exists := stepIndex[ev.Worktree][ev.Step]; exists {
			groups[gi].steps[si].ev = ev
		} else {
			stepIndex[ev.Worktree][ev.Step] = len(groups[gi].steps)
			groups[gi].steps = append(groups[gi].steps, integrationStepOutcome{step: ev.Step, ev: ev})
		}
	}
	return groups
}

func (m Model) updateIntegrationSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Confirm), key.Matches(msg, keys.Back):
			// Reload from disk so the list's active/staged markers always
			// match persisted state, on success and save-failure alike.
			m.view = integrationListView
			return m, m.loadIntegrationState()
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewIntegrationSummary() string {
	var b strings.Builder

	b.WriteString(viewTitle("Apply Complete"))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

	groups := groupIntegrationEvents(m.integ.events)
	applied, failed := 0, 0
	for _, g := range groups {
		for _, s := range g.steps {
			switch s.ev.Status {
			case integration.StatusDone:
				applied++
			case integration.StatusFailed:
				failed++
			}
		}
	}

	if m.integ.saveErr != nil {
		b.WriteString(styleError.Render(truncateWithEllipsis(
			fmt.Sprintf("  %s Integration state was not saved: %s", indicatorFailed, m.integ.saveErr),
			max(m.width, 40))))
		b.WriteString("\n")
		b.WriteString(styleDim.Render("    The list will show what is actually on disk."))
		b.WriteString("\n\n")
	}

	switch failed {
	case 0:
		b.WriteString(styleSuccess.Render(fmt.Sprintf("  %s %d %s applied", indicatorDone, applied, pluralize(applied, "step", "steps"))))
	default:
		fmt.Fprintf(&b, "  %s, %s",
			styleSuccess.Render(fmt.Sprintf("%d %s applied", applied, pluralize(applied, "step", "steps"))),
			styleError.Render(fmt.Sprintf("%d failed", failed)),
		)
	}
	b.WriteString("\n\n")

	for _, g := range groups {
		fmt.Fprintf(&b, "  %s\n", filepath.Base(g.worktree))
		for _, s := range g.steps {
			switch s.ev.Status {
			case integration.StatusDone:
				fmt.Fprintf(&b, "    %s %s\n", styleIndicatorDone.Render(indicatorDone), s.step)
			case integration.StatusFailed:
				line := fmt.Sprintf("    %s %s", styleIndicatorFailed.Render(indicatorFailed), s.step)
				if s.ev.Error != nil {
					line += " " + styleError.Render(s.ev.Error.Error())
				}
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	b.WriteString(viewKeyHints(KeyHint{"enter", "integrations"}, KeyHint{"q", "quit"}))
	b.WriteString("\n")

	return b.String()
}
