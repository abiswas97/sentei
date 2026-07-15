package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/worktree"
)

type teardownCompleteMsg struct {
	results []progress.StepResult
}

const (
	unlockPhaseID   progress.PhaseID = "unlock-worktrees"
	teardownPhaseID progress.PhaseID = "teardown-integrations"
	cleanupPhaseID  progress.PhaseID = "prune-and-cleanup"
	pruneStepID     progress.StepID  = "prune-worktree-metadata"
	cleanupStepID   progress.StepID  = "repository-cleanup"
)

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
// progress. Entered from the at-risk confirmation gate, or directly from
// the list when every selected worktree is clean and pushed.
func (m Model) beginRemoval() (tea.Model, tea.Cmd) {
	m.progressStartedAt = time.Now()
	m.progressToken++
	m.view = progressView
	selected := m.selectedWorktrees()
	m.remove.run = newRemovalRun(selected)

	integrations := integration.All()
	var teardownOps []teardownOperation
	for wtIndex, wt := range selected {
		for _, artifact := range creator.ScanArtifacts(wt.Path, integrations) {
			if integ := findIntegrationByName(integrations, artifact.IntegrationName); integ != nil {
				stepID := progress.StepID(fmt.Sprintf("teardown-%d-%d", wtIndex, len(teardownOps)))
				teardownOps = append(teardownOps, teardownOperation{
					stepID:  stepID,
					label:   integration.TeardownStepName(*integ),
					wtPath:  wt.Path,
					command: integ.Teardown.Command,
					dirs:    append([]string(nil), artifact.Dirs...),
				})
			}
		}
	}

	var phases []progress.PlannedPhase
	var unlockSteps []progress.PlannedStep
	for i, wt := range selected {
		if wt.IsLocked {
			unlockSteps = append(unlockSteps, progress.PlannedStep{ID: progress.StepID(fmt.Sprintf("unlock-%d", i)), Label: worktreeLabel(wt)})
		}
	}
	if len(unlockSteps) > 0 {
		phases = append(phases, progress.PlannedPhase{ID: unlockPhaseID, Label: "Unlock", Steps: unlockSteps})
	}
	if len(teardownOps) > 0 {
		steps := make([]progress.PlannedStep, len(teardownOps))
		for i, op := range teardownOps {
			steps[i] = progress.PlannedStep{ID: op.stepID, Label: op.label}
		}
		phases = append(phases, progress.PlannedPhase{ID: teardownPhaseID, Label: "Teardown", Steps: steps})
	}
	targets := make([]worktree.RemovalTarget, len(selected))
	removeSteps := make([]progress.PlannedStep, len(selected))
	for i, wt := range selected {
		stepID := progress.StepID(fmt.Sprintf("remove-%d", i))
		targets[i] = worktree.RemovalTarget{Worktree: wt, StepID: stepID}
		removeSteps[i] = progress.PlannedStep{ID: stepID, Label: worktreeLabel(wt), Checkpoints: 2}
	}
	phases = append(phases,
		progress.PlannedPhase{ID: worktree.RemovalPhaseID, Label: worktree.RemovalPhaseName, Steps: removeSteps},
		progress.PlannedPhase{ID: cleanupPhaseID, Label: "Prune & cleanup", Steps: []progress.PlannedStep{
			{ID: pruneStepID, Label: "Prune worktree metadata"},
			{ID: cleanupStepID, Label: "Repository cleanup"},
		}},
	)
	ch := make(chan progress.Event, (len(selected)+len(teardownOps)+len(unlockSteps)+2)*5)
	execution, err := progress.Start(progress.Plan{Phases: phases}, func(event progress.Event) { ch <- event })
	if err != nil {
		close(ch)
		return m, waitForRemovalEvent(ch)
	}
	m.remove.run.execution = execution
	m.remove.run.progressCh = ch
	m.remove.run.targets = targets
	m.remove.run.teardownOps = teardownOps
	declarationEvents := len(unlockSteps) + len(teardownOps) + len(removeSteps) + 2 + len(phases)
	for range declarationEvents {
		m.remove.run.events = append(m.remove.run.events, <-ch)
	}
	for i, wt := range selected {
		if !wt.IsLocked {
			continue
		}
		stepID := progress.StepID(fmt.Sprintf("unlock-%d", i))
		_, _ = execution.Run(unlockPhaseID, stepID, func() (string, error) {
			return "", worktree.UnlockWorktree(m.runner, m.repoPath, wt.Path)
		})
	}

	if len(teardownOps) > 0 {
		m.remove.run.teardownRunning = true
		for _, op := range teardownOps {
			m.remove.run.teardownPlanned = append(m.remove.run.teardownPlanned, op.label)
		}
		return m, tea.Batch(waitForRemovalEvent(ch), m.runFrozenTeardownPhase(teardownOps, execution))
	}

	updated, cmd := m.startDeletions()
	return updated, tea.Batch(waitForRemovalEvent(ch), cmd)
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
	execution := m.remove.run.execution
	targets := append([]worktree.RemovalTarget(nil), m.remove.run.targets...)
	return m, func() tea.Msg {
		result := worktree.DeleteWorktrees(execution, worktree.RemovalPhaseID, os.RemoveAll, targets, 5)
		return deletionsCompleteMsg{Result: result}
	}
}

