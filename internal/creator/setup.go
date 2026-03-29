package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
)

func SanitizeBranchPath(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func runSetup(runner git.CommandRunner, opts Options, emit func(Event)) Phase {
	phase := Phase{Name: "Setup"}

	wtResult, wtPath := createWorktreeStep(runner, opts.RepoPath, opts.BranchName, opts.BaseBranch, emit)
	phase.Steps = append(phase.Steps, wtResult)

	if wtResult.Status == StepFailed {
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
		phase.Steps = append(phase.Steps, StepResult{
			Name:   "Copy env files",
			Status: StepSkipped,
		})
	}

	return phase
}

func createWorktreeStep(runner git.CommandRunner, repoPath, branch, baseBranch string, emit func(Event)) (StepResult, string) {
	stepName := "Create worktree"
	emit(Event{Phase: "Setup", Step: stepName, Status: StepRunning})

	sanitized := SanitizeBranchPath(branch)
	wtPath := filepath.Join(repoPath, sanitized)

	_, err := runner.Run(repoPath, "worktree", "add", wtPath, "-b", branch, baseBranch)
	if err != nil {
		emit(Event{Phase: "Setup", Step: stepName, Status: StepFailed, Error: err})
		return StepResult{
			Name:   stepName,
			Status: StepFailed,
			Error:  fmt.Errorf("creating worktree: %w", err),
		}, ""
	}

	emit(Event{Phase: "Setup", Step: stepName, Status: StepDone, Message: wtPath})
	return StepResult{
		Name:    stepName,
		Status:  StepDone,
		Message: wtPath,
	}, wtPath
}

func mergeBaseStep(runner git.CommandRunner, wtPath, baseBranch string, enabled bool, emit func(Event)) StepResult {
	stepName := "Merge base branch"

	if !enabled {
		return StepResult{Name: stepName, Status: StepSkipped}
	}

	emit(Event{Phase: "Setup", Step: stepName, Status: StepRunning})

	_, err := runner.Run(wtPath, "merge", baseBranch, "--no-edit")
	if err != nil {
		emit(Event{Phase: "Setup", Step: stepName, Status: StepFailed, Error: err, Message: "merge conflict — resolve manually"})
		return StepResult{
			Name:    stepName,
			Status:  StepFailed,
			Message: "merge conflict — resolve manually",
			Error:   err,
		}
	}

	emit(Event{Phase: "Setup", Step: stepName, Status: StepDone})
	return StepResult{Name: stepName, Status: StepDone}
}

func copyEnvFilesStep(srcDir, dstDir string, envFiles []string, emit func(Event)) StepResult {
	stepName := "Copy env files"

	if len(envFiles) == 0 {
		return StepResult{Name: stepName, Status: StepSkipped}
	}

	emit(Event{Phase: "Setup", Step: stepName, Status: StepRunning})

	var copied []string
	for _, name := range envFiles {
		src := filepath.Join(srcDir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		dst := filepath.Join(dstDir, name)
		if err := fileutil.CopyFile(src, dst); err != nil {
			emit(Event{Phase: "Setup", Step: stepName, Status: StepFailed, Error: err})
			return StepResult{
				Name:   stepName,
				Status: StepFailed,
				Error:  fmt.Errorf("copying %s: %w", name, err),
			}
		}
		copied = append(copied, name)
	}

	msg := strings.Join(copied, ", ")
	if msg == "" {
		msg = "no source files found"
	}
	emit(Event{Phase: "Setup", Step: stepName, Status: StepDone, Message: msg})
	return StepResult{
		Name:    stepName,
		Status:  StepDone,
		Message: msg,
	}
}
