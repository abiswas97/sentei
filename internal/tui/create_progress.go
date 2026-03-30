package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/creator"
)

type phaseDisplay struct {
	name   string
	steps  []stepDisplay
	done   int
	total  int
	failed int
}

type stepDisplay struct {
	name   string
	status creator.StepStatus
}

func (m Model) updateCreateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case createEventMsg:
		m.create.events = append(m.create.events, msg.Event)
		return m, m.waitForCreateEvent()

	case createCompleteMsg:
		m.create.result = &msg.Result
		m.view = createSummaryView
		return m, nil
	}
	return m, nil
}

func (m Model) buildPhaseDisplays() []phaseDisplay {
	phases := map[string]*phaseDisplay{}
	var order []string

	for _, ev := range m.create.events {
		pd, exists := phases[ev.Phase]
		if !exists {
			pd = &phaseDisplay{name: ev.Phase}
			phases[ev.Phase] = pd
			order = append(order, ev.Phase)
		}

		found := false
		for i := range pd.steps {
			if pd.steps[i].name == ev.Step {
				pd.steps[i].status = ev.Status
				found = true
				break
			}
		}
		if !found {
			pd.steps = append(pd.steps, stepDisplay{name: ev.Step, status: ev.Status})
		}
	}

	var result []phaseDisplay
	for _, name := range order {
		pd := phases[name]
		pd.total = len(pd.steps)
		for _, s := range pd.steps {
			switch s.status {
			case creator.StepDone:
				pd.done++
			case creator.StepFailed:
				pd.failed++
				pd.done++
			}
		}
		result = append(result, *pd)
	}

	return result
}

func (m Model) viewCreateProgress() string {
	var b strings.Builder

	branch := m.create.branchInput.Value()
	base := m.create.baseInput.Value()

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Creating Worktree", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(styleAccent.Render(fmt.Sprintf("  %s \u2192 from %s", branch, base)))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	displays := m.buildPhaseDisplays()

	for i, pd := range displays {
		isComplete := pd.done == pd.total && pd.total > 0
		hasFailure := pd.failed > 0
		isActive := !isComplete && pd.total > 0

		var headerStyle func(strs ...string) string
		var statusText string

		pct := 0
		if pd.total > 0 {
			pct = (pd.done * 100) / pd.total
		}

		switch {
		case isComplete && !hasFailure:
			headerStyle = stylePhaseDone.Render
			statusText = fmt.Sprintf("%d%% %s", pct, styleIndicatorDone.Render(indicatorDone))
		case isComplete && hasFailure:
			headerStyle = stylePhaseActive.Render
			statusText = fmt.Sprintf("%d%% %s", pct, styleIndicatorWarning.Render(indicatorWarning))
		case isActive:
			headerStyle = stylePhaseActive.Render
			statusText = fmt.Sprintf("%d%%", pct)
		default:
			headerStyle = stylePhasePending.Render
			statusText = "pending"
		}

		headerLine := fmt.Sprintf("  %-30s %s", headerStyle(pd.name), styleDim.Render(statusText))
		b.WriteString(headerLine)
		b.WriteString("\n")

		if !isComplete || hasFailure {
			for _, step := range pd.steps {
				var ind string
				switch step.status {
				case creator.StepDone:
					ind = styleIndicatorDone.Render(indicatorDone)
				case creator.StepRunning:
					ind = styleIndicatorActive.Render(indicatorActive)
				case creator.StepFailed:
					ind = styleIndicatorFailed.Render(indicatorFailed)
				default:
					ind = styleIndicatorPending.Render(indicatorPending)
				}
				fmt.Fprintf(&b, "  %s %s\n", ind, step.name)
			}
		}

		if i < len(displays)-1 {
			b.WriteString("\n")
		}
	}

	// Show pending phases that haven't started
	knownPhases := make(map[string]bool)
	for _, pd := range displays {
		knownPhases[pd.name] = true
	}
	pendingNames := []string{"Setup", "Dependencies", "Integrations"}
	for _, name := range pendingNames {
		if !knownPhases[name] {
			fmt.Fprintf(&b, "\n  %-30s %s\n", stylePhasePending.Render(name), styleDim.Render("pending"))
		}
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n")

	return b.String()
}
