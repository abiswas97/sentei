package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
)

type CloneOptions struct {
	URL      string
	Location string
	Name     string
}

type CloneResult struct {
	RepoPath      string
	WorktreePath  string
	DefaultBranch string
	OriginURL     string
	Phases        []Phase
}

// DeriveRepoName extracts a repository name from a git URL.
// "git@github.com:user/repo.git" → "repo"
// "https://github.com/user/repo.git" → "repo"
// "https://github.com/user/repo" → "repo"
func DeriveRepoName(url string) string {
	// Handle SSH-style URLs: git@host:path
	if idx := strings.LastIndex(url, ":"); idx != -1 && !strings.Contains(url, "://") {
		url = url[idx+1:]
	}

	// Take last path segment
	name := url
	if idx := strings.LastIndex(name, "/"); idx != -1 {
		name = name[idx+1:]
	}

	// Strip .git suffix
	name = strings.TrimSuffix(name, ".git")

	return name
}

func Clone(runner git.CommandRunner, opts CloneOptions, emit func(Event)) CloneResult {
	result := CloneResult{OriginURL: opts.URL}
	repoPath := filepath.Join(opts.Location, opts.Name)
	result.RepoPath = repoPath
	barePath := filepath.Join(repoPath, ".bare")

	// Phase 1: Clone
	clonePhase := runClonePhase(runner, opts.Location, opts.URL, barePath, emit)
	result.Phases = append(result.Phases, clonePhase)
	if clonePhase.HasFailures() {
		return result
	}

	// Phase 2: Structure
	structPhase := runCloneStructure(runner, repoPath, barePath, emit)
	result.Phases = append(result.Phases, structPhase)
	if structPhase.HasFailures() {
		return result
	}

	// Phase 3: Worktree
	wtPhase, branch := runCloneWorktree(runner, repoPath, barePath, emit)
	result.Phases = append(result.Phases, wtPhase)
	result.DefaultBranch = branch
	result.WorktreePath = filepath.Join(repoPath, branch)

	return result
}

func runClonePhase(runner git.CommandRunner, location, url, barePath string, emit func(Event)) Phase {
	phase := Phase{Name: "Clone"}
	phaseName := "Clone"

	emit(Event{Phase: phaseName, Step: "Clone bare repository", Status: StepRunning})
	_, err := runner.Run(location, "clone", "--bare", url, barePath)
	if err != nil {
		step := StepResult{Name: "Clone bare repository", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Clone bare repository", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Clone bare repository", Status: StepDone})

	return phase
}

func runCloneStructure(runner git.CommandRunner, repoPath, barePath string, emit func(Event)) Phase {
	phase := Phase{Name: "Structure"}
	phaseName := "Structure"

	// Create .git pointer (ensure repoPath exists — bare clone creates .bare but not necessarily the parent)
	emit(Event{Phase: phaseName, Step: "Create .git pointer", Status: StepRunning})
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		step := StepResult{Name: "Create .git pointer", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	gitPointerPath := filepath.Join(repoPath, ".git")
	if err := os.WriteFile(gitPointerPath, []byte("gitdir: .bare\n"), 0644); err != nil {
		step := StepResult{Name: "Create .git pointer", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create .git pointer", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create .git pointer", Status: StepDone})

	// Configure refspec
	emit(Event{Phase: phaseName, Step: "Configure refspec", Status: StepRunning})
	_, err := runner.Run(barePath, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	if err != nil {
		step := StepResult{Name: "Configure refspec", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Configure refspec", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Configure refspec", Status: StepDone})

	return phase
}

func runCloneWorktree(runner git.CommandRunner, repoPath, barePath string, emit func(Event)) (Phase, string) {
	phase := Phase{Name: "Worktree"}
	phaseName := "Worktree"

	// Detect default branch
	emit(Event{Phase: phaseName, Step: "Detect default branch", Status: StepRunning})
	branch := detectDefaultBranch(runner, barePath)
	phase.Steps = append(phase.Steps, StepResult{
		Name: "Detect default branch", Status: StepDone, Message: branch,
	})
	emit(Event{Phase: phaseName, Step: "Detect default branch", Status: StepDone, Message: branch})

	// Create worktree
	emit(Event{Phase: phaseName, Step: "Create worktree", Status: StepRunning})
	_, err := runner.Run(repoPath, "worktree", "add", branch, branch)
	if err != nil {
		step := StepResult{Name: "Create worktree", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase, branch
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create worktree", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create worktree", Status: StepDone})

	// Set upstream
	emit(Event{Phase: phaseName, Step: "Set upstream tracking", Status: StepRunning})
	wtPath := filepath.Join(repoPath, branch)
	_, err = runner.Run(wtPath, "branch", fmt.Sprintf("--set-upstream-to=origin/%s", branch))
	if err != nil {
		step := StepResult{Name: "Set upstream tracking", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase, branch
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Set upstream tracking", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Set upstream tracking", Status: StepDone})

	return phase, branch
}

func detectDefaultBranch(runner git.CommandRunner, barePath string) string {
	output, err := runner.Run(barePath, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		// "refs/remotes/origin/main" → "main"
		branch := strings.TrimPrefix(output, "refs/remotes/origin/")
		if branch != output && branch != "" {
			return branch
		}
	}

	// Fallback: try main, then master
	for _, candidate := range []string{"main", "master"} {
		_, err := runner.Run(barePath, "show-ref", "--verify", fmt.Sprintf("refs/heads/%s", candidate))
		if err == nil {
			return candidate
		}
	}

	return "main" // last resort
}
