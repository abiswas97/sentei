package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/worktree"
)

type teardownCompleteMsg struct {
	results []creator.StepResult
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Yes):
			m.view = progressView
			selected := m.selectedWorktrees()
			m.remove.deletionTotal = len(selected)
			for _, wt := range selected {
				m.remove.deletionStatuses[wt.Path] = statusPending
			}

			integrations := integration.All()
			hasTeardown := false
			for _, wt := range selected {
				if len(creator.ScanArtifacts(wt.Path, integrations)) > 0 {
					hasTeardown = true
					break
				}
			}

			if hasTeardown {
				return m, m.runTeardownPhase(selected, integrations)
			}

			ch := make(chan worktree.DeletionEvent, len(selected)*2)
			m.remove.progressCh = ch
			go worktree.DeleteWorktrees(os.RemoveAll, selected, 5, ch)
			return m, waitForDeletionEvent(m.remove.progressCh)

		case key.Matches(msg, keys.No), key.Matches(msg, keys.Back):
			m.view = listView
		}

	case teardownCompleteMsg:
		m.remove.teardownResults = msg.results
		selected := m.selectedWorktrees()
		ch := make(chan worktree.DeletionEvent, len(selected)*2)
		m.remove.progressCh = ch
		go worktree.DeleteWorktrees(os.RemoveAll, selected, 5, ch)
		return m, waitForDeletionEvent(m.remove.progressCh)
	}
	return m, nil
}

func (m Model) runTeardownPhase(worktrees []git.Worktree, integrations []integration.Integration) tea.Cmd {
	return func() tea.Msg {
		shell := &git.DefaultShellRunner{}
		var allResults []creator.StepResult
		for _, wt := range worktrees {
			results := creator.Teardown(shell, wt.Path, integrations, func(creator.Event) {})
			allResults = append(allResults, results...)
		}
		return teardownCompleteMsg{results: allResults}
	}
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	selected := m.selectedWorktrees()

	b.WriteString(styleHeader.Render("  Confirm Deletion  "))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  You are about to delete %d worktree(s):\n\n", len(selected))

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

		fmt.Fprintf(&b, "    * %s %s\n", branch, label)
	}

	b.WriteString("\n")

	// Integration teardown info
	integrations := integration.All()
	dirCounts := make(map[string]int)
	for _, wt := range selected {
		artifacts := creator.ScanArtifacts(wt.Path, integrations)
		for _, a := range artifacts {
			for _, d := range a.Dirs {
				dirCounts[d]++
			}
		}
	}
	if len(dirCounts) > 0 {
		b.WriteString("  Cleaning up:\n\n")
		for dir, count := range dirCounts {
			noun := "worktree"
			if count > 1 {
				noun = "worktrees"
			}
			fmt.Fprintf(&b, "    %-28s in %d %s\n", dir, count, noun)
		}
		b.WriteString("\n")
	}

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
