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
)

func runIntegrations(shell git.ShellRunner, wtPath string, opts Options, emit func(Event)) Phase {
	phase := Phase{Name: "Integrations"}

	if len(opts.Integrations) == 0 {
		return phase
	}

	for _, integ := range opts.Integrations {
		steps := setupIntegration(shell, wtPath, opts.RepoPath, opts.SourceWorktree, integ, emit)
		phase.Steps = append(phase.Steps, steps...)
	}

	return phase
}

func setupIntegration(shell git.ShellRunner, wtPath, repoPath, sourceWorktree string, integ integration.Integration, emit func(Event)) []StepResult {
	var steps []StepResult

	if integ.IndexCopyDir != "" && sourceWorktree != "" {
		stepName := "Copy index from main"
		emit(Event{Phase: "Integrations", Step: stepName, Status: StepRunning})
		if err := copyIntegrationIndex(sourceWorktree, wtPath, integ.IndexCopyDir); err != nil {
			emit(Event{Phase: "Integrations", Step: stepName, Status: StepSkipped, Message: err.Error()})
		} else {
			emit(Event{Phase: "Integrations", Step: stepName, Status: StepDone})
			steps = append(steps, StepResult{Name: stepName, Status: StepDone})
		}
	}

	installed := detectIntegration(shell, wtPath, integ)

	if !installed {
		depSteps := checkAndInstallDeps(shell, wtPath, integ, emit)
		steps = append(steps, depSteps...)

		for _, s := range depSteps {
			if s.Status == StepFailed {
				return steps
			}
		}

		installStep := installIntegration(shell, wtPath, integ, emit)
		steps = append(steps, installStep)
		if installStep.Status == StepFailed {
			return steps
		}
	}

	setupStep := runSetupCommand(shell, wtPath, repoPath, integ, emit)
	steps = append(steps, setupStep)

	if setupStep.Status != StepFailed && len(integ.GitignoreEntries) > 0 {
		if err := appendGitignore(wtPath, integ.GitignoreEntries); err != nil {
			emit(Event{Phase: "Integrations", Step: fmt.Sprintf("Gitignore %s", integ.Name), Status: StepFailed, Error: err})
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

func checkAndInstallDeps(shell git.ShellRunner, wtPath string, integ integration.Integration, emit func(Event)) []StepResult {
	var steps []StepResult

	for _, dep := range integ.Dependencies {
		stepName := fmt.Sprintf("Check %s", dep.Name)
		emit(Event{Phase: "Integrations", Step: stepName, Status: StepRunning})

		_, err := shell.RunShell(wtPath, dep.Detect)
		if err == nil {
			emit(Event{Phase: "Integrations", Step: stepName, Status: StepDone})
			steps = append(steps, StepResult{Name: stepName, Status: StepDone})
			continue
		}

		if dep.Install == "" {
			emit(Event{Phase: "Integrations", Step: stepName, Status: StepFailed, Error: fmt.Errorf("%s not found and no install command available", dep.Name)})
			steps = append(steps, StepResult{
				Name:   stepName,
				Status: StepFailed,
				Error:  fmt.Errorf("%s not found and no install command available", dep.Name),
			})
			return steps
		}

		installName := fmt.Sprintf("Install %s", dep.Name)
		emit(Event{Phase: "Integrations", Step: installName, Status: StepRunning})
		_, installErr := shell.RunShell(wtPath, dep.Install)
		if installErr != nil {
			emit(Event{Phase: "Integrations", Step: installName, Status: StepFailed, Error: installErr})
			steps = append(steps, StepResult{
				Name:   installName,
				Status: StepFailed,
				Error:  fmt.Errorf("installing dependency %s: %w", dep.Name, installErr),
			})
			return steps
		}

		emit(Event{Phase: "Integrations", Step: installName, Status: StepDone})
		steps = append(steps, StepResult{Name: installName, Status: StepDone})
	}

	return steps
}

func installIntegration(shell git.ShellRunner, wtPath string, integ integration.Integration, emit func(Event)) StepResult {
	stepName := fmt.Sprintf("Install %s", integ.Name)
	emit(Event{Phase: "Integrations", Step: stepName, Status: StepRunning})

	_, err := shell.RunShell(wtPath, integ.Install.Command)
	if err != nil {
		emit(Event{Phase: "Integrations", Step: stepName, Status: StepFailed, Error: err})
		return StepResult{
			Name:   stepName,
			Status: StepFailed,
			Error:  fmt.Errorf("installing %s: %w", integ.Name, err),
		}
	}

	emit(Event{Phase: "Integrations", Step: stepName, Status: StepDone})
	return StepResult{Name: stepName, Status: StepDone}
}

func runSetupCommand(shell git.ShellRunner, wtPath, repoPath string, integ integration.Integration, emit func(Event)) StepResult {
	stepName := fmt.Sprintf("Setup %s", integ.Name)

	if integ.Setup.Command == "" {
		return StepResult{Name: stepName, Status: StepSkipped}
	}

	emit(Event{Phase: "Integrations", Step: stepName, Status: StepRunning})

	command := strings.ReplaceAll(integ.Setup.Command, "{path}", wtPath)

	var runDir string
	switch integ.Setup.WorkingDir {
	case "repo":
		runDir = repoPath
	default:
		runDir = wtPath
	}

	_, err := shell.RunShell(runDir, command)
	if err != nil {
		emit(Event{Phase: "Integrations", Step: stepName, Status: StepFailed, Error: err})
		return StepResult{
			Name:   stepName,
			Status: StepFailed,
			Error:  fmt.Errorf("setting up %s: %w", integ.Name, err),
		}
	}

	emit(Event{Phase: "Integrations", Step: stepName, Status: StepDone})
	return StepResult{Name: stepName, Status: StepDone}
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
