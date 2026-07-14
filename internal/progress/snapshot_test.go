package progress

import (
	"errors"
	"reflect"
	"testing"
)

func TestSnapshot_EmptyEvents(t *testing.T) {
	if got := Snapshot(nil); got != nil {
		t.Errorf("Snapshot(nil) = %v, want nil", got)
	}
}

func TestSnapshot_FoldsEventsIntoPhases(t *testing.T) {
	events := []Event{
		{Phase: "Setup", Step: "Create worktree", Status: StepRunning},
		{Phase: "Setup", Step: "Create worktree", Status: StepDone},
		{Phase: "Setup", Step: "Merge base", Status: StepSkipped},
		{Phase: "Deps", Step: "npm install", Status: StepRunning},
		{Phase: "Deps", Step: "npm install", Status: StepFailed},
	}

	got := Snapshot(events)

	if len(got) != 2 {
		t.Fatalf("expected 2 phases, got %d: %+v", len(got), got)
	}
	if got[0].Name != "Setup" || got[1].Name != "Deps" {
		t.Errorf("phase order = [%s %s], want [Setup Deps]", got[0].Name, got[1].Name)
	}

	setup := got[0]
	if setup.Total != 2 || setup.Done != 2 || setup.Failed != 0 {
		t.Errorf("Setup counts = total %d done %d failed %d, want 2/2/0 (skipped counts as done)",
			setup.Total, setup.Done, setup.Failed)
	}
	if setup.Steps[0].Status != StepDone {
		t.Error("a later event for the same step must overwrite its status")
	}

	deps := got[1]
	if deps.Total != 1 || deps.Done != 1 || deps.Failed != 1 {
		t.Errorf("Deps counts = total %d done %d failed %d, want 1/1/1 (failed counts as done)",
			deps.Total, deps.Done, deps.Failed)
	}
}

func TestSnapshot_RunningStepNotCounted(t *testing.T) {
	got := Snapshot([]Event{
		{Phase: "Setup", Step: "Create worktree", Status: StepRunning},
	})

	if got[0].Done != 0 || got[0].Failed != 0 || got[0].Total != 1 {
		t.Errorf("running step counts = total %d done %d failed %d, want 1/0/0",
			got[0].Total, got[0].Done, got[0].Failed)
	}
}

func TestSnapshot_DeterministicAndOrderPreserving(t *testing.T) {
	events := []Event{
		{Phase: "B", Step: "b1", Status: StepRunning},
		{Phase: "A", Step: "a1", Status: StepDone},
		{Phase: "B", Step: "b2", Status: StepDone},
		{Phase: "A", Step: "a2", Status: StepRunning},
	}

	first := Snapshot(events)
	second := Snapshot(events)

	if !reflect.DeepEqual(first, second) {
		t.Errorf("two folds of the same stream differ:\n%+v\n%+v", first, second)
	}
	if first[0].Name != "B" || first[1].Name != "A" {
		t.Errorf("phases must keep first-mention order, got [%s %s]", first[0].Name, first[1].Name)
	}
	if first[0].Steps[0].Name != "b1" || first[0].Steps[1].Name != "b2" {
		t.Errorf("steps must keep first-mention order within their phase, got %+v", first[0].Steps)
	}
}

func TestSnapshot_PreservesErrorAndLabel(t *testing.T) {
	stepErr := errors.New("boom")
	got := Snapshot([]Event{
		{Phase: "p", PhaseLabel: "Readable phase", Step: "s", StepLabel: "Readable step", Status: StepPending, Of: 1},
		{Phase: "p", PhaseLabel: "Readable phase", Close: true},
		{Phase: "p", Step: "s", Status: StepFailed, Error: stepErr},
	})
	if len(got) != 1 || got[0].ID != "p" || got[0].Name != "Readable phase" {
		t.Fatalf("phase = %#v", got)
	}
	step := got[0].Steps[0]
	if step.ID != "s" || step.Name != "Readable step" || !errors.Is(step.Error, stepErr) {
		t.Fatalf("step = %#v", step)
	}
}

func TestWithPendingPhases_InsertsMissingCanonicalPhases(t *testing.T) {
	states := []PhaseState{{Name: "Integrations", Total: 1, Done: 1}}

	got := WithPendingPhases(states, "Setup", "Dependencies", "Integrations")

	if len(got) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(got))
	}
	wantNames := []string{"Setup", "Dependencies", "Integrations"}
	for i, want := range wantNames {
		if got[i].Name != want {
			t.Errorf("phase[%d] = %q, want %q", i, got[i].Name, want)
		}
	}
	if got[0].Total != 0 {
		t.Error("inserted Setup phase should be empty (pending)")
	}
	if got[2].Done != 1 {
		t.Error("existing Integrations state should be carried over")
	}
}

func TestWithPendingPhases_NonCanonicalPhasesKeptAtEnd(t *testing.T) {
	states := []PhaseState{
		{Name: "Extra"},
		{Name: "Setup", Done: 1, Total: 1},
	}

	got := WithPendingPhases(states, "Setup")

	if len(got) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(got))
	}
	if got[0].Name != "Setup" || got[1].Name != "Extra" {
		t.Errorf("order = [%s %s], want canonical first, extras last", got[0].Name, got[1].Name)
	}
}
