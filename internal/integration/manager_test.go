package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/progress"
)

type managerMockShell struct {
	responses map[string]mockShellResponse
	calls     []string
}

type mockShellResponse struct {
	output string
	err    error
}

func (m *managerMockShell) RunShell(dir string, command string) (string, error) {
	key := fmt.Sprintf("%s:shell[%s]", dir, command)
	m.calls = append(m.calls, key)
	if resp, ok := m.responses[key]; ok {
		return resp.output, resp.err
	}
	return "", fmt.Errorf("unexpected shell call: %s", key)
}

func TestEnableIntegration_RunsSetupOnEachWorktree(t *testing.T) {
	shell := &managerMockShell{responses: map[string]mockShellResponse{
		"/repo/main:shell[code-review-graph --version]":            {output: "1.0"},
		"/repo:shell[code-review-graph build --repo '/repo/main']": {output: "built"},
		"/repo/feat:shell[code-review-graph --version]":            {output: "1.0"},
		"/repo:shell[code-review-graph build --repo '/repo/feat']": {output: "built"},
	}}

	integ := codeReviewGraph()
	wtPaths := []string{"/repo/main", "/repo/feat"}
	var events []progress.Event

	EnableIntegration(shell, "/repo", "/repo/main", wtPaths, integ, func(e progress.Event) {
		events = append(events, e)
	})

	type wantEvent struct {
		worktree string
		step     string
		status   progress.StepStatus
	}
	// Filter to non-skipped events for setup verification.
	var setupEvents []progress.Event
	var skipCount int
	for _, ev := range events {
		if ev.Status == progress.StepSkipped {
			skipCount++
		} else {
			setupEvents = append(setupEvents, ev)
		}
	}

	// Should have skip events for deps/install (tool already detected).
	if skipCount == 0 {
		t.Error("expected some skipped events for deps/install (tool detected)")
	}

	// Setup events: Running + Done per worktree.
	want := []wantEvent{
		{"/repo/main", "Setup code-review-graph", progress.StepRunning},
		{"/repo/main", "Setup code-review-graph", progress.StepDone},
		{"/repo/feat", "Setup code-review-graph", progress.StepRunning},
		{"/repo/feat", "Setup code-review-graph", progress.StepDone},
	}
	if len(setupEvents) != len(want) {
		t.Fatalf("setup event count = %d, want %d\nevents: %+v", len(setupEvents), len(want), setupEvents)
	}
	for i, w := range want {
		got := setupEvents[i]
		if got.Phase != w.worktree || got.Step != w.step || got.Status != w.status {
			t.Errorf("event[%d] = {%s, %s, %d}, want {%s, %s, %d}",
				i, got.Phase, got.Step, got.Status, w.worktree, w.step, w.status)
		}
	}
}

func TestDisableIntegration_RemovesArtifacts(t *testing.T) {
	shell := &managerMockShell{responses: map[string]mockShellResponse{
		"/repo/main:shell[ccc reset --all --force]": {output: ""},
		"/repo/feat:shell[ccc reset --all --force]": {output: ""},
	}}

	integ := cocoindexCode()
	wtPaths := []string{"/repo/main", "/repo/feat"}
	var events []progress.Event

	DisableIntegration(shell, wtPaths, integ, func(e progress.Event) {
		events = append(events, e)
	})

	if len(events) == 0 {
		t.Fatal("expected events to be emitted")
	}

	var teardownSteps int
	for _, e := range events {
		if strings.Contains(e.Step, "Teardown") || strings.Contains(e.Step, "Remove") {
			teardownSteps++
		}
	}
	if teardownSteps < 2 {
		t.Errorf("teardown steps = %d, want at least 2", teardownSteps)
	}
}

func TestEnableIntegration_InstallsWhenNotDetected(t *testing.T) {
	shell := &managerMockShell{responses: map[string]mockShellResponse{
		// detect fails (not installed)
		"/repo/wt:shell[code-review-graph --version]": {err: fmt.Errorf("not found")},
		// dependency checks
		"/repo/wt:shell[python3 -c \"import sys; assert sys.version_info >= (3,10)\"]": {output: ""},
		"/repo/wt:shell[pipx --version]":                                               {output: "22.0"},
		// install
		"/repo/wt:shell[pipx install code-review-graph]": {output: "installed"},
		// setup
		"/repo:shell[code-review-graph build --repo '/repo/wt']": {output: "built"},
	}}

	integ := codeReviewGraph()
	var events []progress.Event

	EnableIntegration(shell, "/repo", "/repo/wt", []string{"/repo/wt"}, integ, func(e progress.Event) {
		events = append(events, e)
	})

	var hasInstall bool
	for _, e := range events {
		if strings.Contains(e.Step, "Install") {
			hasInstall = true
		}
		if e.Status == progress.StepFailed {
			t.Errorf("unexpected failure event: %+v", e)
		}
	}
	if !hasInstall {
		t.Errorf("expected Install event, got: %+v", events)
	}
}

func TestEnableIntegration_SetupFailureEmitsFailedEvent(t *testing.T) {
	shell := &managerMockShell{responses: map[string]mockShellResponse{
		"/repo/wt:shell[code-review-graph --version]":            {output: "1.0"},
		"/repo:shell[code-review-graph build --repo '/repo/wt']": {err: fmt.Errorf("build failed")},
	}}

	integ := codeReviewGraph()
	var events []progress.Event

	EnableIntegration(shell, "/repo", "/repo/wt", []string{"/repo/wt"}, integ, func(e progress.Event) {
		events = append(events, e)
	})

	var failed bool
	for _, e := range events {
		if e.Status == progress.StepFailed && strings.Contains(e.Step, "Setup") {
			failed = true
		}
	}
	if !failed {
		t.Errorf("expected progress.StepFailed for setup, got: %+v", events)
	}
}

func TestDisableIntegration_TeardownFailureEmitsFailedEvent(t *testing.T) {
	shell := &managerMockShell{responses: map[string]mockShellResponse{
		"/repo/wt:shell[ccc reset --all --force]": {err: fmt.Errorf("reset failed")},
	}}

	integ := cocoindexCode()
	var events []progress.Event

	DisableIntegration(shell, []string{"/repo/wt"}, integ, func(e progress.Event) {
		events = append(events, e)
	})

	var failed bool
	for _, e := range events {
		if e.Status == progress.StepFailed && strings.Contains(e.Step, "Teardown") {
			failed = true
		}
	}
	if !failed {
		t.Errorf("expected progress.StepFailed for teardown, got: %+v", events)
	}
}

func TestDetectTool_BinaryNameUsesPresenceProbe(t *testing.T) {
	shell := &managerMockShell{responses: map[string]mockShellResponse{
		"/repo/wt:shell[command -v ccc]": {output: "/Users/x/.local/bin/ccc"},
	}}
	if !detectTool(shell, "/repo/wt", cocoindexCode()) {
		t.Error("a binary on PATH must be detected even if it lacks --version")
	}

	missing := &managerMockShell{responses: map[string]mockShellResponse{
		"/repo/wt:shell[command -v ccc]": {err: fmt.Errorf("not found")},
	}}
	if detectTool(missing, "/repo/wt", cocoindexCode()) {
		t.Error("a binary missing from PATH must not be detected")
	}
}
