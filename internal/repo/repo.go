package repo

import (
	"path/filepath"

	"github.com/abiswas97/sentei/internal/git"
)

type RepoContext int

const (
	ContextBareRepo    RepoContext = iota // bare repo with worktrees — full menu
	ContextNonBareRepo                    // regular git repo — offer migrate
	ContextNoRepo                         // not in a git repo — offer create/clone
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

// DetectContext determines the repo context at the given path.
//
// Detection logic:
//  1. git rev-parse --is-bare-repository → "true" means ContextBareRepo
//  2. Check for .bare directory at repo root (sentei's bare structure from a worktree)
//  3. git rev-parse --git-dir succeeds → ContextNonBareRepo
//  4. Otherwise → ContextNoRepo
func DetectContext(runner git.CommandRunner, path string) RepoContext {
	output, err := runner.Run(path, "rev-parse", "--is-bare-repository")
	if err == nil && output == "true" {
		return ContextBareRepo
	}

	// Check if git repo at all
	_, err = runner.Run(path, "rev-parse", "--git-dir")
	if err != nil {
		return ContextNoRepo
	}

	// Inside a git repo — check for sentei's .bare directory via --git-common-dir
	// (--show-toplevel returns the worktree root, not the bare repo root)
	commonDir, err := runner.Run(path, "rev-parse", "--git-common-dir")
	if err == nil {
		if !filepath.IsAbs(commonDir) {
			commonDir = filepath.Join(path, commonDir)
		}
		// commonDir is .bare itself (sentei's structure)
		if filepath.Base(commonDir) == ".bare" {
			return ContextBareRepo
		}
	}

	return ContextNonBareRepo
}

// ResolveBareRoot resolves the bare repo root from any path (worktree, bare root, or inside .bare).
// Falls back to path if git commands fail or the structure isn't recognized.
func ResolveBareRoot(runner git.CommandRunner, path string) string {
	commonDir, err := runner.Run(path, "rev-parse", "--git-common-dir")
	if err != nil {
		return path
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(path, commonDir)
	}
	// If commonDir is .bare, parent is the bare repo root
	if filepath.Base(commonDir) == ".bare" {
		return filepath.Dir(commonDir)
	}
	// Otherwise return parent of commonDir (e.g., parent of .git)
	return filepath.Dir(commonDir)
}

func (r *Phase) HasFailures() bool {
	for _, s := range r.Steps {
		if s.Status == StepFailed {
			return true
		}
	}
	return false
}
