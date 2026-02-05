package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas/wt-sweep/internal/worktree"
)

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Yes):
			m.view = progressView
			selected := m.selectedWorktrees()
			m.deletionTotal = len(selected)
			m.deletionDone = 0
			for _, wt := range selected {
				m.deletionStatuses[wt.Path] = statusPending
			}
			ch := make(chan worktree.DeletionEvent, len(selected)*2)
			m.progressCh = ch
			go worktree.DeleteWorktrees(m.runner, m.repoPath, selected, 5, ch)
			return m, waitForDeletionEvent(m.progressCh)

		case key.Matches(msg, keys.No), key.Matches(msg, keys.Back):
			m.view = listView
		}
	}
	return m, nil
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	selected := m.selectedWorktrees()

	b.WriteString(styleHeader.Render("  Confirm Deletion  "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  You are about to delete %d worktree(s):\n\n", len(selected)))

	var dirtyCount, untrackedCount, lockedCount int
	for _, wt := range selected {
		branch := stripBranchPrefix(wt.Branch)

		var label string
		switch {
		case wt.IsLocked:
			label = styleWarning.Render("[L] LOCKED - will force-remove")
			lockedCount++
		case wt.HasUncommittedChanges:
			label = styleWarning.Render("[~] HAS UNCOMMITTED CHANGES")
			dirtyCount++
		case wt.HasUntrackedFiles:
			label = styleWarning.Render("[!] has untracked files")
			untrackedCount++
		default:
			label = styleSuccess.Render("(clean)")
		}

		b.WriteString(fmt.Sprintf("    * %s %s\n", branch, label))
	}

	b.WriteString("\n")

	if dirtyCount > 0 {
		b.WriteString(styleWarning.Render(
			fmt.Sprintf("  WARNING: %d worktree(s) have uncommitted changes that will be LOST", dirtyCount),
		))
		b.WriteString("\n")
	}
	if untrackedCount > 0 {
		b.WriteString(styleWarning.Render(
			fmt.Sprintf("  WARNING: %d worktree(s) have untracked files that will be LOST", untrackedCount),
		))
		b.WriteString("\n")
	}
	if lockedCount > 0 {
		b.WriteString(styleWarning.Render(
			fmt.Sprintf("  WARNING: %d worktree(s) are locked and will be force-removed", lockedCount),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  [y] Yes, delete  |  [n] No, go back\n")

	return styleDialogBox.Render(b.String())
}
