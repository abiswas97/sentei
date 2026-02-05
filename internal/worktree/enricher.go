package worktree

import (
	"strings"
	"sync"
	"time"

	"github.com/abiswas/wt-sweep/internal/git"
)

func ParseStatusPorcelain(output string) (hasUncommitted bool, hasUntracked bool) {
	output = strings.TrimSpace(output)
	if output == "" {
		return false, false
	}

	for _, line := range strings.Split(output, "\n") {
		if len(line) < 2 {
			continue
		}
		if strings.HasPrefix(line, "??") {
			hasUntracked = true
		} else {
			hasUncommitted = true
		}
	}
	return hasUncommitted, hasUntracked
}

func ParseCommitDate(output string) (time.Time, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02 15:04:05 -0700", output)
}

func enrichWorktree(runner git.CommandRunner, wt *git.Worktree) {
	if wt.IsBare || wt.IsPrunable {
		return
	}

	dateOutput, err := runner.Run(wt.Path, "log", "-1", "--format=%ai")
	if err != nil {
		wt.EnrichmentError = err.Error()
		return
	}

	commitDate, err := ParseCommitDate(dateOutput)
	if err != nil {
		wt.EnrichmentError = err.Error()
		return
	}
	wt.LastCommitDate = commitDate

	subject, err := runner.Run(wt.Path, "log", "-1", "--format=%s")
	if err != nil {
		wt.EnrichmentError = err.Error()
		return
	}
	wt.LastCommitSubject = strings.TrimSpace(subject)

	statusOutput, err := runner.Run(wt.Path, "status", "--porcelain")
	if err != nil {
		wt.EnrichmentError = err.Error()
		return
	}
	wt.HasUncommittedChanges, wt.HasUntrackedFiles = ParseStatusPorcelain(statusOutput)

	wt.IsEnriched = true
}

func EnrichWorktrees(runner git.CommandRunner, worktrees []git.Worktree, maxConcurrency int) []git.Worktree {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i := range worktrees {
		if worktrees[i].IsBare || worktrees[i].IsPrunable {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			enrichWorktree(runner, &worktrees[idx])
		}(i)
	}

	wg.Wait()
	return worktrees
}
