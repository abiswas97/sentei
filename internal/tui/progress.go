package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas/wt-sweep/internal/worktree"
)

type worktreeDeleteStartedMsg struct{ Path string }
type worktreeDeletedMsg struct{ Path string }
type worktreeDeleteFailedMsg struct {
	Path string
	Err  error
}
type allDeletionsCompleteMsg struct {
	Result worktree.DeletionResult
}

func (m Model) startDeletion() tea.Cmd {
	selected := m.selectedWorktrees()

	return func() tea.Msg {
		progress := make(chan worktree.DeletionEvent, len(selected)*2)

		var result worktree.DeletionResult
		done := make(chan struct{})

		go func() {
			result = worktree.DeleteWorktrees(m.runner, m.repoPath, selected, 5, progress)
			close(done)
		}()

		for ev := range progress {
			switch ev.Type {
			case worktree.DeletionStarted:
				// handled via batch below
			case worktree.DeletionCompleted:
				// handled via batch below
			case worktree.DeletionFailed:
				// handled via batch below
			}
			_ = ev
		}

		<-done
		return allDeletionsCompleteMsg{Result: result}
	}
}

func (m Model) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case allDeletionsCompleteMsg:
		m.deletionResult = msg.Result
		m.deletionDone = m.deletionTotal
		for _, o := range msg.Result.Outcomes {
			if o.Success {
				m.deletionStatuses[o.Path] = "removed"
			} else {
				m.deletionStatuses[o.Path] = "failed"
			}
		}
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

	barWidth := 40
	filled := barWidth * pct / 100
	empty := barWidth - filled

	bar := strings.Repeat("#", filled) + strings.Repeat("-", empty)
	b.WriteString(fmt.Sprintf("  [%s] %d/%d (%d%%)\n\n", bar, m.deletionDone, m.deletionTotal, pct))

	selected := m.selectedWorktrees()
	for _, wt := range selected {
		branch := stripBranchPrefix(wt.Branch)
		status := m.deletionStatuses[wt.Path]

		var indicator string
		switch status {
		case "removed":
			indicator = styleSuccess.Render("v") + " " + branch + "  " + styleDim.Render("removed")
		case "failed":
			indicator = styleError.Render("x") + " " + branch + "  " + styleError.Render("failed")
		default:
			indicator = styleDim.Render("~") + " " + branch + "  " + styleDim.Render("removing...")
		}

		b.WriteString("  " + indicator + "\n")
	}

	return b.String()
}
