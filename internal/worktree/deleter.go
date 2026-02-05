package worktree

import (
	"fmt"
	"sync"

	"github.com/abiswas/wt-sweep/internal/git"
)

type DeletionEventType int

const (
	DeletionStarted DeletionEventType = iota
	DeletionCompleted
	DeletionFailed
)

type DeletionEvent struct {
	Type  DeletionEventType
	Path  string
	Error error
}

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

func DeleteWorktrees(runner git.CommandRunner, repoPath string, worktrees []git.Worktree, maxConcurrency int, progress chan<- DeletionEvent) DeletionResult {
	defer close(progress)

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

			progress <- DeletionEvent{Type: DeletionStarted, Path: w.Path}

			args := []string{"worktree", "remove", "--force"}
			if w.IsLocked {
				args = append(args, "--force")
			}
			args = append(args, w.Path)
			_, err := runner.Run(repoPath, args...)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.FailureCount++
				result.Outcomes[idx] = WorktreeOutcome{
					Path:    w.Path,
					Success: false,
					Error:   fmt.Errorf("removing %s: %w", w.Path, err),
				}
				progress <- DeletionEvent{Type: DeletionFailed, Path: w.Path, Error: err}
			} else {
				result.SuccessCount++
				result.Outcomes[idx] = WorktreeOutcome{
					Path:    w.Path,
					Success: true,
				}
				progress <- DeletionEvent{Type: DeletionCompleted, Path: w.Path}
			}
		}(i, wt)
	}

	wg.Wait()
	return result
}
