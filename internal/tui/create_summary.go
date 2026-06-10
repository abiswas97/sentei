package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/pipeline"
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
	wtPath := git.WorktreePath(m.repoPath, branch)
	if result != nil && result.WorktreePath != "" {
		wtPath = result.WorktreePath
	}

	hasFailures := result != nil && result.HasFailures()

	b.WriteString(viewTitle("Worktree Created"))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

	if hasFailures {
		fmt.Fprintf(&b, "  %s %s created with issues\n\n",
			styleIndicatorWarning.Render(indicatorWarning), branch)
	} else {
		fmt.Fprintf(&b, "  %s %s ready\n\n",
			styleIndicatorDone.Render(indicatorDone), branch)
	}

	fmt.Fprintf(&b, "    %-10s %s\n", styleDim.Render("Path"), truncateWithEllipsis(wtPath, max(m.width-16, 20)))
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
				if step.Status == pipeline.StepFailed {
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
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "    cd %s\n", truncateWithEllipsis(wtPath, max(m.width-8, 20)))
	b.WriteString("\n")

	if m.menuItems != nil {
		b.WriteString(viewKeyHints(KeyHint{"enter", "menu"}, KeyHint{"q", "quit"}))
	} else {
		b.WriteString(viewKeyHints(KeyHint{"enter", "quit"}, KeyHint{"q", "quit"}))
	}
	b.WriteString("\n")

	return b.String()
}
