package creator

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/progress"
)

const (
	setupPhaseID        progress.PhaseID = "setup"
	dependenciesPhaseID progress.PhaseID = "dependencies"
	integrationsPhaseID progress.PhaseID = "integrations"
)

type preparedDependency struct {
	stepID    progress.StepID
	label     string
	command   string
	parallel  bool
	ecosystem string
}

type preparedCreation struct {
	opts            Options
	plan            progress.Plan
	worktreePath    string
	createStepID    progress.StepID
	mergeStepID     progress.StepID
	envStepID       progress.StepID
	envFiles        []string
	dependencies    []preparedDependency
	integrations    integration.PreparedApply
	hasDependencies bool
	hasIntegrations bool
}

func prepareCreation(runner git.CommandRunner, shell git.ShellRunner, opts Options) (preparedCreation, error) {
	if strings.TrimSpace(opts.BranchName) == "" || strings.TrimSpace(opts.BaseBranch) == "" || strings.TrimSpace(opts.RepoPath) == "" {
		return preparedCreation{}, errors.New("preparing worktree creation: branch, base branch, and repository path are required")
	}
	if err := validateEcosystemIdentities(opts); err != nil {
		return preparedCreation{}, fmt.Errorf("preparing worktree creation: %w", err)
	}
	targets, err := prepareDependencyTargets(runner, opts)
	if err != nil {
		return preparedCreation{}, err
	}

	prepared := preparedCreation{
		opts: opts, worktreePath: git.WorktreePath(opts.RepoPath, opts.BranchName),
		createStepID: semanticStepID("create-worktree", opts.BranchName),
	}
	setup := progress.PlannedPhase{ID: setupPhaseID, Label: "Setup"}
	setup.Steps = append(setup.Steps, progress.PlannedStep{ID: prepared.createStepID, Label: "Create worktree"})
	if opts.MergeBase {
		prepared.mergeStepID = semanticStepID("merge-base", opts.BaseBranch)
		setup.Steps = append(setup.Steps, progress.PlannedStep{ID: prepared.mergeStepID, Label: "Merge base branch"})
	}
	if opts.CopyEnvFiles {
		prepared.envFiles = uniqueEnvFiles(opts)
		if len(prepared.envFiles) > 0 {
			prepared.envStepID = semanticStepID("copy-env", strings.Join(prepared.envFiles, "\x00"))
			setup.Steps = append(setup.Steps, progress.PlannedStep{ID: prepared.envStepID, Label: "Copy env files"})
		}
	}
	prepared.plan.Phases = append(prepared.plan.Phases, setup)

	seenDependencies := map[string]bool{}
	dependencyPhase := progress.PlannedPhase{ID: dependenciesPhaseID, Label: "Dependencies"}
	for _, target := range targets {
		command := strings.TrimSpace(target.ecosystem.Install.Command)
		label := target.ecosystem.Name
		if target.workspace != "" {
			command = strings.TrimSpace(strings.ReplaceAll(target.ecosystem.Install.WorkspaceInstall, "{dir}", target.workspace))
			label = fmt.Sprintf("%s (%s)", target.ecosystem.Name, target.workspace)
		}
		if command == "" {
			continue
		}
		identity := target.ecosystem.Name + "\x00" + target.workspace
		if seenDependencies[identity] {
			return preparedCreation{}, fmt.Errorf("preparing worktree creation: duplicate dependency operation identity %q", label)
		}
		seenDependencies[identity] = true
		operation := preparedDependency{
			stepID: semanticStepID("install-dependencies", identity), label: label, command: command,
			parallel: target.ecosystem.Install.IsParallel(), ecosystem: target.ecosystem.Name,
		}
		prepared.dependencies = append(prepared.dependencies, operation)
		dependencyPhase.Steps = append(dependencyPhase.Steps, progress.PlannedStep{ID: operation.stepID, Label: operation.label})
	}
	if len(dependencyPhase.Steps) > 0 {
		prepared.hasDependencies = true
		prepared.plan.Phases = append(prepared.plan.Phases, dependencyPhase)
	}

	if len(opts.Integrations) > 0 {
		probeDir := opts.SourceWorktree
		if probeDir == "" {
			probeDir = opts.RepoPath
		}
		apply, err := integration.PrepareApplyForTarget(shell, opts.RepoPath, opts.SourceWorktree, probeDir, opts.Integrations, nil, []string{prepared.worktreePath})
		if err != nil {
			return preparedCreation{}, err
		}
		apply, err = apply.BindPhase(integrationsPhaseID, "Integrations")
		if err != nil {
			return preparedCreation{}, err
		}
		if !apply.Empty() {
			prepared.hasIntegrations = true
			prepared.integrations = apply
			prepared.plan.Phases = append(prepared.plan.Phases, apply.Plan().Phases...)
		}
	}
	return prepared, nil
}

func validateEcosystemIdentities(opts Options) error {
	seen := make(map[string]bool, len(opts.Ecosystems))
	for _, ecosystem := range opts.Ecosystems {
		name := strings.TrimSpace(ecosystem.Name)
		if name == "" {
			return errors.New("ecosystem has empty name")
		}
		if seen[name] {
			return fmt.Errorf("duplicate ecosystem identity %q", name)
		}
		seen[name] = true
	}
	return nil
}

