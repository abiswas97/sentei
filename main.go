package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas/wt-sweep/internal/git"
	"github.com/abiswas/wt-sweep/internal/tui"
	"github.com/abiswas/wt-sweep/internal/worktree"
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

	worktrees = worktree.EnrichWorktrees(runner, worktrees, 10)

	var filtered []git.Worktree
	for _, wt := range worktrees {
		if !wt.IsBare {
			filtered = append(filtered, wt)
		}
	}

	if len(filtered) == 0 {
		fmt.Println("No worktrees found (only the main working tree exists).")
		os.Exit(0)
	}

	model := tui.NewModel(filtered, runner, repoPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
