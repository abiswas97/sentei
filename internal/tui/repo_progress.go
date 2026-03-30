package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/repo"
)

type repoEventMsg repo.Event

type repoDoneMsg struct {
	result interface{} // CreateResult, CloneResult, or MigrateResult
}

// startRepoPipeline launches the appropriate pipeline based on opts type.
func (m *Model) startRepoPipeline(opts interface{}) tea.Cmd {
	ch := make(chan repo.Event, 50)
	resultCh := make(chan interface{}, 1)
	m.repo.eventCh = ch
	m.repo.resultCh = resultCh

	runner := m.runner
	shell := m.shell

	switch o := opts.(type) {
	case repo.CreateOptions:
		go func() {
			result := repo.Create(runner, shell, o, func(e repo.Event) { ch <- e })
			close(ch)
			resultCh <- result
		}()
	case repo.CloneOptions:
		go func() {
			result := repo.Clone(runner, o, func(e repo.Event) { ch <- e })
			close(ch)
			resultCh <- result
		}()
	case repo.MigrateOptions:
		go func() {
			result := repo.Migrate(runner, shell, o, func(e repo.Event) { ch <- e })
			close(ch)
			resultCh <- result
		}()
	}

	return m.waitForRepoEvent()
}

func (m Model) waitForRepoEvent() tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-m.repo.eventCh
		if !ok {
			result := <-m.repo.resultCh
			return repoDoneMsg{result: result}
		}
		return repoEventMsg(ev)
	}
}

func (m Model) updateRepoProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case repoEventMsg:
		m.repo.events = append(m.repo.events, repo.Event(msg))
		return m, m.waitForRepoEvent()

	case repoDoneMsg:
		m.repo.result = msg.result
		if m.repo.opType == "migrate" {
			m.view = migrateSummaryView
		} else {
			m.view = repoSummaryView
		}
		return m, nil
	}
	return m, nil
}

// repoStepStatus maps repo.StepStatus to creator.StepStatus (same iota values).
func repoStepStatus(s repo.StepStatus) creator.StepStatus {
	return creator.StepStatus(s)
}

func (m Model) buildRepoPhaseDisplays() []phaseDisplay {
	phases := map[string]*phaseDisplay{}
	var order []string

	for _, ev := range m.repo.events {
		pd, exists := phases[ev.Phase]
		if !exists {
			pd = &phaseDisplay{name: ev.Phase}
			phases[ev.Phase] = pd
			order = append(order, ev.Phase)
		}

		found := false
		for i := range pd.steps {
			if pd.steps[i].name == ev.Step {
				pd.steps[i].status = repoStepStatus(ev.Status)
				found = true
				break
			}
		}
		if !found {
			pd.steps = append(pd.steps, stepDisplay{name: ev.Step, status: repoStepStatus(ev.Status)})
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

func (m Model) viewRepoProgress() string {
	var b strings.Builder

	var title string
	var subjectName string
	switch m.repo.opType {
	case "create":
		title = fmt.Sprintf("  sentei %s Creating Repository", "\u2500")
		subjectName = m.repo.nameInput.Value()
	case "clone":
		title = fmt.Sprintf("  sentei %s Cloning Repository", "\u2500")
		subjectName = m.repo.urlInput.Value()
	case "migrate":
		title = fmt.Sprintf("  sentei %s Migrating Repository", "\u2500")
		subjectName = m.repoPath
	}

	b.WriteString(styleTitle.Render(title))
	b.WriteString("\n\n")
	if subjectName != "" {
		b.WriteString(styleAccent.Render("  " + subjectName))
		b.WriteString("\n\n")
	}
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	displays := m.buildRepoPhaseDisplays()

	for i, pd := range displays {
		isComplete := pd.done == pd.total && pd.total > 0
		hasFailure := pd.failed > 0
		isActive := !isComplete && pd.total > 0

		var headerStyle func(strs ...string) string
		var statusText string

		switch {
		case isComplete && !hasFailure:
			headerStyle = stylePhaseDone.Render
			statusText = fmt.Sprintf("%d/%d %s", pd.done, pd.total, styleIndicatorDone.Render(indicatorDone))
		case isComplete && hasFailure:
			headerStyle = stylePhaseActive.Render
			statusText = fmt.Sprintf("%d/%d %s", pd.done, pd.total, styleIndicatorWarning.Render(indicatorWarning))
		case isActive:
			headerStyle = stylePhaseActive.Render
			statusText = fmt.Sprintf("%d/%d", pd.done, pd.total)
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
				fmt.Fprintf(&b, "    %s %s\n", ind, step.name)
			}
		}

		if i < len(displays)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n")

	return b.String()
}
