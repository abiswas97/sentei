package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/cmd"
	"github.com/abiswas97/sentei/internal/cli"
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

func buildRegistry() *cli.Registry {
	r := cli.NewRegistry()

	r.Register(&cli.Command{
		Name: "ecosystems",
		Type: cli.Output,
		RunCLI: func(args []string) error {
			cmd.RunEcosystems(args)
			return nil
		},
	})

	r.Register(&cli.Command{
		Name: "integrations",
		Type: cli.Output,
		RunCLI: func(args []string) error {
			cmd.RunIntegrations()
			return nil
		},
	})

	r.Register(&cli.Command{
		Name:        "cleanup",
		Type:        cli.Decision,
		Destructive: true,
		ParseFlags: func(args []string) (any, error) {
			return cmd.ParseCleanupFlags(args)
		},
		RunCLI: func(args []string) error {
			opts, err := cmd.ParseCleanupFlags(args)
			if err != nil {
				return err
			}
			if err := cmd.ValidateCleanupForNonInteractive(opts); err != nil {
				return err
			}
			cmd.RunCleanup(args)
			return nil
		},
	})

	return r
}

func main() {
	registry := buildRegistry()

	result, err := registry.Dispatch(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if cli.IsUnknownCommand(err) {
			fmt.Fprint(os.Stderr, registry.UsageString())
		}
		os.Exit(1)
	}

	// Dispatch to registered commands.
	if result.Command != nil {
		switch result.Command.Type {
		case cli.Output:
			if err := result.Command.RunCLI(result.Args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return

		case cli.Decision:
			if result.NonInteractive {
				if err := result.Command.RunCLI(result.Args); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			}
			// TODO: For decision commands without --non-interactive,
			// parse flags and launch TUI at the appropriate entry point.
			// For now, fall through to RunCLI (existing behavior).
			if err := result.Command.RunCLI(result.Args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Root: no command specified — handle global flags and launch TUI.
	runRoot(result.Args)
}

func runRoot(args []string) {
	fs := flag.NewFlagSet("sentei", flag.ExitOnError)
	versionFlag := fs.Bool("version", false, "Print version and exit")
	playgroundFlag := fs.Bool("playground", false, "Launch with a temporary test repo")
	dryRunFlag := fs.Bool("dry-run", false, "Print worktree summary and exit (no interactive TUI)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *versionFlag {
		fmt.Printf("sentei %s (%s, %s)\n", version, commit, date)
		os.Exit(0)
	}

	repoPath := "."
	if fs.NArg() > 0 {
		repoPath = fs.Arg(0)
	}
	if absPath, err := filepath.Abs(repoPath); err == nil {
		repoPath = absPath
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

	context := repo.DetectContext(runner, repoPath)
	if context == repo.ContextBareRepo {
		repoPath = repo.ResolveBareRoot(runner, repoPath)
	}

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

	model := tui.NewMenuModel(tuiRunner, shell, repoPath, cfg, context)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
