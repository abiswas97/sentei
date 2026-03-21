package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
)

const (
	green  = "\033[0;32m"
	yellow = "\033[1;33m"
	blue   = "\033[0;34m"
	dim    = "\033[2m"
	nc     = "\033[0m"
)

func RunCleanup(args []string) {
	fs := flag.NewFlagSet("cleanup", flag.ExitOnError)
	mode := fs.String("mode", "safe", "Cleanup mode: safe or aggressive")
	force := fs.Bool("force", false, "Force-delete unmerged branches (aggressive mode)")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without making changes")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sentei cleanup [options] [repo-path]\n\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	repoPath := "."
	if fs.NArg() > 0 {
		repoPath = fs.Arg(0)
	}

	m := cleanup.Mode(*mode)
	if m != cleanup.ModeSafe && m != cleanup.ModeAggressive {
		fmt.Fprintf(os.Stderr, "Unknown mode: %s (use safe or aggressive)\n", *mode)
		os.Exit(1)
	}

	opts := cleanup.Options{
		Mode:   m,
		Force:  *force,
		DryRun: *dryRun,
	}

	if opts.DryRun {
		fmt.Printf("%s(dry run)%s\n", dim, nc)
	}
	fmt.Println()

	runner := &git.GitRunner{}
	result := cleanup.Run(runner, repoPath, opts, printEvent)

	fmt.Println()
	for _, e := range result.Errors {
		fmt.Printf("%s⚠%s  %s: %s\n", yellow, nc, e.Step, e.Err)
	}

	if result.NonWtBranchesRemaining > 0 && opts.Mode == cleanup.ModeSafe {
		fmt.Printf("\n%sTip:%s %d local branch(es) are not checked out in any worktree.\n", blue, nc, result.NonWtBranchesRemaining)
		fmt.Printf("     Run %ssentei cleanup --mode=aggressive%s to remove them.\n", dim, nc)
	}
}

func printEvent(e cleanup.Event) {
	switch e.Level {
	case cleanup.LevelStep:
		fmt.Printf("%s→%s %s\n", blue, nc, e.Message)
	case cleanup.LevelInfo:
		fmt.Printf("%s✓%s %s\n", green, nc, e.Message)
	case cleanup.LevelWarn:
		fmt.Printf("%s⚠%s  %s\n", yellow, nc, e.Message)
	case cleanup.LevelDetail:
		fmt.Printf("  %s%s%s\n", dim, e.Message, nc)
	}
}
