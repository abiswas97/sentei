package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
)

// RunMigrate executes the migrate command in non-interactive mode.
func RunMigrate(args []string) error {
	opts, err := ParseMigrateFlags(args)
	if err != nil {
		return err
	}
	if err := ValidateMigrateForNonInteractive(opts); err != nil {
		return err
	}

	repoPath := opts.RepoPath
	if absPath, err := filepath.Abs(repoPath); err == nil {
		repoPath = absPath
	}

	runner := &git.GitRunner{}
	shell := &git.DefaultShellRunner{}

	context := repo.DetectContext(runner, repoPath)
	if context == repo.ContextBareRepo {
		return fmt.Errorf("repository is already bare: %s", repoPath)
	}
	if context != repo.ContextNonBareRepo {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	migrateOpts := repo.MigrateOptions{
		RepoPath: repoPath,
	}

	result := repo.Migrate(runner, shell, migrateOpts, printMigrateEvent)

	fmt.Println()
	for _, phase := range result.Phases {
		if phase.HasFailures() {
			for _, step := range phase.Steps {
				if step.Status == repo.StepFailed {
					fmt.Fprintf(os.Stderr, "%s✗%s %s: %v\n", yellow, nc, step.Name, step.Error)
				}
			}
			return fmt.Errorf("migration failed during %s phase", phase.Name)
		}
	}

	fmt.Printf("%s✓%s Migration complete\n", green, nc)
	fmt.Printf("  Bare root:  %s\n", result.BareRoot)
	fmt.Printf("  Worktree:   %s\n", result.WorktreePath)
	fmt.Printf("  Backup:     %s (%s)\n", result.BackupPath, result.BackupSize)

	if opts.DeleteBackup && result.BackupPath != "" {
		fmt.Printf("\n%s→%s Deleting backup...\n", blue, nc)
		if err := repo.DeleteBackup(result.BackupPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s⚠%s  Failed to delete backup: %v\n", yellow, nc, err)
		} else {
			fmt.Printf("%s✓%s Backup deleted\n", green, nc)
		}
	}

	return nil
}

func printMigrateEvent(e repo.Event) {
	switch e.Status {
	case repo.StepRunning:
		msg := ""
		if e.Message != "" {
			msg = fmt.Sprintf(" — %s", e.Message)
		}
		fmt.Printf("%s→%s [%s] %s%s\n", blue, nc, e.Phase, e.Step, msg)
	case repo.StepDone:
		msg := ""
		if e.Message != "" {
			msg = fmt.Sprintf(" (%s)", e.Message)
		}
		fmt.Printf("%s✓%s [%s] %s%s\n", green, nc, e.Phase, e.Step, msg)
	case repo.StepFailed:
		fmt.Printf("%s✗%s [%s] %s: %v\n", yellow, nc, e.Phase, e.Step, e.Error)
	}
}
