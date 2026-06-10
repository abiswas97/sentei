package tui

import (
	"testing"

	"github.com/abiswas97/sentei/internal/pipeline"
)

func TestBuildPhaseDisplays_EmptyEvents(t *testing.T) {
	if got := buildPhaseDisplays(nil); got != nil {
		t.Errorf("buildPhaseDisplays(nil) = %v, want nil", got)
	}
}

func TestBuildPhaseDisplays_FoldsEventsIntoPhases(t *testing.T) {
	events := []pipeline.Event{
		{Phase: "Setup", Step: "Create worktree", Status: pipeline.StepRunning},
		{Phase: "Setup", Step: "Create worktree", Status: pipeline.StepDone},
		{Phase: "Setup", Step: "Merge base", Status: pipeline.StepSkipped},
		{Phase: "Deps", Step: "npm install", Status: pipeline.StepRunning},
		{Phase: "Deps", Step: "npm install", Status: pipeline.StepFailed},
	}

	got := buildPhaseDisplays(events)

	if len(got) != 2 {
		t.Fatalf("expected 2 phases, got %d: %+v", len(got), got)
	}
	if got[0].name != "Setup" || got[1].name != "Deps" {
		t.Errorf("phase order = [%s %s], want [Setup Deps]", got[0].name, got[1].name)
	}

	setup := got[0]
	if setup.total != 2 || setup.done != 2 || setup.failed != 0 {
		t.Errorf("Setup counts = total %d done %d failed %d, want 2/2/0 (skipped counts as done)",
			setup.total, setup.done, setup.failed)
	}
	if setup.steps[0].status != pipeline.StepDone {
		t.Error("a later event for the same step must overwrite its status")
	}

	deps := got[1]
	if deps.total != 1 || deps.done != 1 || deps.failed != 1 {
		t.Errorf("Deps counts = total %d done %d failed %d, want 1/1/1 (failed counts as done)",
			deps.total, deps.done, deps.failed)
	}
}

func TestBuildPhaseDisplays_RunningStepNotCounted(t *testing.T) {
	got := buildPhaseDisplays([]pipeline.Event{
		{Phase: "Setup", Step: "Create worktree", Status: pipeline.StepRunning},
	})

	if got[0].done != 0 || got[0].failed != 0 || got[0].total != 1 {
		t.Errorf("running step counts = total %d done %d failed %d, want 1/0/0",
			got[0].total, got[0].done, got[0].failed)
	}
}

func TestWithPendingPhases_InsertsMissingCanonicalPhases(t *testing.T) {
	displays := []phaseDisplay{{name: "Integrations", total: 1, done: 1}}

	got := withPendingPhases(displays, "Setup", "Dependencies", "Integrations")

	if len(got) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(got))
	}
	wantNames := []string{"Setup", "Dependencies", "Integrations"}
	for i, want := range wantNames {
		if got[i].name != want {
			t.Errorf("phase[%d] = %q, want %q", i, got[i].name, want)
		}
	}
	if got[0].total != 0 {
		t.Error("inserted Setup phase should be empty (pending)")
	}
	if got[2].done != 1 {
		t.Error("existing Integrations display should be carried over")
	}
}

func TestWithPendingPhases_NonCanonicalPhasesKeptAtEnd(t *testing.T) {
	displays := []phaseDisplay{
		{name: "Extra"},
		{name: "Setup", done: 1, total: 1},
	}

	got := withPendingPhases(displays, "Setup")

	if len(got) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(got))
	}
	if got[0].name != "Setup" || got[1].name != "Extra" {
		t.Errorf("order = [%s %s], want canonical first, extras last", got[0].name, got[1].name)
	}
}
