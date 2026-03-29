package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CommandRunner interface {
	Run(dir string, args ...string) (string, error)
}

type GitRunner struct{}

func (r *GitRunner) Run(dir string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", fullArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ShellRunner executes arbitrary shell commands (not git-specific).
type ShellRunner interface {
	RunShell(dir string, command string) (string, error)
}

type DefaultShellRunner struct{}

func (r *DefaultShellRunner) RunShell(dir string, command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", command, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

type DelayRunner struct {
	Inner CommandRunner
	Delay time.Duration
}

func (r *DelayRunner) Run(dir string, args ...string) (string, error) {
	time.Sleep(r.Delay)
	return r.Inner.Run(dir, args...)
}

func ValidateRepository(runner CommandRunner, repoPath string) error {
	_, err := runner.Run(repoPath, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}
	return nil
}

func ListWorktrees(runner CommandRunner, repoPath string) ([]Worktree, error) {
	if err := ValidateRepository(runner, repoPath); err != nil {
		return nil, err
	}

	output, err := runner.Run(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	return ParsePorcelain(output)
}
