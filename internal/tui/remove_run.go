package tui

import (
	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/worktree"
)

// removalRun holds all state for a single deletion run. It is created fresh
// when the user confirms a deletion, so no outcomes, statuses, or results
// from a previous run can leak into the next one. The worktrees slice is a
// snapshot of the selection at confirm time; the live list may be reloaded
// while the run renders.
type removalRun struct {
	worktrees  []git.Worktree
	statuses   map[string]string
	result     worktree.DeletionResult
	progressCh chan progress.Event
	execution  *progress.Execution
	events     []progress.Event
	targets    []worktree.RemovalTarget

	teardownRunning bool
	// teardownPlanned is the step list scanned at confirm time (one step per
	// worktree-integration with artifacts present), so the Teardown phase
	// displays its real total from its first frame.
	teardownPlanned []string
	teardownResults []progress.StepResult
	teardownOps     []teardownOperation

	pruneErr      *error
	cleanupResult *cleanup.Result
}

type teardownOperation struct {
	stepID  progress.StepID
	label   string
	wtPath  string
	command string
	dirs    []string
}

type unlockOperation struct {
	stepID   progress.StepID
	worktree git.Worktree
}

type removalPreparation struct {
	plan        progress.Plan
	unlockOps   []unlockOperation
	teardownOps []teardownOperation
	targets     []worktree.RemovalTarget
}

func newRemovalRun(selected []git.Worktree) removalRun {
	statuses := make(map[string]string, len(selected))
	for _, wt := range selected {
		statuses[wt.Path] = statusPending
	}
	return removalRun{worktrees: selected, statuses: statuses}
}

func (r removalRun) total() int { return len(r.worktrees) }
