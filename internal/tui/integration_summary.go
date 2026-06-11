package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

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

	case tea.KeyPressMsg:
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

// inlineSummaryPreview is how many worktree outcomes the apply summary shows
// inline before deferring the rest to the detail portal (`?`).
const inlineSummaryPreview = 3

// countIntegrationOutcomes tallies done and failed steps across all groups.
func countIntegrationOutcomes(groups []integrationWorktreeOutcomes) (applied, failed int) {
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
	return applied, failed
}

func groupHasFailure(g integrationWorktreeOutcomes) bool {
	for _, s := range g.steps {
		if s.ev.Status == integration.StatusFailed {
			return true
		}
	}
	return false
}

// orderOutcomesFailuresFirst moves worktrees with failures ahead of clean ones
// so the summary's limited inline space surfaces problems first. Order within
// each partition is preserved.
func orderOutcomesFailuresFirst(groups []integrationWorktreeOutcomes) []integrationWorktreeOutcomes {
	ordered := make([]integrationWorktreeOutcomes, 0, len(groups))
	for _, g := range groups {
		if groupHasFailure(g) {
			ordered = append(ordered, g)
		}
	}
	for _, g := range groups {
		if !groupHasFailure(g) {
			ordered = append(ordered, g)
		}
	}
	return ordered
}

// renderIntegrationOutcomes writes the per-worktree breakdown to b: a worktree
// header followed by a `●`/`✗` line per resolved step. Shared by the inline
// summary peek and the detail portal.
func renderIntegrationOutcomes(b *strings.Builder, groups []integrationWorktreeOutcomes) {
	for _, g := range groups {
		fmt.Fprintf(b, "  %s\n", filepath.Base(g.worktree))
		for _, s := range g.steps {
			switch s.ev.Status {
			case integration.StatusDone:
				fmt.Fprintf(b, "    %s %s\n", styleIndicatorDone.Render(indicatorDone), s.step)
			case integration.StatusFailed:
				line := fmt.Sprintf("    %s %s", styleIndicatorFailed.Render(indicatorFailed), s.step)
				if s.ev.Error != nil {
					line += " " + styleError.Render(s.ev.Error.Error())
				}
				b.WriteString(line + "\n")
			}
		}
		b.WriteString("\n")
	}
}

// integrationSummaryDetailContent renders the full per-worktree breakdown for
// the detail portal, exposed only when outcomes overflow the inline peek.
func (m Model) integrationSummaryDetailContent() (string, string) {
	groups := orderOutcomesFailuresFirst(groupIntegrationEvents(m.integ.events))
	if len(groups) <= inlineSummaryPreview {
		return "", ""
	}
	var b strings.Builder
	renderIntegrationOutcomes(&b, groups)
	return "Apply Details", strings.TrimRight(b.String(), "\n")
}

func (m Model) viewIntegrationSummary() string {
	var b strings.Builder

	b.WriteString(viewTitle("Apply Complete"))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

	groups := orderOutcomesFailuresFirst(groupIntegrationEvents(m.integ.events))
	applied, failed := countIntegrationOutcomes(groups)

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

	shown := min(len(groups), inlineSummaryPreview)
	renderIntegrationOutcomes(&b, groups[:shown])
	if rest := len(groups) - shown; rest > 0 {
		b.WriteString(styleDim.Render(fmt.Sprintf("  and %d more %s — ? for details",
			rest, pluralize(rest, "worktree", "worktrees"))))
		b.WriteString("\n\n")
	}

	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	hints := []KeyHint{{"enter", "integrations"}}
	if len(groups) > shown {
		hints = append(hints, KeyHint{"?", "details"})
	}
	hints = append(hints, KeyHint{"q", "quit"})
	b.WriteString(viewKeyHints(hints...))
	b.WriteString("\n")

	return b.String()
}
