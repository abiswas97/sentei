package worktree

import (
	"fmt"
	"testing"
	"time"

	"github.com/abiswas/wt-sweep/internal/git"
)

func TestParseStatusPorcelain(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantUncommitted bool
		wantUntracked   bool
	}{
		{
			name:            "empty output (clean)",
			input:           "",
			wantUncommitted: false,
			wantUntracked:   false,
		},
		{
			name:            "whitespace only (clean)",
			input:           "  \n  ",
			wantUncommitted: false,
			wantUntracked:   false,
		},
		{
			name:            "uncommitted changes only (modified)",
			input:           " M internal/git/worktree.go",
			wantUncommitted: true,
			wantUntracked:   false,
		},
		{
			name:            "staged addition",
			input:           "A  newfile.go",
			wantUncommitted: true,
			wantUntracked:   false,
		},
		{
			name:            "untracked files only",
			input:           "?? untracked.txt\n?? another.txt",
			wantUncommitted: false,
			wantUntracked:   true,
		},
		{
			name:            "both uncommitted and untracked",
			input:           " M changed.go\n?? newfile.txt",
			wantUncommitted: true,
			wantUntracked:   true,
		},
		{
			name:            "deleted file",
			input:           " D removed.go",
			wantUncommitted: true,
			wantUntracked:   false,
		},
		{
			name:            "renamed file",
			input:           "R  old.go -> new.go",
			wantUncommitted: true,
			wantUntracked:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUncommitted, gotUntracked := ParseStatusPorcelain(tt.input)
			if gotUncommitted != tt.wantUncommitted {
				t.Errorf("hasUncommitted = %v, want %v", gotUncommitted, tt.wantUncommitted)
			}
			if gotUntracked != tt.wantUntracked {
				t.Errorf("hasUntracked = %v, want %v", gotUntracked, tt.wantUntracked)
			}
		})
	}
}

func TestParseCommitDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "valid date",
			input:   "2024-01-15 10:30:00 -0500",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.FixedZone("", -5*3600)),
			wantErr: false,
		},
		{
			name:    "empty output (orphan branch)",
			input:   "",
			want:    time.Time{},
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "  \n  ",
			want:    time.Time{},
			wantErr: false,
		},
		{
			name:    "malformed date",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommitDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

type mockRunner struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	output string
	err    error
}

func (m *mockRunner) Run(dir string, args ...string) (string, error) {
	key := dir
	for _, a := range args {
		key += " " + a
	}
	resp, ok := m.responses[key]
	if !ok {
		return "", fmt.Errorf("unexpected command: %s", key)
	}
	return resp.output, resp.err
}

func TestEnrichWorktree_Success(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/work/feature log -1 --format=%ai": {output: "2024-06-01 12:00:00 +0000"},
			"/work/feature log -1 --format=%s":  {output: "Add feature X"},
			"/work/feature status --porcelain":  {output: " M file.go\n?? new.txt"},
		},
	}

	wt := &git.Worktree{Path: "/work/feature"}
	enrichWorktree(runner, wt)

	if wt.EnrichmentError != "" {
		t.Fatalf("unexpected error: %s", wt.EnrichmentError)
	}
	if !wt.IsEnriched {
		t.Error("expected IsEnriched=true")
	}
	if wt.LastCommitSubject != "Add feature X" {
		t.Errorf("LastCommitSubject = %q, want %q", wt.LastCommitSubject, "Add feature X")
	}
	if !wt.HasUncommittedChanges {
		t.Error("expected HasUncommittedChanges=true")
	}
	if !wt.HasUntrackedFiles {
		t.Error("expected HasUntrackedFiles=true")
	}
	expectedDate := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if !wt.LastCommitDate.Equal(expectedDate) {
		t.Errorf("LastCommitDate = %v, want %v", wt.LastCommitDate, expectedDate)
	}
}

func TestEnrichWorktree_LogCommandFails(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/work/broken log -1 --format=%ai": {err: fmt.Errorf("git log: fatal: not a git repository")},
		},
	}

	wt := &git.Worktree{Path: "/work/broken"}
	enrichWorktree(runner, wt)

	if wt.EnrichmentError == "" {
		t.Error("expected EnrichmentError to be set")
	}
	if wt.IsEnriched {
		t.Error("expected IsEnriched=false")
	}
}

func TestEnrichWorktree_StatusCommandFails(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/work/broken log -1 --format=%ai": {output: "2024-01-01 00:00:00 +0000"},
			"/work/broken log -1 --format=%s":  {output: "Some commit"},
			"/work/broken status --porcelain":  {err: fmt.Errorf("git status: permission denied")},
		},
	}

	wt := &git.Worktree{Path: "/work/broken"}
	enrichWorktree(runner, wt)

	if wt.EnrichmentError == "" {
		t.Error("expected EnrichmentError to be set")
	}
	if wt.IsEnriched {
		t.Error("expected IsEnriched=false")
	}
}

func TestEnrichWorktrees_MixedSlice(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/work/normal log -1 --format=%ai": {output: "2024-03-15 09:00:00 +0000"},
			"/work/normal log -1 --format=%s":  {output: "Normal commit"},
			"/work/normal status --porcelain":  {output: ""},
		},
	}

	worktrees := []git.Worktree{
		{Path: "/repo", IsBare: true},
		{Path: "/work/normal"},
		{Path: "/work/gone", IsPrunable: true, PruneReason: "directory not found"},
	}

	result := EnrichWorktrees(runner, worktrees, 5)

	if result[0].IsEnriched {
		t.Error("bare entry should not be enriched")
	}
	if !result[1].IsEnriched {
		t.Error("normal entry should be enriched")
	}
	if result[1].LastCommitSubject != "Normal commit" {
		t.Errorf("LastCommitSubject = %q, want %q", result[1].LastCommitSubject, "Normal commit")
	}
	if result[2].IsEnriched {
		t.Error("prunable entry should not be enriched")
	}
}

func TestEnrichWorktrees_PartialFailure(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/work/ok log -1 --format=%ai":     {output: "2024-03-15 09:00:00 +0000"},
			"/work/ok log -1 --format=%s":      {output: "Good commit"},
			"/work/ok status --porcelain":      {output: ""},
			"/work/broken log -1 --format=%ai": {err: fmt.Errorf("directory missing")},
		},
	}

	worktrees := []git.Worktree{
		{Path: "/work/ok"},
		{Path: "/work/broken"},
	}

	result := EnrichWorktrees(runner, worktrees, 5)

	if !result[0].IsEnriched {
		t.Error("first worktree should be enriched")
	}
	if result[0].EnrichmentError != "" {
		t.Errorf("first worktree should have no error, got: %s", result[0].EnrichmentError)
	}
	if result[1].IsEnriched {
		t.Error("second worktree should not be enriched")
	}
	if result[1].EnrichmentError == "" {
		t.Error("second worktree should have enrichment error")
	}
}
