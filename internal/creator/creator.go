package creator

import (
	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/pipeline"
)

type Options struct {
	BranchName     string
	BaseBranch     string
	RepoPath       string
	SourceWorktree string
	MergeBase      bool
	CopyEnvFiles   bool
	Ecosystems     []config.EcosystemConfig
	Integrations   []integration.Integration
}

type Result struct {
	WorktreePath string
	Phases       []pipeline.Phase
}

func (r *Result) HasFailures() bool {
	return pipeline.PhasesHaveFailures(r.Phases)
}

func Run(runner git.CommandRunner, shell git.ShellRunner, opts Options, emit func(pipeline.Event)) Result {
	result := Result{}

	setupPhase := runSetup(runner, opts, emit)
	result.Phases = append(result.Phases, setupPhase)

	if setupPhase.Steps[0].Status == pipeline.StepFailed {
		return result
	}
	result.WorktreePath = git.WorktreePath(opts.RepoPath, opts.BranchName)

	depsPhase := runDeps(shell, result.WorktreePath, opts, emit)
	result.Phases = append(result.Phases, depsPhase)

	intPhase := runIntegrations(shell, result.WorktreePath, opts, emit)
	result.Phases = append(result.Phases, intPhase)

	return result
}
