package integration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/abiswas97/sentei/internal/progress"
)

type applyShell struct {
	mu        sync.Mutex
	responses map[string]mockShellResponse
	calls     []string
}

type mockShellResponse struct {
	output string
	err    error
}

func findStep(t *testing.T, events []progress.Event, label string) progress.StepState {
	t.Helper()
	for _, phase := range progress.Snapshot(events) {
		for _, step := range phase.Steps {
			if step.Name == label {
				return step
			}
		}
	}
	t.Fatalf("step %q not found", label)
	return progress.StepState{}
}

func (s *applyShell) RunShell(dir, command string) (string, error) {
	key := fmt.Sprintf("%s:shell[%s]", dir, command)
	s.mu.Lock()
	s.calls = append(s.calls, key)
	response, ok := s.responses[key]
	s.mu.Unlock()
	if !ok {
		return "", fmt.Errorf("unexpected shell call: %s", key)
	}
	return response.output, response.err
}

func testIntegration(name string, dependencies ...Dependency) Integration {
	return Integration{
		Name:         name,
		Dependencies: dependencies,
		Detect:       DetectSpec{Command: name + " detect"},
		Install:      InstallSpec{Command: name + " install"},
		Setup:        SetupSpec{Command: name + " setup {path}", WorkingDir: "worktree"},
	}
}

func plannedLabels(plan progress.Plan) []string {
	var labels []string
	for _, phase := range plan.Phases {
		for _, step := range phase.Steps {
			labels = append(labels, phase.Label+"/"+step.Label)
		}
	}
	return labels
}

func collectPreparedEvents(prepared PreparedApply, shell *applyShell) ([]progress.Event, []progress.Phase) {
	var events []progress.Event
	phases, err := prepared.Run(shell, func(event progress.Event) { events = append(events, event) })
	if err != nil {
		panic(err)
	}
	return events, phases
}

func assertSettledStream(t *testing.T, events []progress.Event) {
	t.Helper()
	if err := progress.ValidateStream(events); err != nil {
		t.Fatalf("invalid stream: %v\n%#v", err, events)
	}
	for _, phase := range progress.Snapshot(events) {
		if phase.Total > 0 && !phase.Settled() {
			t.Fatalf("phase did not settle: %#v", phase)
		}
	}
}

func TestPrepareApply_MissingToolPlansPrerequisiteOnce(t *testing.T) {
	integ := testIntegration("tool", Dependency{Name: "dep", Detect: "dep detect"})
	probeDir := "/wt/a"
	shell := &applyShell{responses: map[string]mockShellResponse{
		probeDir + ":shell[tool detect]": {err: errors.New("missing")},
		probeDir + ":shell[dep detect]":  {output: "present"},
	}}
	prepared, err := PrepareApply(shell, "/repo", probeDir, []Integration{integ}, nil, []string{probeDir, "/wt/b"})
	if err != nil {
		t.Fatal(err)
	}
	labels := plannedLabels(prepared.Plan())
	if got := strings.Count(strings.Join(labels, "\n"), "Prerequisites/Install tool"); got != 1 {
		t.Fatalf("install steps = %d, want 1: %v", got, labels)
	}
	if got := strings.Count(strings.Join(labels, "\n"), "Setup tool"); got != 2 {
		t.Fatalf("setup steps = %d, want 2: %v", got, labels)
	}
}

func TestPreparedApply_InstallFailureSkipsEverySetup(t *testing.T) {
	integ := testIntegration("tool")
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[tool detect]":        {err: errors.New("missing")},
		"/wt/a:shell[tool install]":       {err: errors.New("install broke")},
		"/wt/a:shell[tool setup '/wt/a']": {output: "must not run"},
		"/wt/b:shell[tool setup '/wt/b']": {output: "must not run"},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{integ}, nil, []string{"/wt/a", "/wt/b"})
	if err != nil {
		t.Fatal(err)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	states := progress.Snapshot(events)
	for _, phase := range states {
		for _, step := range phase.Steps {
			if strings.HasPrefix(step.Name, "Setup tool") && (step.Status != progress.StepSkipped || !strings.Contains(step.Message, "blocked by Install tool")) {
				t.Fatalf("setup step = %#v", step)
			}
		}
	}
}

func TestPreparedApply_DeclarationTotalIsFixedAcrossPrefixes(t *testing.T) {
	a := testIntegration("alpha")
	b := testIntegration("beta")
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[alpha detect]":        {output: "installed"},
		"/wt/a:shell[beta detect]":         {err: errors.New("missing")},
		"/wt/a:shell[alpha setup '/wt/a']": {output: "ok"},
		"/wt/b:shell[alpha setup '/wt/b']": {output: "ok"},
		"/wt/a:shell[beta install]":        {output: "ok"},
		"/wt/a:shell[beta setup '/wt/a']":  {output: "ok"},
		"/wt/b:shell[beta setup '/wt/b']":  {output: "ok"},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{a, b}, nil, []string{"/wt/a", "/wt/b"})
	if err != nil {
		t.Fatal(err)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	declarationEnd := 0
	for i, event := range events {
		if event.Close {
			declarationEnd = i + 1
		}
	}
	want := -1
	for i := declarationEnd; i <= len(events); i++ {
		_, total := progress.CheckpointProgress(progress.Snapshot(events[:i]))
		if want < 0 {
			want = total
		} else if total != want {
			t.Fatalf("prefix %d total = %d, want fixed %d", i, total, want)
		}
	}
}

