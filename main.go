package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/cmd"
	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/dryrun"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/playground"
	"github.com/abiswas97/sentei/internal/repo"
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
	if len(os.Args) > 1 && os.Args[1] == "cleanup" {
		cmd.RunCleanup(os.Args[2:])
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "ecosystems" {
		cmd.RunEcosystems(os.Args[2:])
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "integrations" {
		cmd.RunIntegrations()
		return
	}

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
	shell := &git.DefaultShellRunner{}

	// Detect repo context and resolve to bare root if inside a worktree
	context := repo.DetectContext(runner, repoPath)
	if context == repo.ContextBareRepo {
		repoPath = repo.ResolveBareRoot(runner, repoPath)
	}

	// Dry-run mode only works in bare repos
	if *dryRunFlag {
		if context != repo.ContextBareRepo {
			fmt.Fprintf(os.Stderr, "Error: --dry-run requires a bare repository\n")
			os.Exit(1)
		}

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

		if err := dryrun.Print(filtered, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Load config only for bare repos (config lives in repo)
	var cfg *config.Config
	if context == repo.ContextBareRepo {
		var err error
		cfg, err = config.LoadConfig(repoPath,
			config.WithRunner(runner),
			config.WithKnownIntegrations(integration.Names()),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		}
	}

	var tuiRunner git.CommandRunner = runner
	if *playgroundFlag {
		tuiRunner = &git.DelayRunner{Inner: runner, Delay: playgroundDelay}
	}

	// Start at menu — worktrees loaded lazily
	model := tui.NewMenuModel(tuiRunner, shell, repoPath, cfg, context)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
