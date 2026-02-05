package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
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
