package progress

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecution_FinishSettlesDistinctStepsWithEqualLabels(t *testing.T) {
	plan := Plan{Phases: []PlannedPhase{{ID: "integrations", Label: "Integrations", Steps: []PlannedStep{
		{ID: "ccc.copy-index", Label: "Copy index from main"},
		{ID: "crg.copy-index", Label: "Copy index from main"},
	}}}}
	var events []Event
	x, err := Start(plan, func(ev Event) { events = append(events, ev) })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Done("integrations", "ccc.copy-index", ""); err != nil {
		t.Fatal(err)
	}
	if err := x.Finish("blocked by earlier failure"); err != nil {
		t.Fatal(err)
	}

	states := Snapshot(events)
	if len(states) != 1 || len(states[0].Steps) != 2 {
		t.Fatalf("states = %#v", states)
	}
	if !states[0].Settled() {
		t.Fatalf("phase did not settle: %#v", states[0])
	}
	if states[0].Steps[0].ID != "ccc.copy-index" || states[0].Steps[1].ID != "crg.copy-index" {
		t.Fatalf("step IDs collapsed: %#v", states[0].Steps)
	}
	if states[0].Steps[1].Status != StepSkipped {
		t.Fatalf("second step = %#v", states[0].Steps[1])
	}
	if states[0].Steps[1].Message != "blocked by earlier failure" {
		t.Fatalf("skip reason = %q", states[0].Steps[1].Message)
	}
	if err := ValidateStream(events); err != nil {
		t.Fatalf("execution emitted invalid stream: %v", err)
	}
}

func TestExecution_RejectsUndeclaredAndTerminalMutation(t *testing.T) {
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Label: "Phase", Steps: []PlannedStep{{ID: "s", Label: "Step"}}}}}, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Done("p", "missing", ""); err == nil {
		t.Fatal("undeclared step accepted")
	}
	if _, err := x.Done("p", "s", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := x.Fail("p", "s", errors.New("late")); err == nil {
		t.Fatal("terminal mutation accepted")
	}
}

func TestExecution_CheckpointsAreMonotonicUnderConcurrency(t *testing.T) {
	var mu sync.Mutex
	var events []Event
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Label: "Phase", Steps: []PlannedStep{{ID: "s", Label: "Step", Checkpoints: 2}}}}}, func(ev Event) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = x.Running("p", "s", 1, "first")
		}()
	}
	wg.Wait()
	if err := x.Running("p", "s", 2, "second"); err != nil {
		t.Fatal(err)
	}
	if _, err := x.Done("p", "s", "complete"); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	gotEvents := append([]Event(nil), events...)
	mu.Unlock()
	if err := ValidateStream(gotEvents); err != nil {
		t.Fatalf("concurrent stream invalid: %v", err)
	}
	step := Snapshot(gotEvents)[0].Steps[0]
	if step.Reached != 2 || step.Declared != 2 {
		t.Fatalf("checkpoint progress = %d/%d, want 2/2", step.Reached, step.Declared)
	}
}

func TestExecution_StartRejectsInvalidIDs(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
	}{
		{"empty phase ID", Plan{Phases: []PlannedPhase{{Label: "Phase"}}}},
		{"duplicate phase ID", Plan{Phases: []PlannedPhase{{ID: "p"}, {ID: "p"}}}},
		{"empty step ID", Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{Label: "Step"}}}}}},
		{"duplicate step ID", Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "s"}, {ID: "s"}}}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Start(tc.plan, func(Event) {}); err == nil {
				t.Fatal("invalid plan accepted")
			}
		})
	}
}

func TestExecution_StartRejectsMixedStableAndLegacyPlanFields(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
	}{
		{"phase name", Plan{Phases: []PlannedPhase{{ID: "p", Label: "Phase", Name: "legacy"}}}},
		{"phase open", Plan{Phases: []PlannedPhase{{ID: "p", Label: "Phase", Open: true}}}},
		{"step name", Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "s", Label: "Step", Name: "legacy"}}}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Start(tc.plan, func(Event) {}); err == nil {
				t.Fatal("mixed stable and legacy plan fields accepted")
			}
		})
	}
}

