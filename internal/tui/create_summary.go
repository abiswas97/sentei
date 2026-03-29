package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/creator"
)

func (m Model) updateCreateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Confirm):
			if m.menuItems != nil {
				m.view = menuView
				return m, nil
			}
			return m, tea.Quit
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewCreateSummary() string {
	var b strings.Builder

	branch := m.create.branchInput.Value()
	base := m.create.baseInput.Value()
	wtPath := m.repoPath + "/" + creator.SanitizeBranchPath(branch)

	hasFailures := false
	for _, ev := range m.create.events {
		if ev.Status == creator.StepFailed {
			hasFailures = true
			break
		}
	}

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Worktree Created", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	if hasFailures {
		b.WriteString(fmt.Sprintf("  %s %s created with issues\n\n",
			styleIndicatorWarning.Render(indicatorWarning), branch))
	} else {
		b.WriteString(fmt.Sprintf("  %s %s ready\n\n",
			styleIndicatorDone.Render(indicatorDone), branch))
	}

	b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Path"), wtPath))
	b.WriteString(fmt.Sprintf("    %-10s %s (from %s)\n", styleDim.Render("Branch"), branch, base))

	// Summarize ecosystems
	for _, eco := range m.create.ecosystems {
		if !m.create.ecoEnabled[eco.Name] {
			continue
		}
		status := styleIndicatorDone.Render(indicatorDone)
		for _, ev := range m.create.events {
			if ev.Phase == "Dependencies" && strings.Contains(ev.Step, eco.Name) && ev.Status == creator.StepFailed {
				status = styleIndicatorFailed.Render(indicatorFailed)
				if ev.Error != nil {
					status += "  " + styleError.Render(ev.Error.Error())
				}
				break
			}
		}
		b.WriteString(fmt.Sprintf("    %-10s %s %s\n", styleDim.Render("Deps"), eco.Name, status))
	}

	// Summarize integrations
	for _, integ := range m.create.integrations {
		if !m.create.intEnabled[integ.Name] {
			continue
		}
		status := styleIndicatorDone.Render(indicatorDone)
		for _, ev := range m.create.events {
			if ev.Phase == "Integrations" && strings.Contains(ev.Step, integ.Name) && ev.Status == creator.StepFailed {
				status = styleIndicatorFailed.Render(indicatorFailed)
				if ev.Error != nil {
					status += "  " + styleError.Render(ev.Error.Error())
				}
				break
			}
		}
		b.WriteString(fmt.Sprintf("    %-10s %s %s\n", styleDim.Render("Index"), integ.Name, status))
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("    cd %s\n", wtPath))
	b.WriteString("\n")

	if m.menuItems != nil {
		b.WriteString(styleDim.Render("  enter menu \u00b7 q quit"))
	} else {
		b.WriteString(styleDim.Render("  enter quit \u00b7 q quit"))
	}
	b.WriteString("\n")

	return b.String()
}
