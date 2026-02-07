package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/dryrun"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/playground"
	"github.com/abiswas97/sentei/internal/tui"
	"github.com/abiswas97/sentei/internal/worktree"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const (
	enrichConcurrency = 10
	playgroundDelay   = 800 * time.Millisecond
)

func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	playgroundFlag := flag.Bool("playground", false, "Launch with a temporary test repo")
	dryRunFlag := flag.Bool("dry-run", false, "Print worktree summary and exit (no interactive TUI)")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("sentei %s (%s, %s)\n", version, commit, date)
		os.Exit(0)
	}

	repoPath := "."
	if flag.NArg() > 0 {
		repoPath = flag.Arg(0)
	}

	if *playgroundFlag {
		var cleanup func()
		var err error
		repoPath, cleanup, err = playground.Setup()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up playground: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Playground repo: %s\n", repoPath)
		defer cleanup()
	}

	runner := &git.GitRunner{}

	worktrees, err := git.ListWorktrees(runner, repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	worktrees = worktree.EnrichWorktrees(runner, worktrees, enrichConcurrency)

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

	if *dryRunFlag {
		dryrun.Print(filtered, os.Stdout)
		os.Exit(0)
	}

	var tuiRunner git.CommandRunner = runner
	if *playgroundFlag {
		tuiRunner = &git.DelayRunner{Inner: runner, Delay: playgroundDelay}
	}

	model := tui.NewModel(filtered, tuiRunner, repoPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
