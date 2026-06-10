package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/worktree"
)

// RunRemove executes the remove worktrees command in non-interactive mode.
func RunRemove(args []string) error {
	opts, err := ParseRemoveFlags(args)
	if err != nil {
		return err
	}
	if err := ValidateRemoveForNonInteractive(opts); err != nil {
		return err
	}

	repoPath := "."
	if opts.RepoPath != "" {
		repoPath = opts.RepoPath
	}
	if absPath, err := filepath.Abs(repoPath); err == nil {
		repoPath = absPath
	}

	runner := &git.GitRunner{}

	context := repo.DetectContext(runner, repoPath)
	if context != repo.ContextBareRepo {
		return fmt.Errorf("remove requires a bare repository (detected: %v)", context)
	}

	// Normalize to the bare root: when run from inside a worktree, the repo path
	// is a worktree whose HEAD is its own checked-out branch, so default-branch
	// detection would return that branch and leave the real default unprotected.
	repoPath = repo.ResolveBareRoot(runner, repoPath)

	worktrees, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		return fmt.Errorf("listing worktrees: %w", err)
	}

	worktrees = worktree.EnrichWorktrees(runner, worktrees, worktree.DefaultEnrichConcurrency)

	// Detect the default branch once: it is always protected (it may be
	// non-standard, e.g. "production"), and --merged needs it as the merge target.
	defaultBranch := git.DetectDefaultBranch(runner, repoPath)

	var isMerged MergedChecker
	if opts.Merged {
		isMerged = CheckMerged(runner, repoPath, defaultBranch)
	}

	filtered := ResolveFilters(worktrees, opts, nil, defaultBranch, isMerged)

	// Count a protected worktree as "skipped" only if the active filter would
	// otherwise have selected it — otherwise the message implies protection saved
	// a worktree the filter never wanted.
	now := time.Now()
	var protectedCount int
	for _, wt := range worktrees {
		if wt.IsBare || !git.IsProtectedBranchWith(wt.Branch, defaultBranch) {
			continue
		}
		if matchesFilters(wt, opts, now, isMerged, shortBranch(wt.Branch)) {
			protectedCount++
		}
	}

	// Unlock locked worktrees so removal + prune can clean them up
	for _, wt := range filtered {
		if wt.IsLocked {
			if err := worktree.UnlockWorktree(runner, repoPath, wt.Path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to unlock %s: %v\n", wt.Path, err)
			}
		}
	}

	if len(filtered) == 0 {
		fmt.Println("No worktrees matched the specified filters.")
		return nil
	}

	if opts.DryRun {
		fmt.Printf("%s(dry run)%s Would remove %d worktree(s):\n", dim, nc, len(filtered))
		dirtyCount := 0
		for _, wt := range filtered {
			marker := ""
			if wt.HasUncommittedChanges || wt.HasUntrackedFiles {
				marker = yellow + "  (uncommitted/untracked — will be LOST)" + nc
				dirtyCount++
			}
			fmt.Printf("  %s%s\n", shortBranch(wt.Branch), marker)
		}
		if dirtyCount > 0 {
			fmt.Printf("\n%sWarning:%s %d worktree(s) have changes that removal will discard.\n", yellow, nc, dirtyCount)
		}
		return nil
	}

	fmt.Printf("Removing %d worktree(s)...\n", len(filtered))

	remover := func(path string) error {
		_, err := runner.Run(repoPath, "worktree", "remove", "--force", path)
		return err
	}

	progress := make(chan worktree.DeletionEvent, 2*len(filtered))
	result := worktree.DeleteWorktrees(remover, filtered, 5, progress)

	if err := worktree.PruneWorktrees(runner, repoPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to prune worktrees: %v\n", err)
	}

	fmt.Printf("\n%sRemoved:%s %d worktree(s)\n", green, nc, result.SuccessCount)
	if result.FailureCount > 0 {
		fmt.Printf("%sFailed:%s %d worktree(s)\n", yellow, nc, result.FailureCount)
		for _, o := range result.Outcomes {
			if !o.Success {
				fmt.Printf("  %s\n", o.Error)
			}
		}
	}
	if protectedCount > 0 {
		fmt.Printf("%sSkipped (protected):%s %d worktree(s)\n", dim, nc, protectedCount)
	}

	return nil
}
