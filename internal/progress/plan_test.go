package progress

import (
	"math/rand"
	"testing"
)

func TestDeclare_EstablishesTotalsBeforeWork(t *testing.T) {
	events, emit := collectEvents()
	Declare(Plan{Phases: []PlannedPhase{
		{Name: "feat-1", Steps: []PlannedStep{{Name: "Setup a"}, {Name: "Setup b"}}},
	}}, emit)

	states := Snapshot(*events)
	if len(states) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(states))
	}
	p := states[0]
	if p.Total != 2 || p.Done != 0 {
		t.Errorf("declared phase = %d/%d, want 0/2 before any work", p.Done, p.Total)
	}
	if !p.Closed {
		t.Error("non-Open phase must be closed by Declare")
	}
	if p.Settled() {
		t.Error("declared-but-unworked phase must not be settled")
	}
}

func TestValidateStream_LegacyOpenPhaseRetainsDiscoveryCompatibility(t *testing.T) {
	events, emit := collectEvents()
	Declare(Plan{Phases: []PlannedPhase{{Name: "scan", Open: true}}}, emit)
	emit(Event{Phase: "scan", Step: "found-1", Status: StepDone})
	ClosePhase("scan", emit)
	if err := ValidateLegacyStream(*events); err != nil {
		t.Fatalf("legacy open stream rejected before producer migration: %v", err)
	}
}

func TestValidateStream_StrictByDefaultForUnlabeledEvents(t *testing.T) {
	events := []Event{{Phase: "p", Step: "undeclared", Status: StepRunning}}
	if err := ValidateStream(events); err == nil {
		t.Fatal("strict validation accepted unlabeled undeclared work")
	}
}

func TestValidateLegacyStream_AllowsLabeledDiscovery(t *testing.T) {
	events := []Event{{Phase: "p", PhaseLabel: "Phase", Step: "discovered", StepLabel: "Discovered", Status: StepDone}}
	if err := ValidateLegacyStream(events); err != nil {
		t.Fatalf("legacy validation inferred strictness from labels: %v", err)
	}
}

func TestValidateStream_MixedMetadataDoesNotFallBack(t *testing.T) {
	events := []Event{
		{Phase: "p", Step: "declared", Status: StepPending, Of: 1},
		{Phase: "p", Close: true},
		{Phase: "p", PhaseLabel: "Phase", Step: "undeclared", StepLabel: "Undeclared", Status: StepDone},
	}
	if err := ValidateStream(events); err == nil {
		t.Fatal("strict validation silently accepted mixed legacy and stable semantics")
	}
}

func TestSnapshot_CheckpointProgressWithinSteps(t *testing.T) {
	events, emit := collectEvents()
	Declare(Plan{Phases: []PlannedPhase{
		{Name: "Removing worktrees", Steps: []PlannedStep{{Name: "wt-a", Checkpoints: 2}}},
	}}, emit)
	emit(Event{Phase: "Removing worktrees", Step: "wt-a", Status: StepRunning, Checkpoint: 1, Of: 2})

	states := Snapshot(*events)
	step := states[0].Steps[0]
	if step.Reached != 1 || step.Declared != 2 {
		t.Errorf("running step checkpoints = %d/%d, want 1/2", step.Reached, step.Declared)
	}
	if step.Status != StepRunning {
		t.Error("checkpoint progress must not resolve the step")
	}
	reached, declared := CheckpointProgress(states)
	if reached != 1 || declared != 2 {
		t.Errorf("overall checkpoints = %d/%d, want 1/2", reached, declared)
	}
}

func TestSnapshot_CheckpointsNeverRegress(t *testing.T) {
	events := []Event{
		{Phase: "P", Step: "s", Status: StepPending, Of: 3},
		{Phase: "P", Step: "s", Status: StepRunning, Checkpoint: 2, Of: 3},
		{Phase: "P", Step: "s", Status: StepRunning, Checkpoint: 1, Of: 3},
	}
	states := Snapshot(events)
	if got := states[0].Steps[0].Reached; got != 2 {
		t.Errorf("reached = %d, want 2 (stale checkpoint must not regress)", got)
	}
}

func TestSnapshot_ResolutionReachesFinalCheckpoint(t *testing.T) {
	for _, status := range []StepStatus{StepDone, StepFailed, StepSkipped} {
		events := []Event{
			{Phase: "P", Step: "s", Status: StepPending, Of: 3},
			{Phase: "P", Step: "s", Status: status},
		}
		step := Snapshot(events)[0].Steps[0]
		if step.Reached != 3 {
			t.Errorf("status %v: reached = %d, want 3 (resolution reaches the final checkpoint)", status, step.Reached)
		}
	}
}

func TestSnapshot_UndeclaredStreamsMatchStepCounting(t *testing.T) {
	events := []Event{
		{Phase: "P", Step: "a", Status: StepDone},
		{Phase: "P", Step: "b", Status: StepRunning},
	}
	states := Snapshot(events)
	reached, declared := CheckpointProgress(states)
	if reached != states[0].Done || declared != states[0].Total {
		t.Errorf("undeclared stream: checkpoints %d/%d must equal step counts %d/%d",
			reached, declared, states[0].Done, states[0].Total)
	}
}

