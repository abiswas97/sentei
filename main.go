package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/cmd"
	"github.com/abiswas97/sentei/internal/cleanup"
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
		Name:        "clone",
		Type:        cli.Decision,
		Destructive: false,
		ParseFlags: func(args []string) (any, error) {
			return cmd.ParseCloneFlags(args)
		},
		RunCLI: func(args []string) error {
			return cmd.RunClone(args)
		},
	})

	r.Register(&cli.Command{
		Name: "create",
		Type: cli.Decision,
		ParseFlags: func(args []string) (any, error) {
			return cmd.ParseCreateFlags(args)
		},
		RunCLI: func(args []string) error {
			return cmd.RunCreate(args)
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

	r.Register(&cli.Command{
		Name:        "migrate",
		Type:        cli.Decision,
		Destructive: true,
		ParseFlags: func(args []string) (any, error) {
			return cmd.ParseMigrateFlags(args)
		},
		RunCLI: func(args []string) error {
			return cmd.RunMigrate(args)
		},
	})

	r.Register(&cli.Command{
		Name:        "remove",
		Type:        cli.Decision,
		Destructive: true,
		ParseFlags: func(args []string) (any, error) {
			return cmd.ParseRemoveFlags(args)
		},
		RunCLI: func(args []string) error {
			return cmd.RunRemove(args)
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
			// Interactive mode: parse flags and launch TUI at the appropriate view.
			launchInteractiveDecision(*result)
			return
		}
	}

	// Root: no command specified — handle global flags and launch TUI.
	runRoot(result.Args)
}

func launchInteractiveDecision(result cli.DispatchResult) {
	repoPath := "."
	if absPath, err := filepath.Abs(repoPath); err == nil {
		repoPath = absPath
	}

	runner := &git.GitRunner{}
	shell := &git.DefaultShellRunner{}

	context := repo.DetectContext(runner, repoPath)
	if context == repo.ContextBareRepo {
		repoPath = repo.ResolveBareRoot(runner, repoPath)
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

	model := tui.NewMenuModel(runner, shell, repoPath, cfg, context)

	switch result.Command.Name {
	case "cleanup":
		opts, err := cmd.ParseCleanupFlags(result.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if opts.Mode == "" {
			opts.Mode = cleanup.ModeSafe
		}
		model.SetCleanupOpts(opts)

	case "create":
		opts, err := cmd.ParseCreateFlags(result.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		model.SetCreateOpts(opts)

	case "clone":
		opts, err := cmd.ParseCloneFlags(result.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		model.SetCloneOpts(opts)

	case "remove":
		opts, err := cmd.ParseRemoveFlags(result.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if opts.Merged || opts.All || opts.Stale > 0 {
			worktrees, err := git.ListWorktrees(runner, repoPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			worktrees = worktree.EnrichWorktrees(runner, worktrees, 10)

			var isMerged cmd.MergedChecker
			if opts.Merged {
				defaultBranch := cmd.DetectDefaultBranch(runner, repoPath)
				isMerged = cmd.CheckMerged(runner, repoPath, defaultBranch)
			}
			filtered := cmd.ResolveFilters(worktrees, opts, nil, isMerged)

			var paths []string
			for _, wt := range filtered {
				paths = append(paths, wt.Path)
			}
			model.SetRemoveOpts(tui.RemovePreSelection{
				Paths:       paths,
				FilterLabel: cmd.FormatFilterLabel(opts),
			})
		}

	case "migrate":
		opts, err := cmd.ParseMigrateFlags(result.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		model.SetMigrateOpts(&tui.MigrateOpts{
			DeleteBackup: opts.DeleteBackup,
			RepoPath:     opts.RepoPath,
		})
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
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