func uniqueEnvFiles(opts Options) []string {
	seen := map[string]bool{}
	var files []string
	for _, ecosystem := range opts.Ecosystems {
		for _, name := range ecosystem.EnvFiles {
			name = filepath.Clean(strings.TrimSpace(name))
			if name == "." || filepath.IsAbs(name) || strings.HasPrefix(name, ".."+string(filepath.Separator)) || seen[name] {
				continue
			}
			seen[name] = true
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return files
}

func (p preparedCreation) run(execution *progress.Execution, runner git.CommandRunner, shell git.ShellRunner, result *Result) error {
	createResult, err := execution.Run(setupPhaseID, p.createStepID, func() (string, error) {
		args := []string{"worktree", "add", p.worktreePath}
		if git.BranchExists(runner, p.opts.RepoPath, p.opts.BranchName) {
			args = append(args, p.opts.BranchName)
		} else {
			args = append(args, "-b", p.opts.BranchName, p.opts.BaseBranch)
		}
		if _, err := runner.Run(p.opts.RepoPath, args...); err != nil {
			return "", fmt.Errorf("creating worktree: %w", err)
		}
		return p.worktreePath, nil
	})
	if err != nil {
		return fmt.Errorf("executing worktree creation: %w", err)
	}
	if createResult.Status == progress.StepFailed {
		return p.skipBlocked(execution, "blocked by Create worktree")
	}
	result.WorktreePath = p.worktreePath

	var runErr error
	if p.mergeStepID != "" {
		_, err := execution.Run(setupPhaseID, p.mergeStepID, func() (string, error) {
			_, err := runner.Run(p.worktreePath, "merge", p.opts.BaseBranch, "--no-edit")
			return "", err
		})
		if err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("executing merge: %w", err))
		}
	}
	if p.envStepID != "" && runErr == nil {
		_, err := execution.Run(setupPhaseID, p.envStepID, func() (string, error) {
			return copyPreparedEnvFiles(p.opts.SourceWorktree, p.worktreePath, p.envFiles)
		})
		if err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("executing env copy: %w", err))
		}
	}
	if runErr != nil {
		return runErr
	}
	if p.hasDependencies {
		runErr = errors.Join(runErr, p.runDependencies(execution, shell))
	}
	if p.hasIntegrations {
		runErr = errors.Join(runErr, p.integrations.RunIn(execution, shell))
	}
	return runErr
}

func (p preparedCreation) skipBlocked(execution *progress.Execution, reason string) error {
	var err error
	err = errors.Join(err, execution.SkipPending(setupPhaseID, reason))
	if p.hasDependencies {
		err = errors.Join(err, execution.SkipPending(dependenciesPhaseID, reason))
	}
	if p.hasIntegrations {
		err = errors.Join(err, execution.SkipPending(integrationsPhaseID, reason))
	}
	return err
}

func (p preparedCreation) runDependencies(execution *progress.Execution, shell git.ShellRunner) error {
	groups := make([][]preparedDependency, 0)
	for _, dependency := range p.dependencies {
		if len(groups) == 0 || groups[len(groups)-1][0].ecosystem != dependency.ecosystem {
			groups = append(groups, []preparedDependency{dependency})
		} else {
			groups[len(groups)-1] = append(groups[len(groups)-1], dependency)
		}
	}
	var runErr error
	for _, group := range groups {
		if len(group) == 1 || !group[0].parallel {
			for _, dependency := range group {
				runErr = errors.Join(runErr, runPreparedDependency(execution, shell, p.worktreePath, dependency))
			}
			continue
		}
		sem := make(chan struct{}, maxDepsConcurrency)
		var wg sync.WaitGroup
		var mu sync.Mutex
		for _, dependency := range group {
			dependency := dependency
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				err := runPreparedDependency(execution, shell, p.worktreePath, dependency)
				if err != nil {
					mu.Lock()
					runErr = errors.Join(runErr, err)
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
	}
	return runErr
}

func runPreparedDependency(execution *progress.Execution, shell git.ShellRunner, worktreePath string, dependency preparedDependency) error {
	_, err := execution.Run(dependenciesPhaseID, dependency.stepID, func() (string, error) {
		_, err := shell.RunShell(worktreePath, dependency.command)
		if err != nil {
			return "", fmt.Errorf("installing %s: %w", dependency.label, err)
		}
		return "", nil
	})
	if err != nil {
		return fmt.Errorf("executing dependency %s: %w", dependency.label, err)
	}
	return nil
}

func copyPreparedEnvFiles(source, destination string, files []string) (string, error) {
	var copied []string
	for _, name := range files {
		if _, err := os.Stat(filepath.Join(source, name)); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return "", fmt.Errorf("inspecting %s: %w", name, err)
		}
		if err := fileutil.CopyFile(filepath.Join(source, name), filepath.Join(destination, name)); err != nil {
			return "", fmt.Errorf("copying %s: %w", name, err)
		}
		copied = append(copied, name)
	}
	if len(copied) == 0 {
		return "no source files found", nil
	}
	return strings.Join(copied, ", "), nil
}

func semanticStepID(kind, identity string) progress.StepID {
	sum := sha256.Sum256([]byte(identity))
	return progress.StepID(fmt.Sprintf("%s:%x", kind, sum[:8]))
}
