package worktree

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/abiswas97/sentei/internal/git"
)

func collectEvents(ch <-chan DeletionEvent) []DeletionEvent {
	var events []DeletionEvent
	for e := range ch {
		events = append(events, e)
	}
	return events
}

func TestDeleteWorktrees_DeletesRealDirectories(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	worktrees := []git.Worktree{
		{Path: dirA, Branch: "refs/heads/a"},
		{Path: dirB, Branch: "refs/heads/b"},
	}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(os.RemoveAll, worktrees, 5, progress)
	events := collectEvents(progress)

	if result.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", result.FailureCount)
	}
	if _, err := os.Stat(dirA); !os.IsNotExist(err) {
		t.Errorf("directory %s should be deleted", dirA)
	}
	if _, err := os.Stat(dirB); !os.IsNotExist(err) {
		t.Errorf("directory %s should be deleted", dirB)
	}

	var started, completed int
	for _, e := range events {
		switch e.Type {
		case DeletionStarted:
			started++
		case DeletionCompleted:
			completed++
		case DeletionFailed:
			t.Error("unexpected failure event")
		}
	}
	if started != 2 {
		t.Errorf("started events = %d, want 2", started)
	}
	if completed != 2 {
		t.Errorf("completed events = %d, want 2", completed)
	}
}

func TestDeleteWorktrees_DirectoryAlreadyMissing_Succeeds(t *testing.T) {
	worktrees := []git.Worktree{
		{Path: "/nonexistent/path/worktree"},
	}

	progress := make(chan DeletionEvent, 10)
	result := DeleteWorktrees(os.RemoveAll, worktrees, 5, progress)
	collectEvents(progress)

	if result.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1 (missing directory should be a no-op success)", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", result.FailureCount)
	}
}

func TestDeleteWorktrees_RemovalFailure_ReportsFailure(t *testing.T) {
	removalErr := fmt.Errorf("permission denied")
	failingRemover := func(path string) error {
		return removalErr
	}

	worktrees := []git.Worktree{
		{Path: "/work/a"},
		{Path: "/work/b"},
	}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(failingRemover, worktrees, 5, progress)
	collectEvents(progress)

	if result.SuccessCount != 0 {
		t.Errorf("SuccessCount = %d, want 0", result.SuccessCount)
	}
	if result.FailureCount != 2 {
		t.Errorf("FailureCount = %d, want 2", result.FailureCount)
	}
	for i, o := range result.Outcomes {
		if o.Success {
			t.Errorf("Outcome[%d] should be failure", i)
		}
		if o.Error == nil {
			t.Errorf("Outcome[%d] should have error", i)
		}
	}
}

func TestDeleteWorktrees_MixedOutcomes(t *testing.T) {
	goodDir := t.TempDir()
	badPath := "/work/fail"

	removalErr := fmt.Errorf("locked")
	remover := func(path string) error {
		if path == badPath {
			return removalErr
		}
		return os.RemoveAll(path)
	}

	worktrees := []git.Worktree{
		{Path: goodDir},
		{Path: badPath},
	}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(remover, worktrees, 5, progress)
	events := collectEvents(progress)

	if result.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", result.SuccessCount)
	}
	if result.FailureCount != 1 {
		t.Errorf("FailureCount = %d, want 1", result.FailureCount)
	}
	if !result.Outcomes[0].Success {
		t.Error("first outcome should be success")
	}
	if result.Outcomes[1].Success {
		t.Error("second outcome should be failure")
	}

	var failed int
	for _, e := range events {
		if e.Type == DeletionFailed {
			failed++
		}
	}
	if failed != 1 {
		t.Errorf("failed events = %d, want 1", failed)
	}
}

func TestDeleteWorktrees_EmptyInput(t *testing.T) {
	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(os.RemoveAll, nil, 5, progress)
	events := collectEvents(progress)

	if result.SuccessCount != 0 || result.FailureCount != 0 {
		t.Errorf("expected zero counts, got success=%d failure=%d", result.SuccessCount, result.FailureCount)
	}
	if len(events) != 0 {
		t.Errorf("expected no events, got %d", len(events))
	}
}

func TestDeleteWorktrees_ConcurrencyBound(t *testing.T) {
	var current, maxConcurrent atomic.Int32
	// Gate controls when removers complete — holds all goroutines until we release.
	gate := make(chan struct{})

	remover := func(path string) error {
		cur := current.Add(1)
		defer current.Add(-1)
		for {
			old := maxConcurrent.Load()
			if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
				break
			}
		}
		<-gate // Wait for release instead of sleeping.
		return nil
	}

	worktrees := make([]git.Worktree, 5)
	for i := range worktrees {
		worktrees[i] = git.Worktree{Path: fmt.Sprintf("/work/%d", i+1)}
	}

	progress := make(chan DeletionEvent, 50)
	go func() {
		DeleteWorktrees(remover, worktrees, 2, progress)
	}()

	// Release all goroutines.
	close(gate)

	// Collect all events (waits for channel close).
	collectEvents(progress)

	if maxConcurrent.Load() > 2 {
		t.Errorf("max concurrent = %d, want <= 2", maxConcurrent.Load())
	}
}

func TestPruneWorktrees_Success(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo worktree prune": {output: ""},
		},
	}

	err := PruneWorktrees(runner, "/repo")
	if err != nil {
		t.Errorf("PruneWorktrees() error = %v, want nil", err)
	}
}

func TestPruneWorktrees_Failure(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo worktree prune": {err: fmt.Errorf("prune failed")},
		},
	}

	err := PruneWorktrees(runner, "/repo")
	if err == nil {
		t.Error("PruneWorktrees() error = nil, want error")
	}
}
