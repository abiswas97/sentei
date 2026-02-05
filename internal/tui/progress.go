package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas/wt-sweep/internal/worktree"
)

const (
	statusPending  = "pending"
	statusRemoving = "removing"
	statusRemoved  = "removed"
	statusFailed   = "failed"

	progressBarWidth = 40
)

type worktreeDeleteStartedMsg struct{ Path string }
type worktreeDeletedMsg struct{ Path string }
type worktreeDeleteFailedMsg struct {
	Path string
	Err  error
}
type allDeletionsCompleteMsg struct{}

func waitForDeletionEvent(ch <-chan worktree.DeletionEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return allDeletionsCompleteMsg{}
		}
		switch ev.Type {
		case worktree.DeletionStarted:
			return worktreeDeleteStartedMsg{Path: ev.Path}
		case worktree.DeletionCompleted:
			return worktreeDeletedMsg{Path: ev.Path}
		case worktree.DeletionFailed:
			return worktreeDeleteFailedMsg{Path: ev.Path, Err: ev.Error}
		default:
			return waitForDeletionEvent(ch)()
		}
	}
}

func (m Model) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case worktreeDeleteStartedMsg:
		m.deletionStatuses[msg.Path] = statusRemoving
		return m, waitForDeletionEvent(m.progressCh)

	case worktreeDeletedMsg:
		m.deletionStatuses[msg.Path] = statusRemoved
		m.deletionDone++
		m.deletionResult.SuccessCount++
		m.deletionResult.Outcomes = append(m.deletionResult.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: true,
		})
		return m, waitForDeletionEvent(m.progressCh)

	case worktreeDeleteFailedMsg:
		m.deletionStatuses[msg.Path] = statusFailed
		m.deletionDone++
		m.deletionResult.FailureCount++
		m.deletionResult.Outcomes = append(m.deletionResult.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: false,
			Error:   fmt.Errorf("removing %s: %w", msg.Path, msg.Err),
		})
		return m, waitForDeletionEvent(m.progressCh)

	case allDeletionsCompleteMsg:
		m.view = summaryView
	}
	return m, nil
}

func (m Model) viewProgress() string {
	var b strings.Builder

	b.WriteString(styleHeader.Render("  Removing Worktrees  "))
	b.WriteString("\n\n")

	pct := 0
	if m.deletionTotal > 0 {
		pct = m.deletionDone * 100 / m.deletionTotal
	}

	filled := progressBarWidth * pct / 100
	empty := progressBarWidth - filled

	bar := strings.Repeat("#", filled) + strings.Repeat("-", empty)
	b.WriteString(fmt.Sprintf("  [%s] %d/%d (%d%%)\n\n", bar, m.deletionDone, m.deletionTotal, pct))

	selected := m.selectedWorktrees()
	for _, wt := range selected {
		branch := stripBranchPrefix(wt.Branch)
		status := m.deletionStatuses[wt.Path]

		var indicator string
		switch status {
		case statusRemoved:
			indicator = styleSuccess.Render("v") + " " + branch + "  " + styleDim.Render("removed")
		case statusFailed:
			indicator = styleError.Render("x") + " " + branch + "  " + styleError.Render("failed")
		case statusRemoving:
			indicator = styleWarning.Render("~") + " " + branch + "  " + styleDim.Render("removing...")
		default:
			indicator = styleDim.Render(".") + " " + branch + "  " + styleDim.Render("pending")
		}

		b.WriteString("  " + indicator + "\n")
	}

	return b.String()
}
