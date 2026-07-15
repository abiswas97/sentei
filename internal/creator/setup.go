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
