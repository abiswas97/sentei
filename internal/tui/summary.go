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
		case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Confirm):
			if m.menuItems != nil {
				m.view = menuView
				return m, nil
			}
			return m, tea.Quit
		case key.Matches(msg, keys.Back):
			if m.menuItems != nil {
				m.view = menuView
				return m, nil
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewSummary() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("  sentei \u2500 Removal Complete"))
	b.WriteString("\n\n")

	r := m.remove.deletionResult
	if r.FailureCount == 0 {
		b.WriteString(styleSuccess.Render(
			fmt.Sprintf("  %d worktree(s) removed successfully", r.SuccessCount),
		))
		b.WriteString("\n")
	} else {
		fmt.Fprintf(&b, "  %s, %s\n",
			styleSuccess.Render(fmt.Sprintf("%d removed", r.SuccessCount)),
			styleError.Render(fmt.Sprintf("%d failed", r.FailureCount)),
		)
		b.WriteString("\n")
		b.WriteString(styleError.Render("  Failures:\n"))
		for _, o := range r.Outcomes {
			if !o.Success {
				fmt.Fprintf(&b, "    x %s: %s\n", o.Path, o.Error)
			}
		}
	}

	b.WriteString("\n")
	if m.remove.pruneErr != nil && *m.remove.pruneErr != nil {
		b.WriteString(styleWarning.Render(fmt.Sprintf("  Warning: failed to prune worktree metadata: %s", *m.remove.pruneErr)))
		b.WriteString("\n")
	} else {
		b.WriteString(styleDim.Render("  Pruned orphaned worktree metadata"))
		b.WriteString("\n")
	}
	if m.remove.cleanupResult != nil {
		r := m.remove.cleanupResult
		b.WriteString("\n")
		b.WriteString(styleDim.Render("  Cleanup:"))
		b.WriteString("\n")
		if r.StaleRefsRemoved > 0 {
			fmt.Fprintf(&b, "    %s Pruned %d remote ref(s)\n", styleSuccess.Render("v"), r.StaleRefsRemoved)
		}
		if r.ConfigDedupResult.Removed > 0 {
			fmt.Fprintf(&b, "    %s Removed %d config duplicates\n", styleSuccess.Render("v"), r.ConfigDedupResult.Removed)
		}
		if r.GoneBranchesDeleted > 0 {
			fmt.Fprintf(&b, "    %s Deleted %d branch(es) with gone upstream\n", styleSuccess.Render("v"), r.GoneBranchesDeleted)
		}
		if r.ConfigOrphanResult.Removed > 0 {
			fmt.Fprintf(&b, "    %s Removed %d orphaned config section(s)\n", styleSuccess.Render("v"), r.ConfigOrphanResult.Removed)
		}
		if r.NonWtBranchesRemaining > 0 {
			b.WriteString("\n")
			b.WriteString(styleDim.Render(fmt.Sprintf("  Tip: %d local branch(es) not in any worktree.", r.NonWtBranchesRemaining)))
			b.WriteString("\n")
			b.WriteString(styleDim.Render("       Run `sentei cleanup --mode=aggressive` to remove them."))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.menuItems != nil {
		b.WriteString(styleDim.Render("  enter menu \u00b7 q quit"))
	} else {
		b.WriteString(styleDim.Render("  enter quit \u00b7 esc quit"))
	}
	b.WriteString("\n")

	return b.String()
}
