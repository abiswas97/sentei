package cleanup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/testutil/mock"
)

// dryRunMock wires the git responses DryRun's scan needs: a gone-upstream
// branch (feature/gone), a worktree-occupied branch (main), a non-worktree
// branch with a live upstream (feature/extra), no origin remote work, and a
// clean config file.
func dryRunMock(t *testing.T) (*mock.Runner, string) {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte("[core]\n\tbare = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := &mock.Runner{Responses: map[string]mock.Response{
		"/repo:[rev-parse --git-common-dir]": {Output: dir},
		"/repo:[remote]":                     {Output: ""},
		"/repo:[branch -vv]": {Output: "  feature/gone abc123 [origin/feature/gone: gone] old work\n" +
			"+ main def456 (/repo/main) [origin/main] latest\n" +
			"  feature/extra aaa111 [origin/feature/extra] wip"},
		"/repo:[worktree list --porcelain]":        {Output: "worktree /repo\nbare\n\nworktree /repo/main\nHEAD def456\nbranch refs/heads/main\n"},
		"/repo:[branch --format=%(refname:short)]": {Output: "feature/gone\nfeature/extra\nmain"},
		"/repo:[for-each-ref --format=%(refname:short)\x1f%(committerdate:iso8601-strict)\x1f%(subject) refs/heads/]": {Output: "feature/gone\x1f2026-02-10T19:00:00+05:30\x1fold work\nfeature/extra\x1f2026-06-01T10:00:00+05:30\x1fwip\nmain\x1f2026-06-10T12:00:00+05:30\x1flatest"},
	}}
	return runner, configPath
}

func TestDryRun_CollectsBothModes(t *testing.T) {
	runner, _ := dryRunMock(t)

	result, err := DryRun(runner, "/repo")
	if err != nil {
		t.Fatalf("DryRun() error: %v", err)
	}

	if len(result.GoneBranches) != 1 || result.GoneBranches[0] != "feature/gone" {
		t.Errorf("GoneBranches = %v, want [feature/gone]", result.GoneBranches)
	}
	// feature/extra and feature/gone are both non-worktree candidates that
	// only aggressive mode would delete; main is protected and in a worktree.
	for _, b := range result.AggressiveBranches {
		if b.Name == "main" {
			t.Fatalf("aggressive candidates include protected branch: %v", result.AggressiveBranches)
		}
	}
	if len(result.AggressiveBranches) != 2 {
		t.Fatalf("AggressiveBranches = %v, want feature/gone + feature/extra", result.AggressiveBranches)
	}
	extra := result.AggressiveBranches[1]
	if extra.Name != "feature/extra" || extra.LastCommitSubject != "wip" || extra.LastCommitDate.IsZero() {
		t.Errorf("expected metadata on aggressive branches, got %+v", extra)
	}
}

func TestDryRun_MutatesNothing(t *testing.T) {
	runner, _ := dryRunMock(t)

	if _, err := DryRun(runner, "/repo"); err != nil {
		t.Fatalf("DryRun() error: %v", err)
	}

	for _, call := range runner.Calls {
		for _, destructive := range []string{"branch -d", "branch -D", "remote prune", "worktree prune", "update-ref"} {
			if strings.Contains(call, destructive) {
				t.Errorf("DryRun issued a mutating git call: %s", call)
			}
		}
	}
}

func TestDryRunResult_HasWork(t *testing.T) {
	cases := []struct {
		name           string
		result         DryRunResult
		wantSafe       bool
		wantAggressive bool
	}{
		{"empty", DryRunResult{}, false, false},
		{"stale refs only", DryRunResult{StaleRefs: 2}, true, false},
		{"gone branches only", DryRunResult{GoneBranches: []string{"a"}}, true, false},
		{"config duplicates only", DryRunResult{ConfigDuplicates: 1}, true, false},
		{"orphaned configs only", DryRunResult{OrphanedConfigs: 1}, true, false},
		{"prunable worktrees only", DryRunResult{PrunableWorktrees: 1}, true, false},
		{"aggressive only", DryRunResult{AggressiveBranches: []BranchInfo{{Name: "x"}}}, false, true},
		{"both", DryRunResult{StaleRefs: 1, AggressiveBranches: []BranchInfo{{Name: "x"}}}, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.result.SafeHasWork(); got != tc.wantSafe {
				t.Errorf("SafeHasWork() = %v, want %v", got, tc.wantSafe)
			}
			if got := tc.result.AggressiveHasWork(); got != tc.wantAggressive {
				t.Errorf("AggressiveHasWork() = %v, want %v", got, tc.wantAggressive)
			}
		})
	}
}

func TestDryRun_ErrorWhenConfigUnresolvable(t *testing.T) {
	runner := &mock.Runner{Responses: map[string]mock.Response{}}
	if _, err := DryRun(runner, "/repo"); err == nil {
		t.Error("expected an error when the repo config cannot be resolved")
	}
}
