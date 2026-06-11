package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/pipeline"
	"github.com/abiswas97/sentei/internal/worktree"
)

type teardownCompleteMsg struct {
	results []pipeline.StepResult
}

func unlockLockedWorktrees(runner git.CommandRunner, repoPath string, worktrees []git.Worktree) {
	for _, wt := range worktrees {
		if wt.IsLocked {
			if err := worktree.UnlockWorktree(runner, repoPath, wt.Path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to unlock %s: %v\n", wt.Path, err)
			}
		}
	}
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Yes):
			return m.beginRemoval()

		case key.Matches(msg, keys.No), key.Matches(msg, keys.Back):
			m.view = listView
		}

	}
	return m, nil
}

// beginRemoval starts the deletion of the current selection: fresh run
// state, teardown of integration artifacts when present, then the deletion
// pipeline. Entered from the at-risk confirmation gate, or directly from
// the list when every selected worktree is clean and pushed.
func (m Model) beginRemoval() (tea.Model, tea.Cmd) {
	m.progressStartedAt = time.Now()
	m.progressToken++
	m.view = progressView
	selected := m.selectedWorktrees()
	m.remove.run = newRemovalRun(selected)

	integrations := integration.All()
	hasTeardown := false
	for _, wt := range selected {
		if len(creator.ScanArtifacts(wt.Path, integrations)) > 0 {
			hasTeardown = true
			break
		}
	}

	unlockLockedWorktrees(m.runner, m.repoPath, selected)

	if hasTeardown {
		m.remove.run.teardownRunning = true
		return m, m.runTeardownPhase(selected, integrations)
	}

	return m.startDeletions()
}

// worktreeAtRisk reports whether removing wt could lose work that exists
// nowhere else: uncommitted or untracked changes, commits not on a remote,
// or a lock someone placed deliberately.
func worktreeAtRisk(wt git.Worktree) bool {
	return wt.HasUncommittedChanges || wt.HasUntrackedFiles || wt.HasUnpushedCommits || wt.IsLocked
}

// startDeletions kicks off the deletion goroutine for the current run's
// worktree snapshot and begins consuming its events.
func (m Model) startDeletions() (tea.Model, tea.Cmd) {
	selected := m.remove.run.worktrees
	ch := make(chan worktree.DeletionEvent, len(selected)*2)
	m.remove.run.progressCh = ch
	go worktree.DeleteWorktrees(os.RemoveAll, selected, 5, ch)
	return m, waitForDeletionEvent(ch)
}

const maxTeardownConcurrency = 5

func (m Model) runTeardownPhase(worktrees []git.Worktree, integrations []integration.Integration) tea.Cmd {
	return func() tea.Msg {
		shell := m.shell
		type indexedResults struct {
			index   int
			results []pipeline.StepResult
		}

		resultsCh := make(chan indexedResults, len(worktrees))
		sem := make(chan struct{}, maxTeardownConcurrency)

		for i, wt := range worktrees {
			sem <- struct{}{}
			go func(idx int, wtPath string) {
				defer func() { <-sem }()
				results := creator.Teardown(shell, wtPath, integrations, func(pipeline.Event) {})
				resultsCh <- indexedResults{index: idx, results: results}
			}(i, wt.Path)
		}

		collected := make([][]pipeline.StepResult, len(worktrees))
		for range worktrees {
			ir := <-resultsCh
			collected[ir.index] = ir.results
		}

		var allResults []pipeline.StepResult
		for _, r := range collected {
			allResults = append(allResults, r...)
		}
		return teardownCompleteMsg{results: allResults}
	}
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	selected := m.selectedWorktrees()

	b.WriteString(viewTitle(titleConfirmDeletion))
	b.WriteString("\n\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  You are about to delete %d %s:\n\n", len(selected), pluralize(len(selected), "worktree", "worktrees"))

	// Columnar rows: badge gutter first so one vertical sweep answers
	// "anything risky?", names aligned after it, notes only on at-risk rows.
	nameWidth := 0
	for _, wt := range selected {
		nameWidth = max(nameWidth, len([]rune(worktreeLabel(wt))))
	}
	nameWidth = min(nameWidth, confirmNameWidthCap)

	var dirtyCount, untrackedCount, lockedCount, unpushedCount int
	for _, wt := range selected {
		var badge, note string
		risky := true
		switch {
		case wt.IsLocked:
			badge, note = "[L]", "locked — will force-remove"
			lockedCount++
		case wt.HasUncommittedChanges:
			badge, note = "[~]", "uncommitted changes — will be lost"
			dirtyCount++
		case wt.HasUntrackedFiles:
			badge, note = "[!]", "untracked files — will be lost"
			untrackedCount++
		case wt.HasUnpushedCommits:
			badge, note = "[^]", "commits not on any remote"
			unpushedCount++
		default:
			badge, risky = "[ok]", false
		}

		badgeStyle := styleStatusClean
		if risky {
			badgeStyle = styleWarning
		}
		name := truncateWithEllipsis(worktreeLabel(wt), nameWidth)
		// Pad by rune count: fmt's %-*s pads by bytes and drifts on …
		pad := strings.Repeat(" ", max(nameWidth-len([]rune(name)), 0))
		line := fmt.Sprintf("    %s  %s%s", badgeStyle.Render(fmt.Sprintf("%-4s", badge)), name, pad)
		if risky {
			line += "  " + styleWarning.Render(note)
		}
		b.WriteString(strings.TrimRight(line, " ") + "\n")
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
			fmt.Fprintf(&b, "    %-28s in %d %s\n", dir, count, pluralize(count, "worktree", "worktrees"))
		}
		b.WriteString("\n")
	}

	if dirtyCount > 0 {
		b.WriteString(styleWarning.Render(
			fmt.Sprintf("  ⚠ %d %s with uncommitted changes that will be lost", dirtyCount, pluralize(dirtyCount, "worktree", "worktrees")),
		))
		b.WriteString("\n")
	}
	if untrackedCount > 0 {
		b.WriteString(styleWarning.Render(
			fmt.Sprintf("  ⚠ %d %s with untracked files that will be lost", untrackedCount, pluralize(untrackedCount, "worktree", "worktrees")),
		))
		b.WriteString("\n")
	}
	if lockedCount > 0 {
		b.WriteString(styleWarning.Render(
			fmt.Sprintf("  ⚠ %d %s locked and will be force-removed", lockedCount, pluralize(lockedCount, "worktree", "worktrees")),
		))
		b.WriteString("\n")
	}
	if unpushedCount > 0 {
		b.WriteString(styleWarning.Render(
			fmt.Sprintf("  ⚠ %d %s with commits not pushed to any remote", unpushedCount, pluralize(unpushedCount, "worktree", "worktrees")),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(viewSeparator(m.width))
	b.WriteString("\n\n")
	b.WriteString(viewFooterDanger(m.width, confirmFooter) + "\n")

	return b.String()
}
