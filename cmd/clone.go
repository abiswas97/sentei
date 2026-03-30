package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
)

// RunClone executes the clone command in non-interactive mode.
func RunClone(args []string) error {
	opts, err := ParseCloneFlags(args)
	if err != nil {
		return err
	}

	if err := ValidateCloneForNonInteractive(opts); err != nil {
		return err
	}

	name := opts.Name
	if name == "" {
		name = repo.DeriveRepoName(opts.URL)
	}

	location, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	runner := &git.GitRunner{}
	cloneOpts := repo.CloneOptions{
		URL:      opts.URL,
		Location: location,
		Name:     name,
	}

	result := repo.Clone(runner, cloneOpts, printCloneEvent)

	fmt.Println()
	for _, phase := range result.Phases {
		if phase.HasFailures() {
			for _, step := range phase.Steps {
				if step.Status == repo.StepFailed {
					fmt.Fprintf(os.Stderr, "%s✗%s %s: %v\n", yellow, nc, step.Name, step.Error)
				}
			}
			return fmt.Errorf("clone failed")
		}
	}

	fmt.Printf("%s✓%s Cloned to %s\n", green, nc, filepath.Join(location, name))
	return nil
}

func printCloneEvent(e repo.Event) {
	switch e.Status {
	case repo.StepRunning:
		fmt.Printf("%s→%s [%s] %s\n", blue, nc, e.Phase, e.Step)
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