func findIntegrationByName(integrations []integration.Integration, name string) *integration.Integration {
	for i := range integrations {
		if integrations[i].Name == name {
			return &integrations[i]
		}
	}
	return nil
}

const maxTeardownConcurrency = 5

func (m Model) runFrozenTeardownPhase(operations []teardownOperation, execution *progress.Execution) tea.Cmd {
	return func() tea.Msg {
		shell := m.shell
		type indexedResults struct {
			index   int
			results []progress.StepResult
		}

		resultsCh := make(chan indexedResults, len(operations))
		sem := make(chan struct{}, maxTeardownConcurrency)

		for i, operation := range operations {
			sem <- struct{}{}
			go func(idx int, op teardownOperation) {
				defer func() { <-sem }()
				result, _ := execution.Run(teardownPhaseID, op.stepID, func() (string, error) {
					if op.command != "" {
						if _, err := shell.RunShell(op.wtPath, op.command); err == nil {
							return "", nil
						}
					}
					for _, dir := range op.dirs {
						if err := os.RemoveAll(filepath.Join(op.wtPath, strings.TrimSuffix(dir, "/"))); err != nil {
							return "", err
						}
					}
					return "removed artifact dirs", nil
				})
				resultsCh <- indexedResults{index: idx, results: []progress.StepResult{result}}
			}(i, operation)
		}

		collected := make([][]progress.StepResult, len(operations))
		for range operations {
			ir := <-resultsCh
			collected[ir.index] = ir.results
		}

		var allResults []progress.StepResult
		for _, r := range collected {
			allResults = append(allResults, r...)
		}
		return teardownCompleteMsg{results: allResults}
	}
}

// runTeardownPhase is retained as a focused test seam. The confirmation path
// prepares and freezes these operations before calling runFrozenTeardownPhase.
func (m Model) runTeardownPhase(worktrees []git.Worktree, integrations []integration.Integration) tea.Cmd {
	var operations []teardownOperation
	for _, wt := range worktrees {
		for _, artifact := range creator.ScanArtifacts(wt.Path, integrations) {
			integ := findIntegrationByName(integrations, artifact.IntegrationName)
			if integ == nil {
				continue
			}
			operations = append(operations, teardownOperation{
				stepID: progress.StepID(fmt.Sprintf("teardown-%d", len(operations))),
				label:  integration.TeardownStepName(*integ), wtPath: wt.Path,
				command: integ.Teardown.Command, dirs: append([]string(nil), artifact.Dirs...),
			})
		}
	}
	steps := make([]progress.PlannedStep, len(operations))
	for i, op := range operations {
		steps[i] = progress.PlannedStep{ID: op.stepID, Label: op.label}
	}
	execution, err := progress.Start(progress.Plan{Phases: []progress.PlannedPhase{{ID: teardownPhaseID, Label: "Teardown", Steps: steps}}}, nil)
	if err != nil {
		return func() tea.Msg { return teardownCompleteMsg{} }
	}
	return m.runFrozenTeardownPhase(operations, execution)
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
