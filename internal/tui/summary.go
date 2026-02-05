package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Confirm), key.Matches(msg, keys.Back):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewSummary() string {
	var b strings.Builder

	b.WriteString(styleHeader.Render("  Summary  "))
	b.WriteString("\n\n")

	r := m.deletionResult
	if r.FailureCount == 0 {
		b.WriteString(styleSuccess.Render(
			fmt.Sprintf("  %d worktree(s) removed successfully", r.SuccessCount),
		))
		b.WriteString("\n")
	} else {
		b.WriteString(fmt.Sprintf("  %s, %s\n",
			styleSuccess.Render(fmt.Sprintf("%d removed", r.SuccessCount)),
			styleError.Render(fmt.Sprintf("%d failed", r.FailureCount)),
		))
		b.WriteString("\n")
		b.WriteString(styleError.Render("  Failures:\n"))
		for _, o := range r.Outcomes {
			if !o.Success {
				b.WriteString(fmt.Sprintf("    x %s: %s\n", o.Path, o.Error))
			}
		}
	}

	b.WriteString("\n")
	if m.pruneErr != nil && *m.pruneErr != nil {
		b.WriteString(styleWarning.Render(fmt.Sprintf("  Warning: failed to prune worktree metadata: %s", *m.pruneErr)))
		b.WriteString("\n")
	} else {
		b.WriteString(styleDim.Render("  Pruned orphaned worktree metadata"))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styleDim.Render("  Press q, enter, or esc to exit"))
	b.WriteString("\n")

	return b.String()
}
