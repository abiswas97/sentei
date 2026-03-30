package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/worktree"
)

const (
	statusPending  = "pending"
	statusRemoving = "removing"
	statusRemoved  = "removed"
	statusFailed   = "failed"
)

type cleanupCompleteMsg struct {
	Result cleanup.Result
}

type worktreeDeleteStartedMsg struct{ Path string }
type worktreeDeletedMsg struct{ Path string }
type worktreeDeleteFailedMsg struct {
	Path string
	Err  error
}
type allDeletionsCompleteMsg struct{}
type pruneCompleteMsg struct{ Err error }

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

func runPrune(runner git.CommandRunner, repoPath string) tea.Cmd {
	return func() tea.Msg {
		err := worktree.PruneWorktrees(runner, repoPath)
		return pruneCompleteMsg{Err: err}
	}
}

func runCleanup(runner git.CommandRunner, repoPath string) tea.Cmd {
	return func() tea.Msg {
		result := cleanup.Run(runner, repoPath, cleanup.Options{Mode: cleanup.ModeSafe}, func(cleanup.Event) {})
		return cleanupCompleteMsg{Result: result}
	}
}

func (m Model) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case worktreeDeleteStartedMsg:
		m.remove.deletionStatuses[msg.Path] = statusRemoving
		return m, waitForDeletionEvent(m.remove.progressCh)

	case worktreeDeletedMsg:
		m.remove.deletionStatuses[msg.Path] = statusRemoved
		m.remove.deletionResult.SuccessCount++
		m.remove.deletionResult.Outcomes = append(m.remove.deletionResult.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: true,
		})
		return m, waitForDeletionEvent(m.remove.progressCh)

	case worktreeDeleteFailedMsg:
		m.remove.deletionStatuses[msg.Path] = statusFailed
		m.remove.deletionResult.FailureCount++
		m.remove.deletionResult.Outcomes = append(m.remove.deletionResult.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: false,
			Error:   fmt.Errorf("removing %s: %w", msg.Path, msg.Err),
		})
		return m, waitForDeletionEvent(m.remove.progressCh)

	case allDeletionsCompleteMsg:
		return m, runPrune(m.runner, m.repoPath)

	case pruneCompleteMsg:
		pruneErr := msg.Err
		m.remove.pruneErr = &pruneErr
		return m, runCleanup(m.runner, m.repoPath)

	case cleanupCompleteMsg:
		m.remove.cleanupResult = &msg.Result
		m.view = summaryView
	}
	return m, nil
}

func (m Model) viewProgress() string {
	var b strings.Builder

	b.WriteString(styleHeader.Render("  Removing Worktrees  "))
	b.WriteString("\n\n")

	// Teardown phase (if any)
	if len(m.remove.teardownResults) > 0 {
		hasFailed := false
		for _, r := range m.remove.teardownResults {
			if r.Status == creator.StepFailed {
				hasFailed = true
			}
		}

		statusText := fmt.Sprintf("%d/%d", len(m.remove.teardownResults), len(m.remove.teardownResults))
		if hasFailed {
			statusText += " " + styleIndicatorWarning.Render(indicatorWarning)
		} else {
			statusText += " " + styleIndicatorDone.Render(indicatorDone)
		}
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseDone.Render("Teardown"), styleDim.Render(statusText))
		b.WriteString("\n")
	}

	// Remove phase
	done := len(m.remove.deletionResult.Outcomes)

	phaseStatus := fmt.Sprintf("%d/%d", done, m.remove.deletionTotal)
	if done == m.remove.deletionTotal && m.remove.deletionTotal > 0 {
		phaseStatus += " " + styleIndicatorDone.Render(indicatorDone)
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseDone.Render("Removing worktrees"), styleDim.Render(phaseStatus))
	} else {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseActive.Render("Removing worktrees"), styleDim.Render(phaseStatus))

		selected := m.selectedWorktrees()
		for _, wt := range selected {
			branch := stripBranchPrefix(wt.Branch)
			status := m.remove.deletionStatuses[wt.Path]

			var ind string
			switch status {
			case statusRemoved:
				ind = styleIndicatorDone.Render(indicatorDone)
			case statusFailed:
				ind = styleIndicatorFailed.Render(indicatorFailed)
			case statusRemoving:
				ind = styleIndicatorActive.Render(indicatorActive)
			default:
				ind = styleIndicatorPending.Render(indicatorPending)
			}

			fmt.Fprintf(&b, "  %s %s\n", ind, branch)
		}
	}

	b.WriteString("\n")

	// Prune & cleanup phase
	if m.remove.pruneErr != nil {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseDone.Render("Prune & cleanup"), styleDim.Render(styleIndicatorDone.Render(indicatorDone)))
	} else if done == m.remove.deletionTotal && m.remove.deletionTotal > 0 {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseActive.Render("Prune & cleanup"), styleDim.Render(styleIndicatorActive.Render(indicatorActive)))
	} else {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhasePending.Render("Prune & cleanup"), styleDim.Render("pending"))
	}

	return b.String()
}
