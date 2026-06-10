package creator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/ecosystem"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/pipeline"
)

const maxDepsConcurrency = 5

func runDeps(shell git.ShellRunner, wtPath string, opts Options, emit func(pipeline.Event)) pipeline.Phase {
	phase := pipeline.Phase{Name: "Dependencies"}

	if len(opts.Ecosystems) == 0 {
		return phase
	}

	for _, eco := range opts.Ecosystems {
		steps := installEcosystem(shell, wtPath, eco, emit)
		phase.Steps = append(phase.Steps, steps...)
	}

	return phase
}

func installEcosystem(shell git.ShellRunner, wtPath string, eco config.EcosystemConfig, emit func(pipeline.Event)) []pipeline.StepResult {
	if eco.Install.WorkspaceDetect == "" || eco.Install.WorkspaceInstall == "" {
		rootStep := runInstallCommand(shell, wtPath, eco.Name, eco.Install.Command, emit)
		return []pipeline.StepResult{rootStep}
	}

	workspaces, err := ecosystem.DetectWorkspaces(wtPath, eco.Install.WorkspaceDetect)
	if err != nil || len(workspaces) == 0 {
		rootStep := runInstallCommand(shell, wtPath, eco.Name, eco.Install.Command, emit)
		return []pipeline.StepResult{rootStep}
	}

	var steps []pipeline.StepResult

	if eco.Install.IsParallel() {
		wsSteps := installWorkspacesParallel(shell, wtPath, eco, workspaces, emit)
		steps = append(steps, wsSteps...)
	} else {
		for _, ws := range workspaces {
			cmd := strings.ReplaceAll(eco.Install.WorkspaceInstall, "{dir}", ws)
			step := runInstallCommand(shell, wtPath, fmt.Sprintf("%s (%s)", eco.Name, ws), cmd, emit)
			steps = append(steps, step)
		}
	}

	return steps
}

func installWorkspacesParallel(shell git.ShellRunner, wtPath string, eco config.EcosystemConfig, workspaces []string, emit func(pipeline.Event)) []pipeline.StepResult {
	results := make([]pipeline.StepResult, len(workspaces))
	sem := make(chan struct{}, maxDepsConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	safeEmit := func(e pipeline.Event) {
		mu.Lock()
		defer mu.Unlock()
		emit(e)
	}

	for i, ws := range workspaces {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, workspace string) {
			defer wg.Done()
			defer func() { <-sem }()

			cmd := strings.ReplaceAll(eco.Install.WorkspaceInstall, "{dir}", workspace)
			stepName := fmt.Sprintf("%s (%s)", eco.Name, workspace)
			results[idx] = runInstallCommand(shell, wtPath, stepName, cmd, safeEmit)
		}(i, ws)
	}

	wg.Wait()
	return results
}

func runInstallCommand(shell git.ShellRunner, wtPath, stepName, command string, emit func(pipeline.Event)) pipeline.StepResult {
	emit(pipeline.Event{Phase: "Dependencies", Step: stepName, Status: pipeline.StepRunning})

	if command == "" {
		emit(pipeline.Event{Phase: "Dependencies", Step: stepName, Status: pipeline.StepFailed, Error: fmt.Errorf("empty install command")})
		return pipeline.StepResult{
			Name:   stepName,
			Status: pipeline.StepFailed,
			Error:  fmt.Errorf("empty install command for %s", stepName),
		}
	}

	_, err := shell.RunShell(wtPath, command)
	if err != nil {
		emit(pipeline.Event{Phase: "Dependencies", Step: stepName, Status: pipeline.StepFailed, Error: err})
		return pipeline.StepResult{
			Name:   stepName,
			Status: pipeline.StepFailed,
			Error:  fmt.Errorf("installing %s: %w", stepName, err),
		}
	}

	emit(pipeline.Event{Phase: "Dependencies", Step: stepName, Status: pipeline.StepDone})
	return pipeline.StepResult{Name: stepName, Status: pipeline.StepDone}
}
