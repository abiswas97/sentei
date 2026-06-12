package creator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/progress"
)

func runIntegrations(shell git.ShellRunner, wtPath string, opts Options, emit func(progress.Event)) progress.Phase {
	phase := progress.Phase{Name: "Integrations"}

	if len(opts.Integrations) == 0 {
		return phase
	}

	for _, integ := range opts.Integrations {
		steps := setupIntegration(shell, wtPath, opts.RepoPath, opts.SourceWorktree, integ, emit)
		phase.Steps = append(phase.Steps, steps...)
	}

	return phase
}

func setupIntegration(shell git.ShellRunner, wtPath, repoPath, sourceWorktree string, integ integration.Integration, emit func(progress.Event)) []progress.StepResult {
	var steps []progress.StepResult

	if integ.IndexCopyDir != "" && sourceWorktree != "" {
		stepName := "Copy index from main"
		emit(progress.Event{Phase: "Integrations", Step: stepName, Status: progress.StepRunning})
		if err := copyIntegrationIndex(sourceWorktree, wtPath, integ.IndexCopyDir); err != nil {
			emit(progress.Event{Phase: "Integrations", Step: stepName, Status: progress.StepSkipped, Message: err.Error()})
		} else {
			emit(progress.Event{Phase: "Integrations", Step: stepName, Status: progress.StepDone})
			steps = append(steps, progress.StepResult{Name: stepName, Status: progress.StepDone})
		}
	}

	installed := detectIntegration(shell, wtPath, integ)

	if !installed {
		depSteps := checkAndInstallDeps(shell, wtPath, integ, emit)
		steps = append(steps, depSteps...)

		for _, s := range depSteps {
			if s.Status == progress.StepFailed {
				return steps
			}
		}

		installStep := installIntegration(shell, wtPath, integ, emit)
		steps = append(steps, installStep)
		if installStep.Status == progress.StepFailed {
			return steps
		}
	}

	setupStep := runSetupCommand(shell, wtPath, repoPath, integ, emit)
	steps = append(steps, setupStep)

	if setupStep.Status != progress.StepFailed && len(integ.GitignoreEntries) > 0 {
		if err := appendGitignore(wtPath, integ.GitignoreEntries); err != nil {
			gitignoreStep := fmt.Sprintf("Gitignore %s", integ.Name)
			emit(progress.Event{Phase: "Integrations", Step: gitignoreStep, Status: progress.StepFailed, Error: err})
			// Record the failure so HasFailures() and the summary reflect it; an
			// emitted event alone is invisible to the result.
			steps = append(steps, progress.StepResult{Name: gitignoreStep, Status: progress.StepFailed, Error: err})
		}
	}

	return steps
}

func detectIntegration(shell git.ShellRunner, wtPath string, integ integration.Integration) bool {
	if integ.Detect.Command != "" {
		_, err := shell.RunShell(wtPath, integ.Detect.Command)
		return err == nil
	}
	if integ.Detect.BinaryName != "" {
		if _, err := exec.LookPath(integ.Detect.BinaryName); err == nil {
			return true
		}
	}
	return false
}

func checkAndInstallDeps(shell git.ShellRunner, wtPath string, integ integration.Integration, emit func(progress.Event)) []progress.StepResult {
	var steps []progress.StepResult

	for _, dep := range integ.Dependencies {
		stepName := fmt.Sprintf("Check %s", dep.Name)
		emit(progress.Event{Phase: "Integrations", Step: stepName, Status: progress.StepRunning})

		_, err := shell.RunShell(wtPath, dep.Detect)
		if err == nil {
			emit(progress.Event{Phase: "Integrations", Step: stepName, Status: progress.StepDone})
			steps = append(steps, progress.StepResult{Name: stepName, Status: progress.StepDone})
			continue
		}

		if dep.Install == "" {
			emit(progress.Event{Phase: "Integrations", Step: stepName, Status: progress.StepFailed, Error: fmt.Errorf("%s not found and no install command available", dep.Name)})
			steps = append(steps, progress.StepResult{
				Name:   stepName,
				Status: progress.StepFailed,
				Error:  fmt.Errorf("%s not found and no install command available", dep.Name),
			})
			return steps
		}

		installName := fmt.Sprintf("Install %s", dep.Name)
		emit(progress.Event{Phase: "Integrations", Step: installName, Status: progress.StepRunning})
		_, installErr := shell.RunShell(wtPath, dep.Install)
		if installErr != nil {
			emit(progress.Event{Phase: "Integrations", Step: installName, Status: progress.StepFailed, Error: installErr})
			steps = append(steps, progress.StepResult{
				Name:   installName,
				Status: progress.StepFailed,
				Error:  fmt.Errorf("installing dependency %s: %w", dep.Name, installErr),
			})
			return steps
		}

		emit(progress.Event{Phase: "Integrations", Step: installName, Status: progress.StepDone})
		steps = append(steps, progress.StepResult{Name: installName, Status: progress.StepDone})
	}

	return steps
}

func installIntegration(shell git.ShellRunner, wtPath string, integ integration.Integration, emit func(progress.Event)) progress.StepResult {
	return progress.RunStep("Integrations", fmt.Sprintf("Install %s", integ.Name), emit, func() (string, error) {
		if _, err := shell.RunShell(wtPath, integ.Install.Command); err != nil {
			return "", fmt.Errorf("installing %s: %w", integ.Name, err)
		}
		return "", nil
	})
}

func runSetupCommand(shell git.ShellRunner, wtPath, repoPath string, integ integration.Integration, emit func(progress.Event)) progress.StepResult {
	stepName := fmt.Sprintf("Setup %s", integ.Name)

	if integ.Setup.Command == "" {
		return progress.StepResult{Name: stepName, Status: progress.StepSkipped}
	}

	return progress.RunStep("Integrations", stepName, emit, func() (string, error) {
		// The worktree path embeds the branch name and is interpolated into a command
		// run via `sh -c`; quote it so a branch like "a&&rm -rf x" cannot inject.
		command := strings.ReplaceAll(integ.Setup.Command, "{path}", git.ShellQuote(wtPath))

		runDir := wtPath
		if integ.Setup.WorkingDir == "repo" {
			runDir = repoPath
		}

		if _, err := shell.RunShell(runDir, command); err != nil {
			return "", fmt.Errorf("setting up %s: %w", integ.Name, err)
		}
		return "", nil
	})
}

// copyIntegrationIndex copies the IndexCopyDir from source to target worktree.
func copyIntegrationIndex(sourceWT, targetWT, indexDir string) error {
	srcDir := filepath.Join(sourceWT, indexDir)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("no index at %s", srcDir)
	}

	dstDir := filepath.Join(targetWT, indexDir)
	_ = os.RemoveAll(dstDir)

	return fileutil.CopyDir(srcDir, dstDir)
}

func appendGitignore(dir string, entries []string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")

	existing, _ := os.ReadFile(gitignorePath)
	content := string(existing)

	var toAdd []string
	for _, entry := range entries {
		if !strings.Contains(content, entry) {
			toAdd = append(toAdd, entry)
		}
	}

	if len(toAdd) == 0 {
		return nil
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening .gitignore: %w", err)
	}
	defer func() { _ = f.Close() }()

	for _, entry := range toAdd {
		if _, err := fmt.Fprintln(f, entry); err != nil {
			return fmt.Errorf("writing to .gitignore: %w", err)
		}
	}
	return nil
}
