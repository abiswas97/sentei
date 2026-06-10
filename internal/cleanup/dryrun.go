package cleanup

import (
	"strings"
	"time"

	"github.com/abiswas97/sentei/internal/git"
)

// DryRunResult is one scan's answer to "what would cleanup do?", structured
// for display: counts for the bulk categories and branch names where the
// preview shows them. Safe and aggressive findings come from the same scan
// so the TUI never runs the repository inspection twice.
type DryRunResult struct {
	StaleRefs         int
	ConfigDuplicates  int
	GoneBranches      []string // safe mode deletes these (gone upstream, not in a worktree)
	OrphanedConfigs   int
	PrunableWorktrees int

	// AggressiveBranches are local branches in no worktree and not protected:
	// the additional set only aggressive mode deletes.
	AggressiveBranches []BranchInfo

	Errors []OperationError
}

// BranchInfo carries the metadata the detail portal shows per branch.
type BranchInfo struct {
	Name              string
	LastCommitDate    time.Time
	LastCommitSubject string
	// Merged reports whether the branch is fully merged into HEAD: without
	// --force, aggressive cleanup only deletes merged branches, and the
	// preview must not promise more than that.
	Merged bool
}

// SafeHasWork reports whether a safe-mode cleanup would change anything.
func (r DryRunResult) SafeHasWork() bool {
	return r.StaleRefs > 0 || r.ConfigDuplicates > 0 || len(r.GoneBranches) > 0 ||
		r.OrphanedConfigs > 0 || r.PrunableWorktrees > 0
}

// AggressiveHasWork reports whether aggressive mode would delete branches
// beyond what safe mode does.
func (r DryRunResult) AggressiveHasWork() bool {
	return len(r.AggressiveBranches) > 0
}

// UnmergedAggressiveCount reports how many aggressive candidates are not
// fully merged and therefore survive aggressive mode unless --force is given.
func (r DryRunResult) UnmergedAggressiveCount() int {
	n := 0
	for _, b := range r.AggressiveBranches {
		if !b.Merged {
			n++
		}
	}
	return n
}

// DryRun inspects the repository without mutating it and returns what both
// cleanup modes would do. Individual probe failures are collected in
// Errors; only an unresolvable repository aborts the scan.
func DryRun(runner git.CommandRunner, repoPath string) (DryRunResult, error) {
	configPath, err := resolveConfigPath(runner, repoPath)
	if err != nil {
		return DryRunResult{}, err
	}

	noop := func(Event) {}
	probe := Options{Mode: ModeSafe, DryRun: true}
	var result DryRunResult

	if n, err := PruneRemoteRefs(runner, repoPath, probe, noop); err != nil {
		result.Errors = append(result.Errors, OperationError{Step: "prune-refs", Err: err})
	} else {
		result.StaleRefs = n
	}

	if r, err := DedupConfig(configPath, probe, noop); err != nil {
		result.Errors = append(result.Errors, OperationError{Step: "dedup-config", Err: err})
	} else {
		result.ConfigDuplicates = r.Removed
	}

	if output, err := runner.Run(repoPath, "branch", "-vv"); err != nil {
		result.Errors = append(result.Errors, OperationError{Step: "gone-branches", Err: err})
	} else {
		gone, _ := parseGoneBranches(output)
		result.GoneBranches = gone
	}

	if r, err := PurgeOrphanedBranchConfigs(runner, repoPath, configPath, probe, noop); err != nil {
		result.Errors = append(result.Errors, OperationError{Step: "orphaned-configs", Err: err})
	} else {
		result.OrphanedConfigs = r.Removed
	}

	if output, err := runner.Run(repoPath, "worktree", "list", "--porcelain"); err != nil {
		result.Errors = append(result.Errors, OperationError{Step: "worktree-prune", Err: err})
	} else {
		result.PrunableWorktrees = countPrunable(output)
	}

	if candidates, err := listNonWorktreeCandidates(runner, repoPath); err != nil {
		result.Errors = append(result.Errors, OperationError{Step: "non-wt-branches", Err: err})
	} else if len(candidates) > 0 {
		meta, err := branchMetadata(runner, repoPath)
		if err != nil {
			result.Errors = append(result.Errors, OperationError{Step: "branch-metadata", Err: err})
			meta = nil
		}
		merged := make(map[string]bool)
		if output, err := runner.Run(repoPath, "branch", "--merged", "--format=%(refname:short)"); err != nil {
			result.Errors = append(result.Errors, OperationError{Step: "merged-branches", Err: err})
		} else {
			for _, line := range strings.Split(output, "\n") {
				if b := strings.TrimSpace(line); b != "" {
					merged[b] = true
				}
			}
		}
		for _, name := range candidates {
			result.AggressiveBranches = append(result.AggressiveBranches, BranchInfo{
				Name:              name,
				LastCommitDate:    meta[name].LastCommitDate,
				LastCommitSubject: meta[name].LastCommitSubject,
				Merged:            merged[name],
			})
		}
	}

	return result, nil
}

// branchMetadata fetches commit date and subject for every local branch in
// one git call.
func branchMetadata(runner git.CommandRunner, repoPath string) (map[string]BranchInfo, error) {
	output, err := runner.Run(repoPath, "for-each-ref",
		"--format=%(refname:short)\x1f%(committerdate:iso8601-strict)\x1f%(subject)", "refs/heads/")
	if err != nil {
		return nil, err
	}
	meta := make(map[string]BranchInfo)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "\x1f", 3)
		if len(parts) != 3 {
			continue
		}
		date, _ := time.Parse(time.RFC3339, parts[1])
		meta[parts[0]] = BranchInfo{Name: parts[0], LastCommitDate: date, LastCommitSubject: parts[2]}
	}
	return meta, nil
}
