package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/pipeline"
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
	Phases        []pipeline.Phase
}

// DeriveRepoName extracts a repository name from a git URL.
// "git@github.com:user/repo.git" → "repo"
// "https://github.com/user/repo.git" → "repo"
// "https://github.com/user/repo" → "repo"
func DeriveRepoName(url string) string {
	url = strings.TrimSpace(url)

	// Drop query string / fragment so "repo?ref=main" / "repo#frag" don't leak in.
	if idx := strings.IndexAny(url, "?#"); idx != -1 {
		url = url[:idx]
	}
	// Drop trailing slashes so "host/user/repo/" yields "repo", not "".
	url = strings.TrimRight(url, "/")

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

func Clone(runner git.CommandRunner, opts CloneOptions, emit func(pipeline.Event)) CloneResult {
	result := CloneResult{OriginURL: opts.URL}
	repoPath := filepath.Join(opts.Location, opts.Name)
	result.RepoPath = repoPath
	barePath := filepath.Join(repoPath, ".bare")

	// Validate the target before touching the filesystem. An empty or path-like
	// name would otherwise turn the current directory (or an existing repo) into
	// a bare repo with a "success" message.
	if vphase, ok := validateCloneTarget(opts, repoPath); !ok {
		result.Phases = append(result.Phases, vphase)
		return result
	}

	// repoPath did not exist (validated above), so on a failure that leaves no
	// usable checkout we can remove exactly what we created and leave nothing
	// half-built behind.
	rollback := func() { _ = fileutil.RemoveAllRetry(repoPath) }

	// Phase 1: Clone
	clonePhase := runClonePhase(runner, opts.Location, opts.URL, barePath, emit)
	result.Phases = append(result.Phases, clonePhase)
	if clonePhase.HasFailures() {
		rollback()
		return result
	}

	// Phase 2: Structure
	structPhase := runCloneStructure(runner, repoPath, barePath, emit)
	result.Phases = append(result.Phases, structPhase)
	if structPhase.HasFailures() {
		rollback()
		return result
	}

	// Phase 3: Worktree
	wtPhase, branch, worktreeCreated := runCloneWorktree(runner, repoPath, barePath, emit)
	result.Phases = append(result.Phases, wtPhase)
	result.DefaultBranch = branch
	// Only advertise a worktree path when one was actually created. A failed
	// worktree add must not leave WorktreePath set, or consumers report success
	// and point the user at a directory that does not exist.
	if worktreeCreated {
		result.WorktreePath = git.WorktreePath(repoPath, branch)
	}
	// Roll back only when no usable checkout exists. If the worktree was created
	// and merely upstream tracking failed, the repo is usable: keep it.
	if wtPhase.HasFailures() && !worktreeCreated {
		rollback()
	}

	return result
}

// validateCloneTarget rejects inputs that would corrupt an unintended directory.
// The returned phase is only meaningful (and only surfaced) when ok is false.
func validateCloneTarget(opts CloneOptions, repoPath string) (pipeline.Phase, bool) {
	phase := pipeline.Phase{Name: "Validate"}
	fail := func(err error) (pipeline.Phase, bool) {
		phase.Steps = append(phase.Steps, pipeline.StepResult{Name: "Validate target", Status: pipeline.StepFailed, Error: err})
		return phase, false
	}

	switch {
	case opts.Name == "":
		return fail(errors.New("could not derive a repository name from the URL; pass --name"))
	case opts.Name == "." || opts.Name == ".." || strings.ContainsAny(opts.Name, `/\`):
		return fail(fmt.Errorf("invalid repository name %q: must be a directory name, not a path", opts.Name))
	}
	if _, err := os.Stat(repoPath); err == nil {
		return fail(fmt.Errorf("target already exists: %s", repoPath))
	}

	return phase, true
}

func runClonePhase(runner git.CommandRunner, location, url, barePath string, emit func(pipeline.Event)) pipeline.Phase {
	rec := pipeline.NewPhaseRecorder("Clone", emit)
	rec.Step("Clone bare repository", func() (string, error) {
		_, err := runner.Run(location, "clone", "--bare", url, barePath)
		return "", err
	})
	return rec.Phase()
}

func runCloneStructure(runner git.CommandRunner, repoPath, barePath string, emit func(pipeline.Event)) pipeline.Phase {
	rec := pipeline.NewPhaseRecorder("Structure", emit)

	// Ensure repoPath exists — bare clone creates .bare but not necessarily the parent.
	ok := rec.Step("Create .git pointer", func() (string, error) {
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return "", err
		}
		return "", os.WriteFile(filepath.Join(repoPath, ".git"), []byte("gitdir: .bare\n"), 0644)
	})
	if !ok {
		return rec.Phase()
	}

	rec.Step("Configure refspec", func() (string, error) {
		_, err := runner.Run(barePath, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
		return "", err
	})
	return rec.Phase()
}

func runCloneWorktree(runner git.CommandRunner, repoPath, barePath string, emit func(pipeline.Event)) (pipeline.Phase, string, bool) {
	rec := pipeline.NewPhaseRecorder("Worktree", emit)

	var branch string
	rec.Step("Detect default branch", func() (string, error) {
		branch = git.DetectDefaultBranch(runner, barePath)
		return branch, nil
	})

	// An empty remote leaves HEAD pointing at a branch with no commit; worktree
	// add would otherwise fail with a cryptic "invalid reference". Surface it.
	if !git.BranchExists(runner, barePath, branch) {
		rec.Fail("Create worktree", fmt.Errorf("remote has no commits on %q yet (nothing to check out)", branch))
		return rec.Phase(), branch, false
	}

	// Create worktree. The branch is passed explicitly as the commit-ish:
	// without it, git derives a NEW branch from the path's basename instead of
	// checking out the existing one.
	wtPath := git.WorktreePath(repoPath, branch)
	ok := rec.Step("Create worktree", func() (string, error) {
		_, err := runner.Run(repoPath, "worktree", "add", wtPath, branch)
		return "", err
	})
	if !ok {
		return rec.Phase(), branch, false
	}

	// Tracking is best-effort: the checkout above is already usable. Populating
	// refs/remotes/origin/* (fetch) and setting upstream both need the remote; a
	// network/auth failure here must NOT fail the clone. pipeline.StepSkipped keeps
	// HasFailures() false so the clone still reports success, just without tracking.
	rec.Emit("Set upstream tracking", pipeline.StepRunning, "")
	trackErr := func() error {
		if _, err := runner.Run(barePath, "fetch", "origin"); err != nil {
			return err
		}
		_, err := runner.Run(wtPath, "branch", fmt.Sprintf("--set-upstream-to=origin/%s", branch))
		return err
	}()
	if trackErr != nil {
		rec.Skip("Set upstream tracking", "no tracking: "+trackErr.Error())
		return rec.Phase(), branch, true
	}
	rec.Done("Set upstream tracking", "")

	return rec.Phase(), branch, true
}
