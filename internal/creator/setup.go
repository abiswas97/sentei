package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/pipeline"
)

func runSetup(runner git.CommandRunner, opts Options, emit func(pipeline.Event)) pipeline.Phase {
	phase := pipeline.Phase{Name: "Setup"}

	wtResult, wtPath := createWorktreeStep(runner, opts.RepoPath, opts.BranchName, opts.BaseBranch, emit)
	phase.Steps = append(phase.Steps, wtResult)

	if wtResult.Status == pipeline.StepFailed {
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
		phase.Steps = append(phase.Steps, pipeline.StepResult{
			Name:   "Copy env files",
			Status: pipeline.StepSkipped,
		})
	}

	return phase
}

func createWorktreeStep(runner git.CommandRunner, repoPath, branch, baseBranch string, emit func(pipeline.Event)) (pipeline.StepResult, string) {
	stepName := "Create worktree"
	emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepRunning})

	wtPath := git.WorktreePath(repoPath, branch)

	var err error
	if git.BranchExists(runner, repoPath, branch) {
		_, err = runner.Run(repoPath, "worktree", "add", wtPath, branch)
	} else {
		_, err = runner.Run(repoPath, "worktree", "add", wtPath, "-b", branch, baseBranch)
	}
	if err != nil {
		emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepFailed, Error: err})
		return pipeline.StepResult{
			Name:   stepName,
			Status: pipeline.StepFailed,
			Error:  fmt.Errorf("creating worktree: %w", err),
		}, ""
	}

	emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepDone, Message: wtPath})
	return pipeline.StepResult{
		Name:    stepName,
		Status:  pipeline.StepDone,
		Message: wtPath,
	}, wtPath
}

func mergeBaseStep(runner git.CommandRunner, wtPath, baseBranch string, enabled bool, emit func(pipeline.Event)) pipeline.StepResult {
	stepName := "Merge base branch"

	if !enabled {
		return pipeline.StepResult{Name: stepName, Status: pipeline.StepSkipped}
	}

	emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepRunning})

	_, err := runner.Run(wtPath, "merge", baseBranch, "--no-edit")
	if err != nil {
		emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepFailed, Error: err, Message: "merge conflict — resolve manually"})
		return pipeline.StepResult{
			Name:    stepName,
			Status:  pipeline.StepFailed,
			Message: "merge conflict — resolve manually",
			Error:   err,
		}
	}

	emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepDone})
	return pipeline.StepResult{Name: stepName, Status: pipeline.StepDone}
}

func copyEnvFilesStep(srcDir, dstDir string, envFiles []string, emit func(pipeline.Event)) pipeline.StepResult {
	stepName := "Copy env files"

	if len(envFiles) == 0 {
		return pipeline.StepResult{Name: stepName, Status: pipeline.StepSkipped}
	}

	emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepRunning})

	var copied []string
	for _, name := range envFiles {
		src := filepath.Join(srcDir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		dst := filepath.Join(dstDir, name)
		if err := fileutil.CopyFile(src, dst); err != nil {
			emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepFailed, Error: err})
			return pipeline.StepResult{
				Name:   stepName,
				Status: pipeline.StepFailed,
				Error:  fmt.Errorf("copying %s: %w", name, err),
			}
		}
		copied = append(copied, name)
	}

	msg := strings.Join(copied, ", ")
	if msg == "" {
		msg = "no source files found"
	}
	emit(pipeline.Event{Phase: "Setup", Step: stepName, Status: pipeline.StepDone, Message: msg})
	return pipeline.StepResult{
		Name:    stepName,
		Status:  pipeline.StepDone,
		Message: msg,
	}
}