func TestExecution_StartEmitsCompleteNormalizedDeclarationPrefix(t *testing.T) {
	var events []Event
	_, err := Start(Plan{Phases: []PlannedPhase{
		{ID: "a", Label: "Alpha", Steps: []PlannedStep{{ID: "one", Label: "One", Checkpoints: 0}}},
		{ID: "b", Label: "Beta", Steps: []PlannedStep{{ID: "two", Label: "Two", Checkpoints: 3}}},
	}}, func(ev Event) { events = append(events, ev) })
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 4 {
		t.Fatalf("events = %#v, want two declarations then two closes", events)
	}
	if events[0].Status != StepPending || events[0].Of != 1 || events[0].PhaseLabel != "Alpha" || events[0].StepLabel != "One" {
		t.Fatalf("first declaration = %#v", events[0])
	}
	if events[1].Status != StepPending || events[1].Of != 3 {
		t.Fatalf("second declaration = %#v", events[1])
	}
	if !events[2].Close || !events[3].Close {
		t.Fatalf("declaration was not a complete prefix: %#v", events)
	}
	if err := ValidateStream(events); err != nil {
		t.Fatalf("declaration stream invalid: %v", err)
	}
}

func TestExecution_RejectsCheckpointRegressionAndOverflow(t *testing.T) {
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "s", Checkpoints: 2}}}}}, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	if err := x.Running("p", "s", 1, "first"); err != nil {
		t.Fatal(err)
	}
	if err := x.Running("p", "s", 0, "stale"); err == nil {
		t.Fatal("checkpoint regression accepted")
	}
	if err := x.Running("p", "s", 3, "overflow"); err == nil {
		t.Fatal("checkpoint overflow accepted")
	}
}

func TestExecution_SkipPendingLeavesRunningStepAlone(t *testing.T) {
	var events []Event
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "running"}, {ID: "pending"}}}}}, func(ev Event) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := x.Running("p", "running", 0, ""); err != nil {
		t.Fatal(err)
	}
	if err := x.SkipPending("p", "blocked"); err != nil {
		t.Fatal(err)
	}
	states := Snapshot(events)
	if states[0].Steps[0].Status != StepRunning || states[0].Steps[1].Status != StepSkipped {
		t.Fatalf("steps = %#v", states[0].Steps)
	}
}

func TestExecution_RunDoesNotHoldLockWhileFunctionRuns(t *testing.T) {
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "run"}, {ID: "other"}}}}}, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := x.Run("p", "run", func() (string, error) {
			close(entered)
			<-release
			return "ok", nil
		})
		done <- err
	}()
	<-entered
	if _, err := x.Skip("p", "other", "not needed"); err != nil {
		t.Fatalf("parallel transition blocked or failed: %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestExecution_EmitCallbackCanSynchronouslyTransitionAnotherStep(t *testing.T) {
	var x *Execution
	var events []Event
	var callbackErr error
	x, startErr := Start(Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "a"}, {ID: "b"}}}}}, func(ev Event) {
		events = append(events, ev)
		if ev.Step == "a" && ev.Status == StepDone {
			callbackErr = x.Running("p", "b", 0, "triggered by a")
			if callbackErr == nil {
				_, callbackErr = x.Skip("p", "b", "triggered by a")
			}
		}
	})
	if startErr != nil {
		t.Fatal(startErr)
	}

	done := make(chan error, 1)
	go func() {
		_, err := x.Done("p", "a", "complete")
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("emit callback deadlocked while transitioning another step")
	}
	if callbackErr != nil {
		t.Fatalf("callback transition: %v", callbackErr)
	}
	if err := ValidateStream(events); err != nil {
		t.Fatalf("reentrant emission order invalid: %v", err)
	}
	statuses := []StepStatus{events[len(events)-3].Status, events[len(events)-2].Status, events[len(events)-1].Status}
	if statuses[0] != StepDone || statuses[1] != StepRunning || statuses[2] != StepSkipped {
		t.Fatalf("reentrant event order = %v, want done/running/skipped", statuses)
	}
	states := Snapshot(events)
	if states[0].Steps[0].Status != StepDone || states[0].Steps[1].Status != StepSkipped {
		t.Fatalf("states = %#v", states)
	}
}

