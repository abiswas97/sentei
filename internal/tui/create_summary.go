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

	result := m.create.result
	wtPath := m.repoPath + "/" + creator.SanitizeBranchPath(branch)
	if result != nil && result.WorktreePath != "" {
		wtPath = result.WorktreePath
	}

	hasFailures := result != nil && result.HasFailures()

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Worktree Created", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	if hasFailures {
		fmt.Fprintf(&b, "  %s %s created with issues\n\n",
			styleIndicatorWarning.Render(indicatorWarning), branch)
	} else {
		fmt.Fprintf(&b, "  %s %s ready\n\n",
			styleIndicatorDone.Render(indicatorDone), branch)
	}

	fmt.Fprintf(&b, "    %-10s %s\n", styleDim.Render("Path"), wtPath)
	fmt.Fprintf(&b, "    %-10s %s (from %s)\n", styleDim.Render("Branch"), branch, base)

	if result != nil {
		for _, phase := range result.Phases {
			label := ""
			switch phase.Name {
			case "Dependencies":
				label = "Deps"
			case "Integrations":
				label = "Index"
			default:
				continue
			}
			for _, step := range phase.Steps {
				status := styleIndicatorDone.Render(indicatorDone)
				if step.Status == creator.StepFailed {
					status = styleIndicatorFailed.Render(indicatorFailed)
					if step.Error != nil {
						status += "  " + styleError.Render(step.Error.Error())
					}
				}
				fmt.Fprintf(&b, "    %-10s %s %s\n", styleDim.Render(label), step.Name, status)
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "    cd %s\n", wtPath)
	b.WriteString("\n")

	if m.menuItems != nil {
		b.WriteString(styleDim.Render("  enter menu \u00b7 q quit"))
	} else {
		b.WriteString(styleDim.Render("  enter quit \u00b7 q quit"))
	}
	b.WriteString("\n")

	return b.String()
}
