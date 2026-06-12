package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
)

func runSetup(runner git.CommandRunner, opts Options, emit func(progress.Event)) progress.Phase {
	phase := progress.Phase{Name: "Setup"}

	// Declare the steps that will certainly emit (all decidable from opts),
	// then close: the Setup step set is final from the first frame.
	planned := []progress.PlannedStep{{Name: "Create worktree"}}
	if opts.MergeBase {
		planned = append(planned, progress.PlannedStep{Name: "Merge base branch"})
	}
	if opts.CopyEnvFiles && setupHasEnvFiles(opts) {
		planned = append(planned, progress.PlannedStep{Name: "Copy env files"})
	}
	progress.Declare(progress.Plan{Phases: []progress.PlannedPhase{{Name: "Setup", Steps: planned}}}, emit)

	wtResult, wtPath := createWorktreeStep(runner, opts.RepoPath, opts.BranchName, opts.BaseBranch, emit)
	phase.Steps = append(phase.Steps, wtResult)

	if wtResult.Status == progress.StepFailed {
		return phase
	}

	mergeResult := mergeBaseStep(runner, wtPath, opts.BaseBranch, opts.MergeBase, emit)
	phase.Steps = append(phase.Steps, mergeResult)

	var envFiles []string
	for _, eco := range opts.Ecosystems {
		envFiles = append(envFiles, eco.EnvFiles...)
	}
	if opts.CopyEnvFiles {
		envResult := copyEnvFilesStep(opts.SourceWorktree, wtPath, envFiles, emit)
		phase.Steps = append(phase.Steps, envResult)
	} else {
		phase.Steps = append(phase.Steps, progress.StepResult{
			Name:   "Copy env files",
			Status: progress.StepSkipped,
		})
	}

	return phase
}

// setupHasEnvFiles reports whether any detected ecosystem lists env files,
// the condition under which the copy step will actually emit.
func setupHasEnvFiles(opts Options) bool {
	for _, eco := range opts.Ecosystems {
		if len(eco.EnvFiles) > 0 {
			return true
		}
	}
	return false
}

func createWorktreeStep(runner git.CommandRunner, repoPath, branch, baseBranch string, emit func(progress.Event)) (progress.StepResult, string) {
	wtPath := git.WorktreePath(repoPath, branch)

	result := progress.RunStep("Setup", "Create worktree", emit, func() (string, error) {
		var err error
		if git.BranchExists(runner, repoPath, branch) {
			_, err = runner.Run(repoPath, "worktree", "add", wtPath, branch)
		} else {
			_, err = runner.Run(repoPath, "worktree", "add", wtPath, "-b", branch, baseBranch)
		}
		if err != nil {
			return "", fmt.Errorf("creating worktree: %w", err)
		}
		return wtPath, nil
	})
	if result.Status == progress.StepFailed {
		return result, ""
	}
	return result, wtPath
}

func mergeBaseStep(runner git.CommandRunner, wtPath, baseBranch string, enabled bool, emit func(progress.Event)) progress.StepResult {
	stepName := "Merge base branch"

	if !enabled {
		return progress.StepResult{Name: stepName, Status: progress.StepSkipped}
	}

	emit(progress.Event{Phase: "Setup", Step: stepName, Status: progress.StepRunning})

	_, err := runner.Run(wtPath, "merge", baseBranch, "--no-edit")
	if err != nil {
		emit(progress.Event{Phase: "Setup", Step: stepName, Status: progress.StepFailed, Error: err, Message: "merge conflict — resolve manually"})
		return progress.StepResult{
			Name:    stepName,
			Status:  progress.StepFailed,
			Message: "merge conflict — resolve manually",
			Error:   err,
		}
	}

	emit(progress.Event{Phase: "Setup", Step: stepName, Status: progress.StepDone})
	return progress.StepResult{Name: stepName, Status: progress.StepDone}
}

func copyEnvFilesStep(srcDir, dstDir string, envFiles []string, emit func(progress.Event)) progress.StepResult {
	stepName := "Copy env files"

	if len(envFiles) == 0 {
		return progress.StepResult{Name: stepName, Status: progress.StepSkipped}
	}

	return progress.RunStep("Setup", stepName, emit, func() (string, error) {
		var copied []string
		for _, name := range envFiles {
			src := filepath.Join(srcDir, name)
			if _, err := os.Stat(src); os.IsNotExist(err) {
				continue
			}

			if err := fileutil.CopyFile(src, filepath.Join(dstDir, name)); err != nil {
				return "", fmt.Errorf("copying %s: %w", name, err)
			}
			copied = append(copied, name)
		}

		if len(copied) == 0 {
			return "no source files found", nil
		}
		return strings.Join(copied, ", "), nil
	})
}