func TestExecution_RunClaimsPendingStepOnce(t *testing.T) {
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "s"}}}}}, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	var invocations atomic.Int32
	entered := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		_, err := x.Run("p", "s", func() (string, error) {
			invocations.Add(1)
			close(entered)
			<-release
			return "first", nil
		})
		firstDone <- err
	}()
	<-entered

	_, secondErr := x.Run("p", "s", func() (string, error) {
		invocations.Add(1)
		return "second", nil
	})
	if secondErr == nil {
		t.Fatal("second Run claimed an already-running step")
	}
	if got := invocations.Load(); got != 1 {
		t.Fatalf("step body invoked %d times, want 1", got)
	}
	close(release)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
}

func TestExecution_FinishWaitsForQueuedTerminalEvents(t *testing.T) {
	delivered := make(chan Event, 16)
	callbackBlocked := make(chan struct{})
	releaseCallback := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(releaseCallback) })

	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "a"}, {ID: "b"}}}}}, func(ev Event) {
		if ev.Step == "a" && ev.Status == StepRunning {
			close(callbackBlocked)
			<-releaseCallback
		}
		delivered <- ev
	})
	if err != nil {
		t.Fatal(err)
	}
	runningDone := make(chan error, 1)
	go func() { runningDone <- x.Running("p", "a", 0, "working") }()
	<-callbackBlocked

	finishStarted := make(chan struct{})
	finishDone := make(chan error, 1)
	go func() {
		close(finishStarted)
		finishDone <- x.Finish("shutdown")
	}()
	<-finishStarted
	select {
	case err := <-finishDone:
		t.Fatalf("Finish returned before blocked callback delivery: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	releaseOnce.Do(func() { close(releaseCallback) })
	if err := <-runningDone; err != nil {
		t.Fatal(err)
	}
	if err := <-finishDone; err != nil {
		t.Fatal(err)
	}
	close(delivered)
	events := make([]Event, 0, 6)
	for ev := range delivered {
		events = append(events, ev)
	}
	if err := ValidateStream(events); err != nil {
		t.Fatalf("flushed stream invalid: %v", err)
	}
	if len(events) != 6 {
		t.Fatalf("delivered %d events, want 6: %#v", len(events), events)
	}
	tail := events[3:]
	if tail[0].Step != "a" || tail[0].Status != StepRunning ||
		tail[1].Step != "a" || tail[1].Status != StepSkipped ||
		tail[2].Step != "b" || tail[2].Status != StepSkipped {
		t.Fatalf("terminal delivery order = %#v", tail)
	}
}

func TestExecution_CallbackPanicBecomesDeliveryError(t *testing.T) {
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Steps: []PlannedStep{{ID: "s"}}}}}, func(ev Event) {
		if ev.Status == StepRunning {
			panic("callback boom")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	var runningErr error
	var escaped any
	func() {
		defer func() { escaped = recover() }()
		runningErr = x.Running("p", "s", 0, "working")
	}()
	if escaped != nil {
		t.Fatalf("callback panic escaped instead of becoming a delivery error: %v", escaped)
	}
	if runningErr == nil {
		t.Fatal("Running hid callback delivery failure")
	}
	if !strings.Contains(runningErr.Error(), "callback boom") {
		t.Fatalf("Running delivery error lost panic detail: %v", runningErr)
	}
	if err := x.Finish("shutdown"); err == nil {
		t.Fatal("Finish did not propagate stored callback delivery failure")
	}
}
