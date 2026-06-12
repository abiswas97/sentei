package progress

import (
	"errors"
	"testing"
)

// The enum ordering is part of the package contract: consolidation must not
// reorder values that switches and comparisons across the codebase rely on.
func TestStepStatus_OrderingIsStable(t *testing.T) {
	want := []StepStatus{StepPending, StepRunning, StepDone, StepFailed, StepSkipped}
	for i, status := range want {
		if int(status) != i {
			t.Errorf("status %v = %d, want %d (enum ordering is a contract)", status, int(status), i)
		}
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
