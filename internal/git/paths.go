package git

import (
	"path/filepath"
	"strings"
)

// WorktreeDirName returns the directory name sentei uses for a branch's
// worktree. Slashes are flattened ("feature/auth" -> "feature-auth") so every
// worktree sits directly under the repo root instead of nesting.
func WorktreeDirName(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

// WorktreePath returns the path of the worktree directory for a branch.
func WorktreePath(repoPath, branch string) string {
	return filepath.Join(repoPath, WorktreeDirName(branch))
}