func TestPrepareApply_InstalledToolPlansSetupOnlyAndDoesNotReprobe(t *testing.T) {
	integ := testIntegration("tool", Dependency{Name: "dep", Detect: "dep detect"})
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[tool detect]":        {output: "installed"},
		"/wt/a:shell[dep detect]":         {output: "installed"},
		"/wt/a:shell[tool setup '/wt/a']": {output: "ok"},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{integ}, nil, []string{"/wt/a"})
	if err != nil {
		t.Fatal(err)
	}
	labels := strings.Join(plannedLabels(prepared.Plan()), "\n")
	if strings.Contains(labels, "Install") || strings.Count(labels, "Setup tool") != 1 {
		t.Fatalf("plan = %s", labels)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	shell.mu.Lock()
	calls := append([]string(nil), shell.calls...)
	shell.mu.Unlock()
	if strings.Count(strings.Join(calls, "\n"), "tool detect") != 1 || strings.Count(strings.Join(calls, "\n"), "dep detect") != 1 {
		t.Fatalf("detection calls repeated: %v", calls)
	}
}

func TestPreparedApply_PlanInspectionCannotMutateExecution(t *testing.T) {
	integ := testIntegration("tool")
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[tool detect]":        {output: "installed"},
		"/wt/a:shell[tool setup '/wt/a']": {output: "ok"},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{integ}, nil, []string{"/wt/a"})
	if err != nil {
		t.Fatal(err)
	}

	inspection := prepared.Plan()
	inspection.Phases[0].Label = "mutated phase"
	inspection.Phases[0].Steps[0].Label = "mutated step"

	var events []progress.Event
	if _, err := prepared.Run(shell, func(event progress.Event) { events = append(events, event) }); err != nil {
		t.Fatal(err)
	}
	states := progress.Snapshot(events)
	if len(states) != 1 || states[0].Name != "/wt/a" || states[0].Steps[0].Name != "Setup tool" {
		t.Fatalf("inspection mutated execution: %#v", states)
	}
}

