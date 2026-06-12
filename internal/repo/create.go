package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
)

// GhRunner executes gh CLI commands directly without a shell, preventing shell injection.
type GhRunner interface {
	RunGh(dir string, args ...string) (string, error)
}

// DefaultGhRunner is the production GhRunner that invokes gh via exec.Command.
type DefaultGhRunner struct{}

func (r *DefaultGhRunner) RunGh(dir string, args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

type CreateOptions struct {
	Name          string
	Location      string
	PublishGitHub bool
	Visibility    string // "private" or "public"
	Description   string
}

// PhaseGitHub is the name of the create flow's publish phase. A failure here is
// "soft" (the local repo is fine, just unpublished); a failure in any other
// phase is "hard" (the local repo is broken). Single source for that test.
const PhaseGitHub = "GitHub"

type CreateResult struct {
	RepoPath     string
	WorktreePath string
	GitHubURL    string
	Phases       []progress.Phase
}

// SetupFailed reports whether a non-GitHub phase failed (the local repo itself is
// broken, not merely unpublished) and the first such error.
func (r CreateResult) SetupFailed() (bool, error) {
	for _, p := range r.Phases {
		if p.Name != PhaseGitHub && p.HasFailures() {
			_, step, _ := progress.FirstFailure([]progress.Phase{p})
			return true, step.Error
		}
	}
	return false, nil
}

func Create(runner git.CommandRunner, shell git.ShellRunner, opts CreateOptions, emit func(progress.Event)) CreateResult {
	return CreateWithGh(runner, shell, &DefaultGhRunner{}, opts, emit)
}

func CreateWithGh(runner git.CommandRunner, _ git.ShellRunner, gh GhRunner, opts CreateOptions, emit func(progress.Event)) CreateResult {
	result := CreateResult{}
	repoPath := filepath.Join(opts.Location, opts.Name)
	result.RepoPath = repoPath

	setupPhase := runCreateSetup(runner, repoPath, opts, emit)
	result.Phases = append(result.Phases, setupPhase)
	if setupPhase.HasFailures() {
		return result
	}
	result.WorktreePath = git.WorktreePath(repoPath, "main")

	if opts.PublishGitHub {
		ghPhase := runCreateGitHub(runner, gh, repoPath, opts, emit)
		result.Phases = append(result.Phases, ghPhase)
		if !ghPhase.HasFailures() {
			// Extract GitHub URL from user lookup
			for _, step := range ghPhase.Steps {
				if step.Name == "Look up GitHub user" && step.Status == progress.StepDone {
					result.GitHubURL = fmt.Sprintf("github.com/%s/%s", step.Message, opts.Name)
				}
			}
		}
	}

	return result
}

func runCreateSetup(runner git.CommandRunner, repoPath string, opts CreateOptions, emit func(progress.Event)) progress.Phase {
	rec := progress.NewPhaseRecorder("Setup", emit)
	barePath := filepath.Join(repoPath, ".bare")

	ok := rec.Step("Create directory", func() (string, error) {
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return "", err
		}
		// Abort if the directory already had content.
		if entries, _ := os.ReadDir(repoPath); len(entries) > 0 {
			return "", fmt.Errorf("directory already exists and is not empty: %s", repoPath)
		}
		return "", nil
	})
	if !ok {
		return rec.Phase()
	}

	ok = rec.Step("Init bare repository", func() (string, error) {
		if err := os.MkdirAll(barePath, 0755); err != nil {
			return "", err
		}
		_, err := runner.Run(barePath, "init", "--bare")
		return "", err
	})
	if !ok {
		return rec.Phase()
	}

	ok = rec.Step("Create .git pointer", func() (string, error) {
		return "", os.WriteFile(filepath.Join(repoPath, ".git"), []byte("gitdir: .bare\n"), 0644)
	})
	if !ok {
		return rec.Phase()
	}

	ok = rec.Step("Configure refspec", func() (string, error) {
		_, err := runner.Run(barePath, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
		return "", err
	})
	if !ok {
		return rec.Phase()
	}

	ok = rec.Step("Create main worktree", func() (string, error) {
		_, err := runner.Run(repoPath, "worktree", "add", git.WorktreePath(repoPath, "main"), "-b", "main")
		return "", err
	})
	if !ok {
		return rec.Phase()
	}

	rec.Step("Initial commit", func() (string, error) {
		mainPath := git.WorktreePath(repoPath, "main")
		if err := os.MkdirAll(mainPath, 0755); err != nil {
			return "", err
		}
		readmePath := filepath.Join(mainPath, "README.md")
		if err := os.WriteFile(readmePath, fmt.Appendf(nil, "# %s\n", opts.Name), 0644); err != nil {
			return "", err
		}
		if _, err := runner.Run(mainPath, "add", "-A"); err != nil {
			return "", err
		}
		_, err := runner.Run(mainPath, "commit", "-m", "Initial commit")
		return "", err
	})

	return rec.Phase()
}

func runCreateGitHub(runner git.CommandRunner, gh GhRunner, repoPath string, opts CreateOptions, emit func(progress.Event)) progress.Phase {
	rec := progress.NewPhaseRecorder(PhaseGitHub, emit)
	barePath := filepath.Join(repoPath, ".bare")

	var ghUser string
	ok := rec.Step("Look up GitHub user", func() (string, error) {
		var err error
		ghUser, err = gh.RunGh(repoPath, "api", "user", "--jq", ".login")
		return ghUser, err
	})
	if !ok {
		return rec.Phase()
	}

	// Create the repo without --source/--push — we push manually after
	// configuring the remote.
	ok = rec.Step("Create GitHub repository", func() (string, error) {
		ghArgs := []string{"repo", "create", opts.Name, "--" + opts.Visibility}
		if opts.Description != "" {
			ghArgs = append(ghArgs, "--description", opts.Description)
		}
		_, err := gh.RunGh(repoPath, ghArgs...)
		return "", err
	})
	if !ok {
		return rec.Phase()
	}

	// Configure the remote using gh's configured protocol so the push uses the
	// auth the user actually has. Forcing SSH breaks push for an HTTPS-only gh
	// login (gh's default), orphaning the just-created empty GitHub repo.
	ok = rec.Step("Configure remote", func() (string, error) {
		_, err := runner.Run(barePath, "remote", "set-url", "origin", ghRemoteURL(gh, repoPath, ghUser, opts.Name))
		return "", err
	})
	if !ok {
		return rec.Phase()
	}

	ok = rec.Step("Push to GitHub", func() (string, error) {
		if _, err := runner.Run(git.WorktreePath(repoPath, "main"), "push", "-u", "origin", "main"); err != nil {
			// The empty remote repo from "Create GitHub repository" is left behind;
			// tell the user so they can delete it or push manually before retrying.
			return "", fmt.Errorf("%w (an empty GitHub repo %q now exists; delete it or push to it manually before retrying)", err, opts.Name)
		}
		return "", nil
	})
	if !ok {
		return rec.Phase()
	}

	rec.Step("Set remote HEAD", func() (string, error) {
		_, err := runner.Run(barePath, "remote", "set-head", "origin", "main")
		return "", err
	})

	return rec.Phase()
}

// ghRemoteURL returns the origin URL matching gh's configured git protocol, so
// the push uses the auth the user actually has. gh defaults to HTTPS; only an
// explicit "ssh" protocol yields an SSH URL.
func ghRemoteURL(gh GhRunner, repoPath, user, name string) string {
	if proto, err := gh.RunGh(repoPath, "config", "get", "git_protocol"); err == nil && strings.TrimSpace(proto) == "ssh" {
		return fmt.Sprintf("git@github.com:%s/%s.git", user, name)
	}
	return fmt.Sprintf("https://github.com/%s/%s.git", user, name)
}
