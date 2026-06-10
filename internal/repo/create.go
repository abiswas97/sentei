package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/git"
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
	Phases       []Phase
}

// SetupFailed reports whether a non-GitHub phase failed (the local repo itself is
// broken, not merely unpublished) and the first such error.
func (r CreateResult) SetupFailed() (bool, error) {
	for _, p := range r.Phases {
		if p.Name != PhaseGitHub && p.HasFailures() {
			_, step, _ := FirstFailure([]Phase{p})
			return true, step.Error
		}
	}
	return false, nil
}

func Create(runner git.CommandRunner, shell git.ShellRunner, opts CreateOptions, emit func(Event)) CreateResult {
	return CreateWithGh(runner, shell, &DefaultGhRunner{}, opts, emit)
}

func CreateWithGh(runner git.CommandRunner, _ git.ShellRunner, gh GhRunner, opts CreateOptions, emit func(Event)) CreateResult {
	result := CreateResult{}
	repoPath := filepath.Join(opts.Location, opts.Name)
	result.RepoPath = repoPath

	setupPhase := runCreateSetup(runner, repoPath, opts, emit)
	result.Phases = append(result.Phases, setupPhase)
	if setupPhase.HasFailures() {
		return result
	}
	result.WorktreePath = filepath.Join(repoPath, "main")

	if opts.PublishGitHub {
		ghPhase := runCreateGitHub(runner, gh, repoPath, opts, emit)
		result.Phases = append(result.Phases, ghPhase)
		if !ghPhase.HasFailures() {
			// Extract GitHub URL from user lookup
			for _, step := range ghPhase.Steps {
				if step.Name == "Look up GitHub user" && step.Status == StepDone {
					result.GitHubURL = fmt.Sprintf("github.com/%s/%s", step.Message, opts.Name)
				}
			}
		}
	}

	return result
}

func runCreateSetup(runner git.CommandRunner, repoPath string, opts CreateOptions, emit func(Event)) Phase {
	phase := Phase{Name: "Setup"}
	phaseName := "Setup"

	// Create directory
	emit(Event{Phase: phaseName, Step: "Create directory", Status: StepRunning})
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		step := StepResult{Name: "Create directory", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}

	// Check directory is empty (abort if it already had content)
	entries, _ := os.ReadDir(repoPath)
	if len(entries) > 0 {
		err := fmt.Errorf("directory already exists and is not empty: %s", repoPath)
		step := StepResult{Name: "Create directory", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create directory", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create directory", Status: StepDone})

	// Init bare repo
	emit(Event{Phase: phaseName, Step: "Init bare repository", Status: StepRunning})
	barePath := filepath.Join(repoPath, ".bare")
	if err := os.MkdirAll(barePath, 0755); err != nil {
		step := StepResult{Name: "Init bare repository", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	_, err := runner.Run(barePath, "init", "--bare")
	if err != nil {
		step := StepResult{Name: "Init bare repository", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Init bare repository", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Init bare repository", Status: StepDone})

	// Create .git pointer file
	emit(Event{Phase: phaseName, Step: "Create .git pointer", Status: StepRunning})
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
	_, err = runner.Run(barePath, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	if err != nil {
		step := StepResult{Name: "Configure refspec", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Configure refspec", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Configure refspec", Status: StepDone})

	// Create main worktree
	emit(Event{Phase: phaseName, Step: "Create main worktree", Status: StepRunning})
	_, err = runner.Run(repoPath, "worktree", "add", "main", "-b", "main")
	if err != nil {
		step := StepResult{Name: "Create main worktree", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create main worktree", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create main worktree", Status: StepDone})

	// Create README and initial commit
	emit(Event{Phase: phaseName, Step: "Initial commit", Status: StepRunning})
	mainPath := filepath.Join(repoPath, "main")
	if err := os.MkdirAll(mainPath, 0755); err != nil {
		step := StepResult{Name: "Initial commit", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	readmePath := filepath.Join(mainPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(fmt.Sprintf("# %s\n", opts.Name)), 0644); err != nil {
		step := StepResult{Name: "Initial commit", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	if _, err := runner.Run(mainPath, "add", "-A"); err != nil {
		step := StepResult{Name: "Initial commit", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	if _, err := runner.Run(mainPath, "commit", "-m", "Initial commit"); err != nil {
		step := StepResult{Name: "Initial commit", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Initial commit", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Initial commit", Status: StepDone})

	return phase
}

func runCreateGitHub(runner git.CommandRunner, gh GhRunner, repoPath string, opts CreateOptions, emit func(Event)) Phase {
	phase := Phase{Name: PhaseGitHub}
	phaseName := PhaseGitHub

	// Look up GitHub user
	emit(Event{Phase: phaseName, Step: "Look up GitHub user", Status: StepRunning})
	ghUser, err := gh.RunGh(repoPath, "api", "user", "--jq", ".login")
	if err != nil {
		step := StepResult{Name: "Look up GitHub user", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Look up GitHub user", Status: StepDone, Message: ghUser})
	emit(Event{Phase: phaseName, Step: "Look up GitHub user", Status: StepDone, Message: ghUser})

	// Create GitHub repo (without --source/--push — we push manually after configuring SSH remote)
	emit(Event{Phase: phaseName, Step: "Create GitHub repository", Status: StepRunning})
	ghArgs := []string{"repo", "create", opts.Name, "--" + opts.Visibility}
	if opts.Description != "" {
		ghArgs = append(ghArgs, "--description", opts.Description)
	}
	_, err = gh.RunGh(repoPath, ghArgs...)
	if err != nil {
		step := StepResult{Name: "Create GitHub repository", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create GitHub repository", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create GitHub repository", Status: StepDone})

	// Configure the remote using gh's configured protocol so the push uses the
	// auth the user actually has. Forcing SSH breaks push for an HTTPS-only gh
	// login (gh's default), orphaning the just-created empty GitHub repo.
	emit(Event{Phase: phaseName, Step: "Configure remote", Status: StepRunning})
	barePath := filepath.Join(repoPath, ".bare")
	remoteURL := ghRemoteURL(gh, repoPath, ghUser, opts.Name)
	_, err = runner.Run(barePath, "remote", "set-url", "origin", remoteURL)
	if err != nil {
		step := StepResult{Name: "Configure remote", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Configure remote", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Configure remote", Status: StepDone})

	// Push
	emit(Event{Phase: phaseName, Step: "Push to GitHub", Status: StepRunning})
	mainPath := filepath.Join(repoPath, "main")
	_, err = runner.Run(mainPath, "push", "-u", "origin", "main")
	if err != nil {
		// The empty remote repo from "Create GitHub repository" is left behind;
		// tell the user so they can delete it or push manually before retrying.
		pushErr := fmt.Errorf("%w (an empty GitHub repo %q now exists; delete it or push to it manually before retrying)", err, opts.Name)
		step := StepResult{Name: "Push to GitHub", Status: StepFailed, Error: pushErr}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: pushErr})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Push to GitHub", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Push to GitHub", Status: StepDone})

	// Set remote HEAD
	emit(Event{Phase: phaseName, Step: "Set remote HEAD", Status: StepRunning})
	_, err = runner.Run(barePath, "remote", "set-head", "origin", "main")
	if err != nil {
		step := StepResult{Name: "Set remote HEAD", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Set remote HEAD", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Set remote HEAD", Status: StepDone})

	return phase
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