func TestPreparedApply_Empty(t *testing.T) {
	prepared, err := PrepareApply(&applyShell{}, "/repo", "", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !prepared.Empty() {
		t.Fatal("an apply with no frozen operations must be empty")
	}

	prepared, err = PrepareApply(&applyShell{}, "/repo", "/wt/a", nil, []Integration{{
		Name: "tool", Teardown: TeardownSpec{Dirs: []string{".tool/"}},
	}}, []string{"/wt/a"})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Empty() {
		t.Fatal("an apply with a frozen operation must not be empty")
	}
}

func TestPreparedApply_RunSurfacesStartError(t *testing.T) {
	prepared := PreparedApply{plan: progress.Plan{Phases: []progress.PlannedPhase{{Label: "missing ID"}}}}
	if _, err := prepared.Run(&applyShell{}, func(progress.Event) {}); err == nil || !strings.Contains(err.Error(), "empty ID") {
		t.Fatalf("Run error = %v, want invalid-plan error", err)
	}
}

func TestPreparedApply_RunSurfacesTransitionDeliveryError(t *testing.T) {
	phaseID := progress.PhaseID("phase")
	stepID := progress.StepID("declared")
	prepared := PreparedApply{
		plan: progress.Plan{Phases: []progress.PlannedPhase{{
			ID: phaseID, Label: "Phase", Steps: []progress.PlannedStep{{ID: stepID, Label: "Step"}},
		}}},
		operations: []applyOperation{{
			phaseID: phaseID, phaseName: "Phase", stepID: progress.StepID("other"), label: "Step",
			kind: applyFailure, failure: errors.New("failed"),
		}},
	}
	if _, err := prepared.Run(&applyShell{}, func(progress.Event) {}); err == nil || !strings.Contains(err.Error(), "no step ID") {
		t.Fatalf("Run error = %v, want transition error", err)
	}
}

func TestPreparedApply_RunSurfacesFinishDeliveryError(t *testing.T) {
	phaseID := progress.PhaseID("phase")
	stepID := progress.StepID("declared")
	prepared := PreparedApply{plan: progress.Plan{Phases: []progress.PlannedPhase{{
		ID: phaseID, Label: "Phase", Steps: []progress.PlannedStep{{ID: stepID, Label: "Step"}},
	}}}}
	emitErr := errors.New("sink closed")
	_, err := prepared.Run(&applyShell{}, func(event progress.Event) {
		if event.Status == progress.StepSkipped {
			panic(emitErr)
		}
	})
	if !errors.Is(err, emitErr) {
		t.Fatalf("Run error = %v, want %v", err, emitErr)
	}
}

func TestPrepareApply_DisableOnlyWithoutTargetsFails(t *testing.T) {
	_, err := PrepareApply(&applyShell{}, "/repo", "", nil, []Integration{{Name: "tool"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "no target worktree") {
		t.Fatalf("PrepareApply error = %v, want no-target error", err)
	}
}

func TestPreparedApply_DependencyInstallFailureSkipsToolAndSetup(t *testing.T) {
	integ := testIntegration("tool", Dependency{Name: "dep", Detect: "dep detect", Install: "dep install"})
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[tool detect]": {err: errors.New("missing")},
		"/wt/a:shell[dep detect]":  {err: errors.New("missing")},
		"/wt/a:shell[dep install]": {err: errors.New("dep broke")},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{integ}, nil, []string{"/wt/a"})
	if err != nil {
		t.Fatal(err)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	if step := findStep(t, events, "Install dependency dep"); step.Status != progress.StepFailed {
		t.Fatalf("dependency = %#v", step)
	}
	if step := findStep(t, events, "Install tool"); step.Status != progress.StepSkipped || !strings.Contains(step.Message, "blocked by Install dependency dep") {
		t.Fatalf("tool = %#v", step)
	}
	if step := findStep(t, events, "Setup tool"); step.Status != progress.StepSkipped || !strings.Contains(step.Message, "blocked by Install tool") {
		t.Fatalf("setup = %#v", step)
	}
}

func TestPreparedApply_MissingInstallableDependencyRunsPipeline(t *testing.T) {
	integ := testIntegration("tool", Dependency{Name: "dep", Detect: "dep detect", Install: "dep install"})
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[tool detect]":        {err: errors.New("missing")},
		"/wt/a:shell[dep detect]":         {err: errors.New("missing")},
		"/wt/a:shell[dep install]":        {output: "ok"},
		"/wt/a:shell[tool install]":       {output: "ok"},
		"/wt/a:shell[tool setup '/wt/a']": {output: "ok"},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{integ}, nil, []string{"/wt/a"})
	if err != nil {
		t.Fatal(err)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	for _, label := range []string{"Install dependency dep", "Install tool", "Setup tool"} {
		if step := findStep(t, events, label); step.Status != progress.StepDone {
			t.Fatalf("%s = %#v", label, step)
		}
	}
}

func TestPreparedApply_MissingDependencyWithoutInstallerFailsHonestly(t *testing.T) {
	integ := testIntegration("tool", Dependency{Name: "dep", Detect: "dep detect"})
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[tool detect]": {err: errors.New("missing")},
		"/wt/a:shell[dep detect]":  {err: errors.New("missing")},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{integ}, nil, []string{"/wt/a"})
	if err != nil {
		t.Fatal(err)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	dep := findStep(t, events, "Install dependency dep")
	if dep.Status != progress.StepFailed || dep.Error == nil || !strings.Contains(dep.Error.Error(), "no install command") {
		t.Fatalf("dependency = %#v", dep)
	}
}

func TestPreparedApply_SetupFailureStillSettles(t *testing.T) {
	integ := testIntegration("tool")
	shell := &applyShell{responses: map[string]mockShellResponse{
		"/wt/a:shell[tool detect]":        {output: "installed"},
		"/wt/a:shell[tool setup '/wt/a']": {err: errors.New("setup broke")},
	}}
	prepared, err := PrepareApply(shell, "/repo", "/wt/a", []Integration{integ}, nil, []string{"/wt/a"})
	if err != nil {
		t.Fatal(err)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	if step := findStep(t, events, "Setup tool"); step.Status != progress.StepFailed {
		t.Fatalf("setup = %#v", step)
	}
}

func TestPreparedApply_TeardownFailureStillRemovesArtifacts(t *testing.T) {
	wt := t.TempDir()
	artifact := filepath.Join(wt, ".tool")
	if err := os.MkdirAll(artifact, 0o755); err != nil {
		t.Fatal(err)
	}
	integ := Integration{Name: "tool", Teardown: TeardownSpec{Command: "tool clean", Dirs: []string{".tool/"}}}
	shell := &applyShell{responses: map[string]mockShellResponse{wt + ":shell[tool clean]": {err: errors.New("clean broke")}}}
	prepared, err := PrepareApply(shell, "/repo", wt, nil, []Integration{integ}, []string{wt})
	if err != nil {
		t.Fatal(err)
	}
	events, _ := collectPreparedEvents(prepared, shell)
	assertSettledStream(t, events)
	if _, err := os.Stat(artifact); !os.IsNotExist(err) {
		t.Fatalf("artifact still exists: %v", err)
	}
	if step := findStep(t, events, "Teardown tool"); step.Status != progress.StepFailed {
		t.Fatalf("teardown = %#v", step)
	}
	if step := findStep(t, events, RemoveDirStepName(".tool/", wt)); step.Status != progress.StepDone {
		t.Fatalf("removal = %#v", step)
	}
}
