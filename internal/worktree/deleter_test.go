package worktree

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/abiswas/wt-sweep/internal/git"
)

func collectEvents(ch <-chan DeletionEvent) []DeletionEvent {
	var events []DeletionEvent
	for e := range ch {
		events = append(events, e)
	}
	return events
}

func TestDeleteWorktrees_AllSuccessful(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo worktree remove --force /work/a": {output: ""},
			"/repo worktree remove --force /work/b": {output: ""},
		},
	}

	worktrees := []git.Worktree{
		{Path: "/work/a", Branch: "refs/heads/a"},
		{Path: "/work/b", Branch: "refs/heads/b"},
	}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(runner, "/repo", worktrees, 5, progress)
	events := collectEvents(progress)

	if result.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", result.FailureCount)
	}
	if len(result.Outcomes) != 2 {
		t.Fatalf("Outcomes length = %d, want 2", len(result.Outcomes))
	}
	for i, o := range result.Outcomes {
		if !o.Success {
			t.Errorf("Outcome[%d] should be success", i)
		}
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

func TestDeleteWorktrees_AllFailed(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo worktree remove --force /work/a": {err: fmt.Errorf("locked")},
			"/repo worktree remove --force /work/b": {err: fmt.Errorf("permission denied")},
		},
	}

	worktrees := []git.Worktree{
		{Path: "/work/a"},
		{Path: "/work/b"},
	}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(runner, "/repo", worktrees, 5, progress)
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

func TestDeleteWorktrees_MixedResults(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo worktree remove --force /work/ok":   {output: ""},
			"/repo worktree remove --force /work/fail": {err: fmt.Errorf("locked")},
		},
	}

	worktrees := []git.Worktree{
		{Path: "/work/ok"},
		{Path: "/work/fail"},
	}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(runner, "/repo", worktrees, 5, progress)
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
	runner := &mockRunner{responses: map[string]mockResponse{}}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(runner, "/repo", nil, 5, progress)
	events := collectEvents(progress)

	if result.SuccessCount != 0 || result.FailureCount != 0 {
		t.Errorf("expected zero counts, got success=%d failure=%d", result.SuccessCount, result.FailureCount)
	}
	if len(events) != 0 {
		t.Errorf("expected no events, got %d", len(events))
	}
}

func TestDeleteWorktrees_LockedWorktreeUsesDoubleForce(t *testing.T) {
	runner := &mockRunner{
		responses: map[string]mockResponse{
			"/repo worktree remove --force --force /work/locked": {output: ""},
		},
	}

	worktrees := []git.Worktree{
		{Path: "/work/locked", IsLocked: true},
	}

	progress := make(chan DeletionEvent, 20)
	result := DeleteWorktrees(runner, "/repo", worktrees, 5, progress)
	collectEvents(progress)

	if result.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", result.FailureCount)
	}
}

func TestDeleteWorktrees_ConcurrencyBound(t *testing.T) {
	var maxConcurrent atomic.Int32
	var current atomic.Int32

	runner := &concurrencyTrackingRunner{
		inner: &mockRunner{
			responses: map[string]mockResponse{
				"/repo worktree remove --force /work/1": {output: ""},
				"/repo worktree remove --force /work/2": {output: ""},
				"/repo worktree remove --force /work/3": {output: ""},
				"/repo worktree remove --force /work/4": {output: ""},
				"/repo worktree remove --force /work/5": {output: ""},
			},
		},
		current:       &current,
		maxConcurrent: &maxConcurrent,
	}

	worktrees := make([]git.Worktree, 5)
	for i := range worktrees {
		worktrees[i] = git.Worktree{Path: fmt.Sprintf("/work/%d", i+1)}
	}

	progress := make(chan DeletionEvent, 50)
	DeleteWorktrees(runner, "/repo", worktrees, 2, progress)
	collectEvents(progress)

	if maxConcurrent.Load() > 2 {
		t.Errorf("max concurrent = %d, want <= 2", maxConcurrent.Load())
	}
}

type concurrencyTrackingRunner struct {
	inner         *mockRunner
	current       *atomic.Int32
	maxConcurrent *atomic.Int32
}

func (c *concurrencyTrackingRunner) Run(dir string, args ...string) (string, error) {
	cur := c.current.Add(1)
	for {
		old := c.maxConcurrent.Load()
		if cur <= old || c.maxConcurrent.CompareAndSwap(old, cur) {
			break
		}
	}
	defer c.current.Add(-1)
	return c.inner.Run(dir, args...)
}
