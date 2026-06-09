package git

import "strings"

var protectedBranches = map[string]bool{
	"main":    true,
	"master":  true,
	"develop": true,
	"dev":     true,
}

// IsProtectedBranch reports whether a branch matches the built-in protected
// convention set. Matching is case-insensitive so "Main"/"MASTER" are caught.
// It does not know the repo's actual default branch; prefer IsProtectedBranchWith
// when that is available.
func IsProtectedBranch(branch string) bool {
	name := strings.ToLower(strings.TrimPrefix(branch, "refs/heads/"))
	return protectedBranches[name]
}

// IsProtectedBranchWith additionally protects the repo's actual default branch,
// which may be non-standard (e.g. "production", "trunk", "mainline"). A bare
// detected name (no refs/heads/ prefix) is expected for defaultBranch.
func IsProtectedBranchWith(branch, defaultBranch string) bool {
	if defaultBranch != "" {
		name := strings.TrimPrefix(branch, "refs/heads/")
		if strings.EqualFold(name, defaultBranch) {
			return true
		}
	}
	return IsProtectedBranch(branch)
}

// DetectDefaultBranch returns the repo's default branch by reading HEAD on the
// bare repo. It works whether repoPath is the bare dir (.bare) or a sentei repo
// root (whose .git pointer resolves HEAD to .bare). This is the single source of
// truth for default-branch detection across clone, remove, and protection.
func DetectDefaultBranch(runner CommandRunner, repoPath string) string {
	// A bare clone records the remote's default branch in HEAD. This survives a
	// non-standard default; refs/remotes/origin/HEAD is not created by a bare
	// clone, so reading that always fails.
	if branch, err := runner.Run(repoPath, "symbolic-ref", "--short", "HEAD"); err == nil && branch != "" {
		return branch
	}

	// Fallback: try main, then master.
	for _, candidate := range []string{"main", "master"} {
		if _, err := runner.Run(repoPath, "show-ref", "--verify", "refs/heads/"+candidate); err == nil {
			return candidate
		}
	}

	return "main" // last resort
}
