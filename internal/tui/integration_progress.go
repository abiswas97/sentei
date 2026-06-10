package tui

import (
	"fmt"
	"maps"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/state"
)

type integrationFinalizedMsg struct {
	err error
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
		// The migrate flow has its own summary; hand off unchanged.
		if m.integ.returnView == migrateNextView {
			return m.holdOrAdvance(migrateNextView)
		}
		// In-memory current/staged are never mutated here: dismissing the
		// summary reloads them from persisted state, so the list always
		// matches disk whether the save succeeded or failed.
		m.integ.saveErr = msg.err
		if msg.err == nil {
			m.worktreeGeneration++
			updated, holdCmd := m.holdOrAdvance(integrationSummaryView)
			return updated, tea.Batch(holdCmd, loadWorktreeContext(m.runner, m.repoPath, m.worktreeGeneration))
		}
		return m.holdOrAdvance(integrationSummaryView)
	}
	return m, nil
}

func (m Model) finalizeIntegrationApply() tea.Cmd {
	repoPath := m.repoPath
	returnView := m.integ.returnView
	integrations := m.integ.integrations
	staged := make(map[string]bool)
	maps.Copy(staged, m.integ.staged)
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

		return integrationFinalizedMsg{err: err}
	}
}

func (m Model) viewIntegrationProgress() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("  sentei \u2500 Applying Integration Changes"))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	for _, g := range groupIntegrationEvents(m.integ.events) {
		fmt.Fprintf(&b, "  %s\n", filepath.Base(g.worktree))

		for _, s := range g.steps {
			if s.ev.Status == integration.StatusSkipped {
				continue // Don't display skipped steps.
			}
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

	// Progress bar — use upfront total, count unique completed steps.
	total := m.integ.totalSteps
	stepStatus := make(map[string]bool)
	done := 0
	for _, ev := range m.integ.events {
		key := ev.Worktree + ":" + ev.Step
		if stepStatus[key] {
			continue // Already counted this step.
		}
		if ev.Status == integration.StatusDone || ev.Status == integration.StatusFailed || ev.Status == integration.StatusSkipped {
			stepStatus[key] = true
			done++
		}
	}
	// If no upfront total was set, fall back to discovered steps.
	if total == 0 {
		total = done
	}

	const barWidth = 20
	filled := 0
	if total > 0 {
		filled = (done * barWidth) / total
	}
	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", barWidth-filled)
	pct := 0
	if total > 0 {
		pct = (done * 100) / total
	}
	fmt.Fprintf(&b, "  %s %d%%\n", bar, pct)

	return b.String()
}
