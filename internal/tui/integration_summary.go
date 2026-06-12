package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/progress"
)

type integrationStepOutcome struct {
	step string
	ev   progress.Event
}

type integrationWorktreeOutcomes struct {
	worktree string
	steps    []integrationStepOutcome
	closed   bool
}

// groupIntegrationEvents folds an apply's event stream into per-worktree step
// outcomes: groups in first-seen order, one entry per step holding its latest
// event, plus the phase-close marker. Shared by the progress and summary views.
func groupIntegrationEvents(events []progress.Event) []integrationWorktreeOutcomes {
	var groups []integrationWorktreeOutcomes
	groupIndex := make(map[string]int)
	stepIndex := make(map[string]map[string]int)

	for _, ev := range events {
		gi, exists := groupIndex[ev.Phase]
		if !exists {
			gi = len(groups)
			groupIndex[ev.Phase] = gi
			groups = append(groups, integrationWorktreeOutcomes{worktree: ev.Phase})
			stepIndex[ev.Phase] = make(map[string]int)
		}
		if ev.Close {
			groups[gi].closed = true
			continue
		}
		if si, exists := stepIndex[ev.Phase][ev.Step]; exists {
			groups[gi].steps[si].ev = ev
		} else {
			stepIndex[ev.Phase][ev.Step] = len(groups[gi].steps)
			groups[gi].steps = append(groups[gi].steps, integrationStepOutcome{step: ev.Step, ev: ev})
		}
	}
	return groups
}

func (m Model) updateIntegrationSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			case progress.StepDone:
				applied++
			case progress.StepFailed:
				failed++
			}
		}
	}
	return applied, failed
}

func groupHasFailure(g integrationWorktreeOutcomes) bool {
	for _, s := range g.steps {
		if s.ev.Status == progress.StepFailed {
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
// header followed by a `✦`/`✗` line per resolved step. Shared by the inline
// summary peek and the detail portal.
// width bounds failure peeks; width <= 0 renders the full untrimmed error
// (the detail portal's mode — nothing is lost, only deferred).
func renderIntegrationOutcomes(b *strings.Builder, groups []integrationWorktreeOutcomes, width int) {
	for _, g := range groups {
		fmt.Fprintf(b, "  %s\n", filepath.Base(g.worktree))
		for _, s := range g.steps {
			switch s.ev.Status {
			case progress.StepDone:
				fmt.Fprintf(b, "    %s %s\n", styleIndicatorDone.Render(indicatorDone), s.step)
			case progress.StepSkipped:
				reason := ""
				if s.ev.Message != "" {
					reason = " (" + s.ev.Message + ")"
				}
				fmt.Fprintf(b, "    %s\n", styleDim.Render("– "+s.step+" – skipped"+reason))
			case progress.StepFailed:
				fmt.Fprintf(b, "    %s %s\n", styleIndicatorFailed.Render(indicatorFailed), s.step)
				if s.ev.Error == nil {
					continue
				}
				if width <= 0 {
					for _, line := range nonEmptyLines(s.ev.Error.Error()) {
						fmt.Fprintf(b, "      %s\n", styleError.Render(line))
					}
					continue
				}
				peek := errorPeekLines(s.ev.Error.Error(), max(width-8, 20))
				for i, line := range peek {
					style := styleDim
					if i == 1 || len(peek) == 1 {
						style = styleError
					}
					fmt.Fprintf(b, "      %s\n", style.Render(line))
				}
			}
		}
		b.WriteString("\n")
	}
}

// integrationSummaryDetailContent renders the full per-worktree breakdown for
// the detail portal, exposed when outcomes overflow the inline peek or any
// failure carries error output (the inline peek promises "? for full output").
func (m Model) integrationSummaryDetailContent() (string, string) {
	groups := orderOutcomesFailuresFirst(groupIntegrationEvents(m.integ.events))
	if len(groups) <= inlineSummaryPreview && !outcomesHaveErrorOutput(groups) {
		return "", ""
	}
	var b strings.Builder
	renderIntegrationOutcomes(&b, groups, 0)
	return portalApplyDetails, strings.TrimRight(b.String(), "\n")
}

func (m Model) viewIntegrationSummary() string {
	var b strings.Builder

	groups := orderOutcomesFailuresFirst(groupIntegrationEvents(m.integ.events))
	applied, failed := countIntegrationOutcomes(groups)

	// "Complete" would oversell a run with failures: the title states the
	// outcome, and the headline leads with the count that matters.
	title := titleApplyComplete
	if failed > 0 {
		title = titleApplyErrors
	}
	b.WriteString(viewTitle(title))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

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
			styleError.Render(fmt.Sprintf("%s %d failed", indicatorFailed, failed)),
			styleSuccess.Render(fmt.Sprintf("%d %s applied", applied, pluralize(applied, "step", "steps"))),
		)
	}
	b.WriteString("\n\n")

	shown := min(len(groups), inlineSummaryPreview)
	renderIntegrationOutcomes(&b, groups[:shown], m.width)
	if rest := len(groups) - shown; rest > 0 {
		b.WriteString(styleDim.Render(fmt.Sprintf("  and %d more %s — ? for details",
			rest, pluralize(rest, "worktree", "worktrees"))))
		b.WriteString("\n\n")
	}

	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	hints := []key.Binding{integrationsOpenHint}
	if len(groups) > shown || outcomesHaveErrorOutput(groups) {
		hints = append(hints, detailsHint)
	}
	hints = append(hints, keys.Quit)
	b.WriteString(viewFooter(m.width, hints))
	b.WriteString("\n")

	return b.String()
}

// outcomesHaveErrorOutput reports whether any failed step carries error text
// worth a portal visit.
func outcomesHaveErrorOutput(groups []integrationWorktreeOutcomes) bool {
	for _, g := range groups {
		for _, s := range g.steps {
			if s.ev.Status == progress.StepFailed && s.ev.Error != nil {
				return true
			}
		}
	}
	return false
}
