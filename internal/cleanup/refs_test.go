package cleanup

import (
	"testing"
)

func TestPruneRemoteRefs(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		pruneOutput string
		wantCount   int
	}{
		{
			name:        "no stale refs",
			pruneOutput: "",
			wantCount:   0,
		},
		{
			name:        "some stale refs",
			pruneOutput: "Pruning origin\nURL: git@github.com:Org/repo.git\n * [would prune] origin/feature/old\n * [would prune] origin/fix/done",
			wantCount:   2,
		},
		{
			name:        "dry run reports count",
			dryRun:      true,
			pruneOutput: "Pruning origin\nURL: git@github.com:Org/repo.git\n * [would prune] origin/feature/old",
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRunner{responses: map[string]mockResponse{
				"/repo:[remote prune origin --dry-run]": {output: tt.pruneOutput},
				"/repo:[fetch --prune origin]":          {output: ""},
			}}
			opts := Options{DryRun: tt.dryRun}
			events := collectEvents(t)

			count, err := PruneRemoteRefs(runner, "/repo", opts, events.emit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tt.wantCount {
				t.Errorf("count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestPruneRemoteRefs_DryRunDoesNotFetch(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:[remote prune origin --dry-run]": {output: " * [would prune] origin/old"},
	}}
	opts := Options{DryRun: true}
	events := collectEvents(t)

	_, err := PruneRemoteRefs(runner, "/repo", opts, events.emit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, call := range runner.calls {
		if call == "/repo:[fetch --prune origin]" {
			t.Error("dry-run should not call fetch --prune")
		}
	}
}
