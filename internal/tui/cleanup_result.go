package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/cleanup"
)

type standaloneCleanupDoneMsg struct {
	result cleanup.Result
}

func (m Model) updateCleanupResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case standaloneCleanupDoneMsg:
		m.remove.cleanupResult = &msg.result
		return m, nil

	case tea.KeyMsg:
		if m.remove.cleanupResult == nil {
			return m, nil
		}
		switch {
		case key.Matches(msg, keys.Confirm), key.Matches(msg, keys.Quit), key.Matches(msg, keys.Back):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewCleanupResult() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("  sentei \u2500 Cleanup Complete"))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	r := m.remove.cleanupResult
	if r == nil {
		b.WriteString(styleDim.Render("  Running cleanup\u2026"))
		b.WriteString("\n")
		return b.String()
	}

	// Check if anything was actually done
	totalActions := r.StaleRefsRemoved + r.ConfigDedupResult.Removed +
		r.GoneBranchesDeleted + r.ConfigOrphanResult.Removed +
		r.NonWtBranchesDeleted + r.WorktreesPruned

	if len(r.Errors) == 0 && totalActions == 0 {
		fmt.Fprintf(&b, "  %s Repository is clean\n\n",
			styleIndicatorDone.Render(indicatorDone))
	}

	// Stale refs
	if r.StaleRefsRemoved > 0 {
		fmt.Fprintf(&b, "  %s Pruned %d stale remote %s\n",
			styleIndicatorDone.Render(indicatorDone),
			r.StaleRefsRemoved,
			pluralize(r.StaleRefsRemoved, "ref", "refs"))
	} else if len(r.Errors) == 0 {
		fmt.Fprintf(&b, "  %s No stale remote refs\n",
			styleIndicatorPending.Render(indicatorPending))
	}

	// Config dedup
	if r.ConfigDedupResult.Removed > 0 {
		fmt.Fprintf(&b, "  %s Removed %d config %s\n",
			styleIndicatorDone.Render(indicatorDone),
			r.ConfigDedupResult.Removed,
			pluralize(r.ConfigDedupResult.Removed, "duplicate", "duplicates"))
	} else if len(r.Errors) == 0 {
		fmt.Fprintf(&b, "  %s No config duplicates\n",
			styleIndicatorPending.Render(indicatorPending))
	}

	// Gone branches
	if r.GoneBranchesDeleted > 0 {
		fmt.Fprintf(&b, "  %s Deleted %d %s with gone upstream\n",
			styleIndicatorDone.Render(indicatorDone),
			r.GoneBranchesDeleted,
			pluralize(r.GoneBranchesDeleted, "branch", "branches"))
	} else if len(r.Errors) == 0 {
		fmt.Fprintf(&b, "  %s No branches with gone upstream\n",
			styleIndicatorPending.Render(indicatorPending))
	}

	// Orphaned configs
	if r.ConfigOrphanResult.Removed > 0 {
		fmt.Fprintf(&b, "  %s Removed %d orphaned config %s\n",
			styleIndicatorDone.Render(indicatorDone),
			r.ConfigOrphanResult.Removed,
			pluralize(r.ConfigOrphanResult.Removed, "section", "sections"))
	} else if len(r.Errors) == 0 {
		fmt.Fprintf(&b, "  %s No orphaned config sections\n",
			styleIndicatorPending.Render(indicatorPending))
	}

	// Pruned worktrees
	if r.WorktreesPruned > 0 {
		fmt.Fprintf(&b, "  %s Pruned %d stale %s\n",
			styleIndicatorDone.Render(indicatorDone),
			r.WorktreesPruned,
			pluralize(r.WorktreesPruned, "worktree", "worktrees"))
	} else if len(r.Errors) == 0 {
		fmt.Fprintf(&b, "  %s No stale worktrees\n",
			styleIndicatorPending.Render(indicatorPending))
	}

	// Errors
	for _, e := range r.Errors {
		fmt.Fprintf(&b, "  %s %s: %s\n",
			styleIndicatorFailed.Render(indicatorFailed),
			e.Step,
			styleError.Render(e.Err.Error()))
	}

	// Tip about aggressive mode
	if r.NonWtBranchesRemaining > 0 {
		b.WriteString("\n")
		b.WriteString(styleDim.Render(fmt.Sprintf("  Tip: %d local %s not in any worktree.",
			r.NonWtBranchesRemaining,
			pluralize(r.NonWtBranchesRemaining, "branch", "branches"))))
		b.WriteString("\n")
		b.WriteString(styleDim.Render("       Run `sentei cleanup --mode=aggressive` to remove them."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  enter quit \u00b7 esc quit"))
	b.WriteString("\n")

	return b.String()
}
