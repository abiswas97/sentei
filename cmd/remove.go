package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

	worktrees, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		return fmt.Errorf("listing worktrees: %w", err)
	}

	worktrees = worktree.EnrichWorktrees(runner, worktrees, worktree.DefaultEnrichConcurrency)

	var isMerged MergedChecker
	if opts.Merged {
		defaultBranch := DetectDefaultBranch(runner, repoPath)
		isMerged = CheckMerged(runner, repoPath, defaultBranch)
	}

	filtered := ResolveFilters(worktrees, opts, nil, isMerged)

	var protectedCount int
	for _, wt := range worktrees {
		if wt.IsBare {
			continue
		}
		if git.IsProtectedBranch(wt.Branch) {
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
		for _, wt := range filtered {
			fmt.Printf("  %s\n", shortBranch(wt.Branch))
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
