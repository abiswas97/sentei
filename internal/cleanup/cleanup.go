package cleanup

import (
	"fmt"
	"path/filepath"

	"github.com/abiswas97/sentei/internal/git"
)

type Mode string

const (
	ModeSafe       Mode = "safe"
	ModeAggressive Mode = "aggressive"
)

type Options struct {
	Mode   Mode
	Force  bool
	DryRun bool
}

type Result struct {
	ConfigDedupResult      ConfigResult
	ConfigOrphanResult     ConfigResult
	StaleRefsRemoved       int
	GoneBranchesDeleted    int
	NonWtBranchesDeleted   int
	NonWtBranchesRemaining int
	BranchesSkipped        []SkippedBranch
	Errors                 []OperationError
}

type ConfigResult struct {
	Before  int
	After   int
	Removed int
}

type OperationError struct {
	Step string
	Err  error
}

type SkipReason string

const (
	SkipUnmerged   SkipReason = "not fully merged"
	SkipInWorktree SkipReason = "checked out in worktree"
	SkipProtected  SkipReason = "protected branch"
)

type SkippedBranch struct {
	Name   string
	Reason SkipReason
}

type Event struct {
	Step    string
	Message string
	Level   EventLevel
}

type EventLevel int

const (
	LevelStep EventLevel = iota
	LevelInfo
	LevelWarn
	LevelDetail
)

func resolveConfigPath(runner git.CommandRunner, repoPath string) (string, error) {
	commonDir, err := runner.Run(repoPath, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("resolving config path: %w", err)
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(repoPath, commonDir)
	}
	return filepath.Join(commonDir, "config"), nil
}

func Run(runner git.CommandRunner, repoPath string, opts Options, emit func(Event)) Result {
	return Result{}
}
