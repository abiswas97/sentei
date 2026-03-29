package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
)

func runIntegrations(runner git.CommandRunner, wtPath string, opts Options, emit func(Event)) Phase {
	phase := Phase{Name: "Integrations"}

	if len(opts.Integrations) == 0 {
		return phase
	}

	for _, integ := range opts.Integrations {
		steps := setupIntegration(runner, wtPath, opts.RepoPath, integ, emit)
		phase.Steps = append(phase.Steps, steps...)
	}

	return phase
}

func setupIntegration(runner git.CommandRunner, wtPath, repoPath string, integ integration.Integration, emit func(Event)) []StepResult {
	var steps []StepResult

	installed := detectIntegration(runner, wtPath, integ)

	if !installed {
		depSteps := checkAndInstallDeps(runner, wtPath, integ, emit)
		steps = append(steps, depSteps...)

		for _, s := range depSteps {
			if s.Status == StepFailed {
				return steps
			}
		}

		installStep := installIntegration(runner, wtPath, integ, emit)
		steps = append(steps, installStep)
		if installStep.Status == StepFailed {
			return steps
		}
	}

	setupStep := runSetupCommand(runner, wtPath, repoPath, integ, emit)
	steps = append(steps, setupStep)

	if setupStep.Status != StepFailed && len(integ.GitignoreEntries) > 0 {
		appendGitignore(wtPath, integ.GitignoreEntries)
	}

	return steps
}

func detectIntegration(runner git.CommandRunner, wtPath string, integ integration.Integration) bool {
	if integ.Detect.Command != "" {
		args := strings.Fields(integ.Detect.Command)
		_, err := runner.Run(wtPath, args...)
		return err == nil
	}
	if integ.Detect.BinaryName != "" {
		_, err := runner.Run(wtPath, integ.Detect.BinaryName, "--version")
		return err == nil
	}
	return false
}

func checkAndInstallDeps(runner git.CommandRunner, wtPath string, integ integration.Integration, emit func(Event)) []StepResult {
	var steps []StepResult

	for _, dep := range integ.Dependencies {
		stepName := fmt.Sprintf("Check %s", dep.Name)
		emit(Event{Phase: "Integrations", Step: stepName, Status: StepRunning})

		args := strings.Fields(dep.Detect)
		_, err := runner.Run(wtPath, args...)
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
		installArgs := strings.Fields(dep.Install)
		_, installErr := runner.Run(wtPath, installArgs...)
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

func installIntegration(runner git.CommandRunner, wtPath string, integ integration.Integration, emit func(Event)) StepResult {
	stepName := fmt.Sprintf("Install %s", integ.Name)
	emit(Event{Phase: "Integrations", Step: stepName, Status: StepRunning})

	args := strings.Fields(integ.Install.Command)
	_, err := runner.Run(wtPath, args...)
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

func runSetupCommand(runner git.CommandRunner, wtPath, repoPath string, integ integration.Integration, emit func(Event)) StepResult {
	stepName := fmt.Sprintf("Setup %s", integ.Name)

	if integ.Setup.Command == "" {
		return StepResult{Name: stepName, Status: StepSkipped}
	}

	emit(Event{Phase: "Integrations", Step: stepName, Status: StepRunning})

	command := strings.ReplaceAll(integ.Setup.Command, "{path}", wtPath)
	args := strings.Fields(command)

	var runDir string
	switch integ.Setup.WorkingDir {
	case "repo":
		runDir = repoPath
	default:
		runDir = wtPath
	}

	_, err := runner.Run(runDir, args...)
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

func appendGitignore(dir string, entries []string) {
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
		return
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	for _, entry := range toAdd {
		_, _ = fmt.Fprintln(f, entry)
	}
}
