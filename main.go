package main

import (
	"fmt"
	"os"

	"github.com/abiswas/wt-sweep/internal/git"
)

func main() {
	repoPath := "."
	if len(os.Args) > 1 {
		repoPath = os.Args[1]
	}

	runner := &git.GitRunner{}
	worktrees, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d worktree(s):\n", len(worktrees))
	for _, wt := range worktrees {
		fmt.Printf("  %s", wt.Path)
		if wt.Branch != "" {
			fmt.Printf(" [%s]", wt.Branch)
		}
		if wt.IsBare {
			fmt.Printf(" (bare)")
		}
		if wt.IsDetached {
			fmt.Printf(" (detached)")
		}
		if wt.IsLocked {
			fmt.Printf(" (locked)")
		}
		if wt.IsPrunable {
			fmt.Printf(" (prunable)")
		}
		fmt.Println()
	}
}
