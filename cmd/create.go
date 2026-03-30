package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/creator"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
)

// RunCreate executes the create worktree command in non-interactive mode.
func RunCreate(args []string) error {
	opts, err := ParseCreateFlags(args)
	if err != nil {
		return err
	}
	if err := ValidateCreateForNonInteractive(opts); err != nil {
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
	shell := &git.DefaultShellRunner{}

	context := repo.DetectContext(runner, repoPath)
	if context != repo.ContextBareRepo {
		return fmt.Errorf("create requires a bare repository (detected: %v)", context)
	}

	// Build creator.Options from CLI flags.
	creatorOpts := creator.Options{
		BranchName:   opts.Branch,
		BaseBranch:   opts.Base,
		RepoPath:     repoPath,
		MergeBase:    opts.MergeBase,
		CopyEnvFiles: opts.CopyEnv,
	}

	// Resolve ecosystems from config if requested.
	if len(opts.Ecosystems) > 0 {
		cfg, err := config.LoadConfig(repoPath,
			config.WithRunner(runner),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		}
		if cfg != nil {
			creatorOpts.Ecosystems = matchEcosystems(cfg.Ecosystems, opts.Ecosystems)
		}
	}

	// Find a source worktree for env file copying.
	if opts.CopyEnv {
		worktrees, err := git.ListWorktrees(runner, repoPath)
		if err == nil {
			creatorOpts.SourceWorktree = findSource(worktrees)
		}
	}

	fmt.Printf("Creating worktree %q from %s...\n", opts.Branch, opts.Base)

	result := creator.Run(runner, shell, creatorOpts, func(e creator.Event) {
		printCreateEvent(e)
	})

	if result.HasFailures() {
		return fmt.Errorf("create completed with errors")
	}

	fmt.Printf("\n%sWorktree created:%s %s\n", green, nc, result.WorktreePath)
	return nil
}

func printCreateEvent(e creator.Event) {
	switch e.Status {
	case creator.StepRunning:
		fmt.Printf("%s→%s %s: %s\n", blue, nc, e.Phase, e.Step)
	case creator.StepDone:
		msg := ""
		if e.Message != "" {
			msg = " — " + e.Message
		}
		fmt.Printf("%s✓%s %s%s\n", green, nc, e.Step, msg)
	case creator.StepFailed:
		msg := ""
		if e.Error != nil {
			msg = " — " + e.Error.Error()
		}
		fmt.Printf("%s✗%s %s%s\n", yellow, nc, e.Step, msg)
	case creator.StepSkipped:
		fmt.Printf("  %s%s (skipped)%s\n", dim, e.Step, nc)
	}
}

// matchEcosystems filters configured ecosystems by the names requested via CLI.
func matchEcosystems(available []config.EcosystemConfig, requested []string) []config.EcosystemConfig {
	want := make(map[string]bool, len(requested))
	for _, name := range requested {
		want[name] = true
	}
	var matched []config.EcosystemConfig
	for _, eco := range available {
		if want[eco.Name] {
			matched = append(matched, eco)
		}
	}
	return matched
}

// findSource picks a source worktree for env file copying (prefers main/master).
func findSource(worktrees []git.Worktree) string {
	for _, wt := range worktrees {
		branch := strings.TrimPrefix(wt.Branch, "refs/heads/")
		if branch == "main" || branch == "master" {
			return wt.Path
		}
	}
	if len(worktrees) > 0 {
		return worktrees[0].Path
	}
	return ""
}
