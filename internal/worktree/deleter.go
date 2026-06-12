package worktree

import (
	"fmt"
	"strings"
	"sync"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
)

// RemovalPhaseName is the canonical phase under which worktree deletions
// report progress; the TUI renders the same phase name.
const RemovalPhaseName = "Removing worktrees"

type WorktreeOutcome struct {
	Path    string
	Success bool
	Error   error
}

type DeletionResult struct {
	SuccessCount int
	FailureCount int
	Outcomes     []WorktreeOutcome
}

func PruneWorktrees(runner git.CommandRunner, repoPath string) error {
	_, err := runner.Run(repoPath, "worktree", "prune")
	return err
}

func UnlockWorktree(runner git.CommandRunner, repoPath, wtPath string) error {
	_, err := runner.Run(repoPath, "worktree", "unlock", wtPath)
	if err != nil && strings.Contains(err.Error(), "is not locked") {
		return nil // already unlocked — not an error
	}
	return err
}

func DeleteWorktrees(remover func(string) error, worktrees []git.Worktree, maxConcurrency int, events chan<- progress.Event) DeletionResult {
	defer close(events)

	if len(worktrees) == 0 {
		return DeletionResult{}
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	var mu sync.Mutex
	result := DeletionResult{
		Outcomes: make([]WorktreeOutcome, len(worktrees)),
	}

	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, wt := range worktrees {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, w git.Worktree) {
			defer wg.Done()
			defer func() { <-sem }()

			events <- progress.Event{Phase: RemovalPhaseName, Step: w.Path, Status: progress.StepRunning}

			err := remover(w.Path)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.FailureCount++
				result.Outcomes[idx] = WorktreeOutcome{
					Path:    w.Path,
					Success: false,
					Error:   fmt.Errorf("removing %s: %w", w.Path, err),
				}
				events <- progress.Event{Phase: RemovalPhaseName, Step: w.Path, Status: progress.StepFailed, Error: err}
			} else {
				result.SuccessCount++
				result.Outcomes[idx] = WorktreeOutcome{
					Path:    w.Path,
					Success: true,
				}
				events <- progress.Event{Phase: RemovalPhaseName, Step: w.Path, Status: progress.StepDone}
			}
		}(i, wt)
	}

	wg.Wait()
	return result
}
