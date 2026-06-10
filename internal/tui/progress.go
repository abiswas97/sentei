package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/pipeline"
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
	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		return m, nil

	case worktreeDeleteStartedMsg:
		m.remove.run.statuses[msg.Path] = statusRemoving
		return m, waitForDeletionEvent(m.remove.run.progressCh)

	case worktreeDeletedMsg:
		m.remove.run.statuses[msg.Path] = statusRemoved
		m.remove.run.result.SuccessCount++
		m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: true,
		})
		return m, waitForDeletionEvent(m.remove.run.progressCh)

	case worktreeDeleteFailedMsg:
		m.remove.run.statuses[msg.Path] = statusFailed
		m.remove.run.result.FailureCount++
		m.remove.run.result.Outcomes = append(m.remove.run.result.Outcomes, worktree.WorktreeOutcome{
			Path:    msg.Path,
			Success: false,
			Error:   fmt.Errorf("removing %s: %w", msg.Path, msg.Err),
		})
		return m, waitForDeletionEvent(m.remove.run.progressCh)

	case allDeletionsCompleteMsg:
		return m, runPrune(m.runner, m.repoPath)

	case pruneCompleteMsg:
		pruneErr := msg.Err
		m.remove.run.pruneErr = &pruneErr
		return m, runCleanup(m.runner, m.repoPath)

	case cleanupCompleteMsg:
		m.remove.run.cleanupResult = &msg.Result
		m.remove.selected = make(map[string]bool)
		m.worktreeGeneration++
		updated, holdCmd := m.holdOrAdvance(summaryView)
		return updated, tea.Batch(holdCmd, loadWorktreeContext(m.runner, m.repoPath, m.worktreeGeneration))
	}
	return m, nil
}

func (m Model) viewProgress() string {
	var b strings.Builder

	b.WriteString(styleHeader.Render("  Removing Worktrees  "))
	b.WriteString("\n\n")

	// Teardown phase (if any)
	if m.remove.run.teardownRunning {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseActive.Render("Teardown"), styleIndicatorActive.Render(indicatorActive))
		b.WriteString("\n")
	} else if len(m.remove.run.teardownResults) > 0 {
		hasFailed := false
		for _, r := range m.remove.run.teardownResults {
			if r.Status == pipeline.StepFailed {
				hasFailed = true
			}
		}

		statusText := "100%"
		if hasFailed {
			statusText += " " + styleIndicatorWarning.Render(indicatorWarning)
		} else {
			statusText += " " + styleIndicatorDone.Render(indicatorDone)
		}
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseDone.Render("Teardown"), styleDim.Render(statusText))
		b.WriteString("\n")
	}

	// Remove phase
	done := len(m.remove.run.result.Outcomes)
	total := m.remove.run.total()

	pct := 0
	if total > 0 {
		pct = (done * 100) / total
	}
	phaseStatus := fmt.Sprintf("%d%%", pct)
	if done == total && total > 0 {
		phaseStatus += " " + styleIndicatorDone.Render(indicatorDone)
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseDone.Render("Removing worktrees"), styleDim.Render(phaseStatus))
	} else {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseActive.Render("Removing worktrees"), styleDim.Render(phaseStatus))

		for _, wt := range m.remove.run.worktrees {
			branch := stripBranchPrefix(wt.Branch)
			status := m.remove.run.statuses[wt.Path]

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
	if m.remove.run.pruneErr != nil {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseDone.Render("Prune & cleanup"), styleDim.Render(styleIndicatorDone.Render(indicatorDone)))
	} else if done == total && total > 0 {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhaseActive.Render("Prune & cleanup"), styleDim.Render(styleIndicatorActive.Render(indicatorActive)))
	} else {
		fmt.Fprintf(&b, "  %-30s %s\n", stylePhasePending.Render("Prune & cleanup"), styleDim.Render("pending"))
	}

	return b.String()
}
