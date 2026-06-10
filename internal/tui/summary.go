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

	b.WriteString(viewTitle("Removal Complete"))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")

	r := m.remove.run.result
	if r.FailureCount == 0 {
		fmt.Fprintf(&b, "  %s %s\n",
			styleIndicatorDone.Render(indicatorDone),
			styleSuccess.Render(fmt.Sprintf("%d worktree(s) removed successfully", r.SuccessCount)))
	} else {
		fmt.Fprintf(&b, "  %s, %s\n",
			styleSuccess.Render(fmt.Sprintf("%d removed", r.SuccessCount)),
			styleError.Render(fmt.Sprintf("%d failed", r.FailureCount)),
		)
		b.WriteString("\n")
		b.WriteString(styleError.Render("  Failures:\n"))
		for _, o := range r.Outcomes {
			if !o.Success {
				fmt.Fprintf(&b, "    %s %s: %s\n",
					styleIndicatorFailed.Render(indicatorFailed),
					truncateWithEllipsis(o.Path, max(m.width-10, 20)),
					truncateWithEllipsis(fmt.Sprint(o.Error), max(m.width-10, 20)))
			}
		}
	}

	b.WriteString("\n")
	if m.remove.run.pruneErr != nil && *m.remove.run.pruneErr != nil {
		b.WriteString(styleWarning.Render(fmt.Sprintf("  Warning: failed to prune worktree metadata: %s", *m.remove.run.pruneErr)))
		b.WriteString("\n")
	} else {
		b.WriteString(styleDim.Render("  Pruned orphaned worktree metadata"))
		b.WriteString("\n")
	}

	if cr := m.remove.run.cleanupResult; cr != nil {
		cleanupActions := cr.StaleRefsRemoved + cr.ConfigDedupResult.Removed +
			cr.GoneBranchesDeleted + cr.ConfigOrphanResult.Removed
		if cleanupActions > 0 {
			b.WriteString("\n")
			b.WriteString(styleDim.Render("  Cleanup:"))
			b.WriteString("\n")
			if cr.StaleRefsRemoved > 0 {
				fmt.Fprintf(&b, "    %s Pruned %d remote ref(s)\n", styleIndicatorDone.Render(indicatorDone), cr.StaleRefsRemoved)
			}
			if cr.ConfigDedupResult.Removed > 0 {
				fmt.Fprintf(&b, "    %s Removed %d config duplicates\n", styleIndicatorDone.Render(indicatorDone), cr.ConfigDedupResult.Removed)
			}
			if cr.GoneBranchesDeleted > 0 {
				fmt.Fprintf(&b, "    %s Deleted %d branch(es) with gone upstream\n", styleIndicatorDone.Render(indicatorDone), cr.GoneBranchesDeleted)
			}
			if cr.ConfigOrphanResult.Removed > 0 {
				fmt.Fprintf(&b, "    %s Removed %d orphaned config section(s)\n", styleIndicatorDone.Render(indicatorDone), cr.ConfigOrphanResult.Removed)
			}
		}
		if cr.NonWtBranchesRemaining > 0 {
			b.WriteString("\n")
			b.WriteString(styleDim.Render(fmt.Sprintf("  Tip: %d local branch(es) not in any worktree.", cr.NonWtBranchesRemaining)))
			b.WriteString("\n")
			b.WriteString(styleDim.Render("       Run `sentei cleanup --mode=aggressive` to remove them."))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	if m.menuItems != nil {
		b.WriteString(viewKeyHints(KeyHint{"enter", "menu"}, KeyHint{"q", "quit"}))
	} else {
		b.WriteString(viewKeyHints(KeyHint{"enter", "quit"}, KeyHint{"esc", "quit"}))
	}
	b.WriteString("\n")

	return b.String()
}
