package creator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/ecosystem"
	"github.com/abiswas97/sentei/internal/git"
)

const maxDepsConcurrency = 5

func runDeps(runner git.CommandRunner, wtPath string, opts Options, emit func(Event)) Phase {
	phase := Phase{Name: "Dependencies"}

	if len(opts.Ecosystems) == 0 {
		return phase
	}

	for _, eco := range opts.Ecosystems {
		steps := installEcosystem(runner, wtPath, eco, emit)
		phase.Steps = append(phase.Steps, steps...)
	}

	return phase
}

func installEcosystem(runner git.CommandRunner, wtPath string, eco config.EcosystemConfig, emit func(Event)) []StepResult {
	rootStep := runInstallCommand(runner, wtPath, eco.Name, eco.Install.Command, emit)
	steps := []StepResult{rootStep}

	if rootStep.Status == StepFailed {
		return steps
	}

	if eco.Install.WorkspaceDetect == "" || eco.Install.WorkspaceInstall == "" {
		return steps
	}

	workspaces, err := ecosystem.DetectWorkspaces(wtPath, eco.Install.WorkspaceDetect)
	if err != nil || len(workspaces) == 0 {
		return steps
	}

	if eco.Install.IsParallel() {
		wsSteps := installWorkspacesParallel(runner, wtPath, eco, workspaces, emit)
		steps = append(steps, wsSteps...)
	} else {
		for _, ws := range workspaces {
			cmd := strings.ReplaceAll(eco.Install.WorkspaceInstall, "{dir}", ws)
			step := runInstallCommand(runner, wtPath, fmt.Sprintf("%s (%s)", eco.Name, ws), cmd, emit)
			steps = append(steps, step)
		}
	}

	return steps
}

func installWorkspacesParallel(runner git.CommandRunner, wtPath string, eco config.EcosystemConfig, workspaces []string, emit func(Event)) []StepResult {
	results := make([]StepResult, len(workspaces))
	sem := make(chan struct{}, maxDepsConcurrency)
	var wg sync.WaitGroup

	for i, ws := range workspaces {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, workspace string) {
			defer wg.Done()
			defer func() { <-sem }()

			cmd := strings.ReplaceAll(eco.Install.WorkspaceInstall, "{dir}", workspace)
			stepName := fmt.Sprintf("%s (%s)", eco.Name, workspace)
			results[idx] = runInstallCommand(runner, wtPath, stepName, cmd, emit)
		}(i, ws)
	}

	wg.Wait()
	return results
}

func runInstallCommand(runner git.CommandRunner, wtPath, stepName, command string, emit func(Event)) StepResult {
	emit(Event{Phase: "Dependencies", Step: stepName, Status: StepRunning})

	args := strings.Fields(command)
	if len(args) == 0 {
		emit(Event{Phase: "Dependencies", Step: stepName, Status: StepFailed, Error: fmt.Errorf("empty install command")})
		return StepResult{
			Name:   stepName,
			Status: StepFailed,
			Error:  fmt.Errorf("empty install command for %s", stepName),
		}
	}

	_, err := runner.Run(wtPath, args...)
	if err != nil {
		emit(Event{Phase: "Dependencies", Step: stepName, Status: StepFailed, Error: err})
		return StepResult{
			Name:   stepName,
			Status: StepFailed,
			Error:  fmt.Errorf("installing %s: %w", stepName, err),
		}
	}

	emit(Event{Phase: "Dependencies", Step: stepName, Status: StepDone})
	return StepResult{Name: stepName, Status: StepDone}
}
