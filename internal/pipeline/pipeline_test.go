package pipeline

import (
	"errors"
	"testing"
)

func collectEvents() (*[]Event, func(Event)) {
	events := &[]Event{}
	return events, func(e Event) { *events = append(*events, e) }
}

func TestRunStep_Success(t *testing.T) {
	events, emit := collectEvents()

	result := RunStep("Phase", "Step", emit, func() (string, error) {
		return "did it", nil
	})

	want := StepResult{Name: "Step", Status: StepDone, Message: "did it"}
	if result != want {
		t.Errorf("result = %+v, want %+v", result, want)
	}
	if len(*events) != 2 {
		t.Fatalf("got %d events, want 2 (Running, Done)", len(*events))
	}
	if (*events)[0].Status != StepRunning {
		t.Errorf("first event status = %v, want StepRunning", (*events)[0].Status)
	}
	if (*events)[1].Status != StepDone || (*events)[1].Message != "did it" {
		t.Errorf("second event = %+v, want Done with message", (*events)[1])
	}
}

func TestRunStep_Failure(t *testing.T) {
	events, emit := collectEvents()
	stepErr := errors.New("boom")

	result := RunStep("Phase", "Step", emit, func() (string, error) {
		return "", stepErr
	})

	if result.Status != StepFailed || !errors.Is(result.Error, stepErr) {
		t.Errorf("result = %+v, want failed with boom", result)
	}
	if len(*events) != 2 {
		t.Fatalf("got %d events, want 2 (Running, Failed)", len(*events))
	}
	if (*events)[1].Status != StepFailed || !errors.Is((*events)[1].Error, stepErr) {
		t.Errorf("second event = %+v, want Failed with boom", (*events)[1])
	}
}

func TestPhaseRecorder(t *testing.T) {
	events, emit := collectEvents()
	rec := NewPhaseRecorder("Phase", emit)

	if ok := rec.Step("first", func() (string, error) { return "msg", nil }); !ok {
		t.Error("successful step should report ok")
	}
	if ok := rec.Step("second", func() (string, error) { return "", errors.New("nope") }); ok {
		t.Error("failed step should report !ok")
	}
	rec.Skip("third", "not needed")
	rec.Done("fourth", "manual done")
	rec.Fail("fifth", errors.New("precondition"))
	rec.Record(StepResult{Name: "sixth", Status: StepSkipped})
	rec.Emit("first", StepRunning, "progress note")

	phase := rec.Phase()
	if phase.Name != "Phase" {
		t.Errorf("phase name = %q, want Phase", phase.Name)
	}
	wantStatuses := []StepStatus{StepDone, StepFailed, StepSkipped, StepDone, StepFailed, StepSkipped}
	if len(phase.Steps) != len(wantStatuses) {
		t.Fatalf("got %d steps, want %d", len(phase.Steps), len(wantStatuses))
	}
	for i, want := range wantStatuses {
		if phase.Steps[i].Status != want {
			t.Errorf("step[%d] status = %v, want %v", i, phase.Steps[i].Status, want)
		}
	}
	if !phase.HasFailures() {
		t.Error("phase with failed steps must report HasFailures")
	}

	// Record must not emit; everything else must. 2 (Step) + 2 (Step) + 1 (Skip)
	// + 1 (Done) + 1 (Fail) + 0 (Record) + 1 (Emit) = 8.
	if len(*events) != 8 {
		t.Errorf("got %d events, want 8", len(*events))
	}
}

func TestFirstFailure(t *testing.T) {
	err := errors.New("broke")
	phases := []Phase{
		{Name: "A", Steps: []StepResult{{Name: "ok", Status: StepDone}}},
		{Name: "B", Steps: []StepResult{{Name: "bad", Status: StepFailed, Error: err}, {Name: "worse", Status: StepFailed}}},
	}

	phaseName, step, ok := FirstFailure(phases)
	if !ok || phaseName != "B" || step.Name != "bad" || !errors.Is(step.Error, err) {
		t.Errorf("FirstFailure = (%q, %+v, %v), want (B, bad, true)", phaseName, step, ok)
	}

	if _, _, ok := FirstFailure(phases[:1]); ok {
		t.Error("FirstFailure on all-done phases must report !ok")
	}
}

func TestPhasesHaveFailures(t *testing.T) {
	clean := []Phase{{Name: "A", Steps: []StepResult{{Status: StepDone}, {Status: StepSkipped}}}}
	if PhasesHaveFailures(clean) {
		t.Error("clean phases must not report failures")
	}
	if !PhasesHaveFailures(append(clean, Phase{Steps: []StepResult{{Status: StepFailed}}})) {
		t.Error("a failed step must be reported")
	}
}