func TestSettled_RequiresCloseAndCompletion(t *testing.T) {
	cases := []struct {
		name string
		p    PhaseState
		want bool
	}{
		{"open complete phase", PhaseState{Total: 2, Done: 2}, false},
		{"closed incomplete phase", PhaseState{Total: 2, Done: 1, Closed: true}, false},
		{"closed complete phase", PhaseState{Total: 2, Done: 2, Closed: true}, true},
		{"closed empty phase", PhaseState{Closed: true}, false},
	}
	for _, tc := range cases {
		if got := tc.p.Settled(); got != tc.want {
			t.Errorf("%s: Settled() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestValidateStream_FlagsNewStepAfterClose(t *testing.T) {
	events := []Event{
		{Phase: "P", Step: "declared", Status: StepPending, Of: 1},
		{Phase: "P", Close: true},
		{Phase: "P", Step: "declared", Status: StepDone},
		{Phase: "P", Step: "smuggled", Status: StepRunning},
	}
	if err := ValidateStream(events); err == nil {
		t.Error("a new step after close must be flagged")
	}
	if err := ValidateStream(events[:3]); err != nil {
		t.Errorf("work on declared steps after close is legitimate: %v", err)
	}
}

func TestValidateStream_RequiresCompleteDeclarationPrefix(t *testing.T) {
	tests := []struct {
		name   string
		events []Event
	}{
		{
			name: "undeclared work",
			events: []Event{
				{Phase: "p", PhaseLabel: "Phase", Close: true},
				{Phase: "p", Step: "missing", Status: StepRunning},
			},
		},
		{
			name: "declaration after close",
			events: []Event{
				{Phase: "p", PhaseLabel: "Phase", Step: "s", StepLabel: "Step", Status: StepPending, Of: 1},
				{Phase: "p", PhaseLabel: "Phase", Close: true},
				{Phase: "q", Step: "late", Status: StepPending, Of: 1},
			},
		},
		{
			name: "work before every phase closes",
			events: []Event{
				{Phase: "p", PhaseLabel: "Phase", Step: "s", StepLabel: "Step", Status: StepPending, Of: 1},
				{Phase: "q", PhaseLabel: "Other", Step: "t", StepLabel: "Other step", Status: StepPending, Of: 1},
				{Phase: "p", Close: true},
				{Phase: "p", Step: "s", Status: StepRunning},
			},
		},
		{
			name: "declared phase never closes",
			events: []Event{
				{Phase: "p", PhaseLabel: "Phase", Step: "s", StepLabel: "Step", Status: StepPending, Of: 1},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateStream(tc.events); err == nil {
				t.Fatal("invalid stream accepted")
			}
		})
	}
}

func TestValidateStream_RejectsTerminalMutationAndCheckpointErrors(t *testing.T) {
	prefix := []Event{
		{Phase: "p", PhaseLabel: "Phase", Step: "s", StepLabel: "Step", Status: StepPending, Of: 2},
		{Phase: "p", PhaseLabel: "Phase", Close: true},
	}
	tests := []struct {
		name string
		work []Event
	}{
		{"terminal mutation", []Event{{Phase: "p", Step: "s", Status: StepDone}, {Phase: "p", Step: "s", Status: StepFailed}}},
		{"checkpoint regression", []Event{{Phase: "p", Step: "s", Status: StepRunning, Checkpoint: 2}, {Phase: "p", Step: "s", Status: StepRunning, Checkpoint: 1}}},
		{"checkpoint overflow", []Event{{Phase: "p", Step: "s", Status: StepRunning, Checkpoint: 3}}},
		{"checkpoint total changed", []Event{{Phase: "p", Step: "s", Status: StepRunning, Checkpoint: 1, Of: 3}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := append(append([]Event(nil), prefix...), tc.work...)
			if err := ValidateStream(events); err == nil {
				t.Fatal("invalid stream accepted")
			}
		})
	}
}

// Property: across any prefix of any interleaving, per-phase totals and the
// overall checkpoint fill are monotonic non-decreasing.
func TestSnapshot_MonotonicUnderRandomInterleavings(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 50; trial++ {
		var stream []Event
		Declare(Plan{Phases: []PlannedPhase{
			{Name: "A", Steps: []PlannedStep{{Name: "a1", Checkpoints: 2}, {Name: "a2"}}},
			{Name: "B", Steps: []PlannedStep{{Name: "b1", Checkpoints: 3}}},
		}}, func(e Event) { stream = append(stream, e) })

		work := []Event{
			{Phase: "A", Step: "a1", Status: StepRunning, Checkpoint: 1, Of: 2},
			{Phase: "A", Step: "a1", Status: StepDone},
			{Phase: "A", Step: "a2", Status: StepFailed},
			{Phase: "B", Step: "b1", Status: StepRunning, Checkpoint: 1, Of: 3},
			{Phase: "B", Step: "b1", Status: StepRunning, Checkpoint: 2, Of: 3},
			{Phase: "B", Step: "b1", Status: StepDone},
		}
		rng.Shuffle(len(work), func(i, j int) {
			// Preserve per-step event order; shuffle across steps only.
			if work[i].Step == work[j].Step {
				return
			}
			work[i], work[j] = work[j], work[i]
		})
		stream = append(stream, work...)

		lastFill := -1.0
		totals := map[string]int{}
		for i := 1; i <= len(stream); i++ {
			states := Snapshot(stream[:i])
			reached, declared := CheckpointProgress(states)
			if declared > 0 {
				fill := float64(reached) / float64(declared)
				if fill+1e-9 < lastFill {
					t.Fatalf("trial %d prefix %d: fill regressed %.3f -> %.3f", trial, i, lastFill, fill)
				}
				lastFill = fill
			}
			for _, p := range states {
				if p.Total < totals[p.Name] {
					t.Fatalf("trial %d prefix %d: phase %s total regressed %d -> %d", trial, i, p.Name, totals[p.Name], p.Total)
				}
				totals[p.Name] = p.Total
				if p.Done > p.Total {
					t.Fatalf("trial %d prefix %d: phase %s done %d exceeds total %d", trial, i, p.Name, p.Done, p.Total)
				}
			}
		}
	}
}
