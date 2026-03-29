package creator

import (
	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
)

type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepDone
	StepFailed
	StepSkipped
)

type StepResult struct {
	Name    string
	Status  StepStatus
	Message string
	Error   error
}

type Phase struct {
	Name  string
	Steps []StepResult
}

type Event struct {
	Phase   string
	Step    string
	Status  StepStatus
	Message string
	Error   error
}

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
	Phases       []Phase
}

func (r *Result) HasFailures() bool {
	for _, p := range r.Phases {
		for _, s := range p.Steps {
			if s.Status == StepFailed {
				return true
			}
		}
	}
	return false
}

func Run(runner git.CommandRunner, opts Options, emit func(Event)) Result {
	result := Result{}

	setupPhase := runSetup(runner, opts, emit)
	result.Phases = append(result.Phases, setupPhase)

	if setupPhase.Steps[0].Status == StepFailed {
		return result
	}
	result.WorktreePath = worktreePath(opts.RepoPath, opts.BranchName)

	depsPhase := runDeps(runner, result.WorktreePath, opts, emit)
	result.Phases = append(result.Phases, depsPhase)

	intPhase := runIntegrations(runner, result.WorktreePath, opts, emit)
	result.Phases = append(result.Phases, intPhase)

	return result
}

func worktreePath(repoPath, branch string) string {
	return repoPath + "/" + SanitizeBranchPath(branch)
}
