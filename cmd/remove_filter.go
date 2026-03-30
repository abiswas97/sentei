package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abiswas97/sentei/internal/git"
)

// MergedChecker is a function that checks whether a branch is fully merged
// into the default branch. It takes the repo path and the short branch name.
type MergedChecker func(repoPath, branch string) bool

// ResolveFilters returns the worktrees that match the given filter options.
// Filters combine with OR logic: a worktree matching any active filter is included.
// Protected branches (built-in + custom), bare worktrees, and locked worktrees
// are always excluded.
func ResolveFilters(worktrees []git.Worktree, opts *RemoveOptions, protectedBranches []string, isMerged MergedChecker) []git.Worktree {
	protectedSet := make(map[string]bool, len(protectedBranches))
	for _, b := range protectedBranches {
		protectedSet[b] = true
	}

	now := time.Now()
	var result []git.Worktree

	for _, wt := range worktrees {
		if wt.IsBare || wt.IsLocked {
			continue
		}

		branch := shortBranch(wt.Branch)
		if git.IsProtectedBranch(wt.Branch) || protectedSet[branch] {
			continue
		}

		if matchesFilters(wt, opts, now, isMerged, branch) {
			result = append(result, wt)
		}
	}

	return result
}

func matchesFilters(wt git.Worktree, opts *RemoveOptions, now time.Time, isMerged MergedChecker, branch string) bool {
	if opts.All {
		return true
	}

	if opts.Stale > 0 {
		if wt.LastCommitDate.IsZero() {
			fmt.Fprintf(os.Stderr, "Warning: skipping worktree %s (no commit date available)\n", wt.Path)
		} else if now.Sub(wt.LastCommitDate) > opts.Stale {
			return true
		}
	}

	if opts.Merged && isMerged != nil && branch != "" {
		if isMerged("", branch) {
			return true
		}
	}

	return false
}

func shortBranch(branch string) string {
	return strings.TrimPrefix(branch, "refs/heads/")
}

// CheckMerged creates a MergedChecker that uses git merge-base --is-ancestor
// to determine if a branch is fully merged into the default branch.
func CheckMerged(runner git.CommandRunner, repoPath string, defaultBranch string) MergedChecker {
	return func(_ string, branch string) bool {
		_, err := runner.Run(repoPath, "merge-base", "--is-ancestor", branch, defaultBranch)
		return err == nil
	}
}

// DetectDefaultBranch detects the default branch name (main or master).
func DetectDefaultBranch(runner git.CommandRunner, repoPath string) string {
	output, err := runner.Run(repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(output)
		if branch := strings.TrimPrefix(ref, "refs/remotes/origin/"); branch != ref {
			return branch
		}
	}

	// Fallback: check if main or master exists.
	if _, err := runner.Run(repoPath, "rev-parse", "--verify", "refs/heads/main"); err == nil {
		return "main"
	}
	if _, err := runner.Run(repoPath, "rev-parse", "--verify", "refs/heads/master"); err == nil {
		return "master"
	}

	return "main"
}
