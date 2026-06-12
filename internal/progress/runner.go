package progress

// StepFunc does the work of one step. The returned message becomes the Done
// result's Message; a non-nil error fails the step.
type StepFunc func() (message string, err error)

// RunStep executes one step with the standard transitions: a Running event,
// then a Done or Failed event mirrored in the returned StepResult.
func RunStep(phase, step string, emit func(Event), fn StepFunc) StepResult {
	emit(Event{Phase: phase, Step: step, Status: StepRunning})
	msg, err := fn()
	if err != nil {
		emit(Event{Phase: phase, Step: step, Status: StepFailed, Error: err})
		return StepResult{Name: step, Status: StepFailed, Error: err}
	}
	emit(Event{Phase: phase, Step: step, Status: StepDone, Message: msg})
	return StepResult{Name: step, Status: StepDone, Message: msg}
}

// PhaseRecorder builds a Phase step by step for sequential flows, mirroring
// each recorded transition to the emit callback.
type PhaseRecorder struct {
	phase Phase
	emit  func(Event)
}

func NewPhaseRecorder(name string, emit func(Event)) *PhaseRecorder {
	return &PhaseRecorder{phase: Phase{Name: name}, emit: emit}
}

// Step runs fn with the standard transitions, records the result, and
// reports success so callers can early-return on failure.
func (r *PhaseRecorder) Step(name string, fn StepFunc) bool {
	result := RunStep(r.phase.Name, name, r.emit, fn)
	r.phase.Steps = append(r.phase.Steps, result)
	return result.Status != StepFailed
}

// Done records a successful step outside the standard Step flow.
func (r *PhaseRecorder) Done(name, message string) {
	r.phase.Steps = append(r.phase.Steps, StepResult{Name: name, Status: StepDone, Message: message})
	r.emit(Event{Phase: r.phase.Name, Step: name, Status: StepDone, Message: message})
}

// Fail records a failed step outside the standard Step flow (e.g. a
// precondition that fails before the step starts running).
func (r *PhaseRecorder) Fail(name string, err error) {
	r.phase.Steps = append(r.phase.Steps, StepResult{Name: name, Status: StepFailed, Error: err})
	r.emit(Event{Phase: r.phase.Name, Step: name, Status: StepFailed, Error: err})
}

// Skip records a non-failing skipped step (e.g. best-effort work that could
// not run).
func (r *PhaseRecorder) Skip(name, message string) {
	r.phase.Steps = append(r.phase.Steps, StepResult{Name: name, Status: StepSkipped, Message: message})
	r.emit(Event{Phase: r.phase.Name, Step: name, Status: StepSkipped, Message: message})
}

// Record appends an externally built result without emitting, for steps with
// non-standard event flows.
func (r *PhaseRecorder) Record(result StepResult) {
	r.phase.Steps = append(r.phase.Steps, result)
}

// Emit forwards an event for the recorder's phase without recording a step,
// for intermediate progress updates within a step.
func (r *PhaseRecorder) Emit(step string, status StepStatus, message string) {
	r.emit(Event{Phase: r.phase.Name, Step: step, Status: status, Message: message})
}

// DeclareSteps emits the Pending burst for steps the phase knows it will
// certainly run, so totals are real before work starts.
func (r *PhaseRecorder) DeclareSteps(names ...string) {
	for _, name := range names {
		r.emit(Event{Phase: r.phase.Name, Step: name, Status: StepPending, Of: 1})
	}
}

// Close emits the phase-close marker: the phase's step set is final.
func (r *PhaseRecorder) Close() {
	r.emit(Event{Phase: r.phase.Name, Close: true})
}

func (r *PhaseRecorder) Phase() Phase { return r.phase }
