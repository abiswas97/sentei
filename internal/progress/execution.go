package progress

import (
	"fmt"
	"sync"
)

// Execution owns the mutable state of a validated Plan. All transitions and
// emissions are serialized; Run releases the lock while user work executes.
type Execution struct {
	mu       sync.Mutex
	emit     func(Event)
	phases   map[PhaseID]*executionPhase
	order    []PhaseID
	pending  []Event
	emitting bool
}

type executionPhase struct {
	id    PhaseID
	label string
	steps map[StepID]*executionStep
	order []StepID
}

type executionStep struct {
	id          StepID
	label       string
	checkpoints int
	checkpoint  int
	status      StepStatus
	message     string
	err         error
}

// Start validates a plan, copies it into execution state, and emits the whole
// declaration prefix before any phase-close markers.
func Start(plan Plan, emit func(Event)) (*Execution, error) {
	if emit == nil {
		emit = func(Event) {}
	}
	x := &Execution{
		emit:   emit,
		phases: make(map[PhaseID]*executionPhase, len(plan.Phases)),
		order:  make([]PhaseID, 0, len(plan.Phases)),
	}
	for phaseIndex, plannedPhase := range plan.Phases {
		if plannedPhase.Name != "" || plannedPhase.Open {
			return nil, fmt.Errorf("phase %d mixes stable and legacy plan fields", phaseIndex)
		}
		if plannedPhase.ID == "" {
			return nil, fmt.Errorf("phase %d has empty ID", phaseIndex)
		}
		if _, exists := x.phases[plannedPhase.ID]; exists {
			return nil, fmt.Errorf("duplicate phase ID %q", plannedPhase.ID)
		}
		label := plannedPhase.Label
		if label == "" {
			label = plannedPhase.ID
		}
		phase := &executionPhase{
			id:    plannedPhase.ID,
			label: label,
			steps: make(map[StepID]*executionStep, len(plannedPhase.Steps)),
			order: make([]StepID, 0, len(plannedPhase.Steps)),
		}
		for stepIndex, plannedStep := range plannedPhase.Steps {
			if plannedStep.Name != "" {
				return nil, fmt.Errorf("phase %q step %d mixes stable and legacy plan fields", plannedPhase.ID, stepIndex)
			}
			if plannedStep.ID == "" {
				return nil, fmt.Errorf("phase %q step %d has empty ID", plannedPhase.ID, stepIndex)
			}
			if _, exists := phase.steps[plannedStep.ID]; exists {
				return nil, fmt.Errorf("phase %q has duplicate step ID %q", plannedPhase.ID, plannedStep.ID)
			}
			stepLabel := plannedStep.Label
			if stepLabel == "" {
				stepLabel = plannedStep.ID
			}
			phase.steps[plannedStep.ID] = &executionStep{
				id:          plannedStep.ID,
				label:       stepLabel,
				checkpoints: max(plannedStep.Checkpoints, 1),
				status:      StepPending,
			}
			phase.order = append(phase.order, plannedStep.ID)
		}
		x.phases[plannedPhase.ID] = phase
		x.order = append(x.order, plannedPhase.ID)
	}

	for _, phaseID := range x.order {
		phase := x.phases[phaseID]
		for _, stepID := range phase.order {
			step := phase.steps[stepID]
			x.emit(Event{
				Phase: phase.id, PhaseLabel: phase.label,
				Step: step.id, StepLabel: step.label,
				Status: StepPending, Of: step.checkpoints,
			})
		}
	}
	for _, phaseID := range x.order {
		phase := x.phases[phaseID]
		x.emit(Event{Phase: phase.id, PhaseLabel: phase.label, Close: true})
	}
	return x, nil
}

// Running marks a step active and optionally advances its checkpoint.
func (x *Execution) Running(phaseID PhaseID, stepID StepID, checkpoint int, message string) error {
	x.mu.Lock()
	phase, step, err := x.step(phaseID, stepID)
	if err != nil {
		x.mu.Unlock()
		return err
	}
	if terminal(step.status) {
		x.mu.Unlock()
		return fmt.Errorf("phase %q step %q is already terminal", phaseID, stepID)
	}
	if checkpoint < step.checkpoint {
		x.mu.Unlock()
		return fmt.Errorf("phase %q step %q checkpoint regressed from %d to %d", phaseID, stepID, step.checkpoint, checkpoint)
	}
	if checkpoint > step.checkpoints {
		x.mu.Unlock()
		return fmt.Errorf("phase %q step %q checkpoint %d exceeds declared %d", phaseID, stepID, checkpoint, step.checkpoints)
	}
	step.status = StepRunning
	step.checkpoint = checkpoint
	if message != "" {
		step.message = message
	}
	drain := x.queueLocked(Event{
		Phase: phase.id, PhaseLabel: phase.label,
		Step: step.id, StepLabel: step.label,
		Status: StepRunning, Checkpoint: checkpoint, Of: step.checkpoints, Message: message,
	})
	x.mu.Unlock()
	if drain {
		x.drain()
	}
	return nil
}

// Done resolves a declared step successfully.
func (x *Execution) Done(phaseID PhaseID, stepID StepID, message string) (StepResult, error) {
	return x.resolve(phaseID, stepID, StepDone, message, nil)
}

// Fail resolves a declared step unsuccessfully.
func (x *Execution) Fail(phaseID PhaseID, stepID StepID, err error) (StepResult, error) {
	return x.resolve(phaseID, stepID, StepFailed, "", err)
}

// Skip resolves a declared step without running it.
func (x *Execution) Skip(phaseID PhaseID, stepID StepID, reason string) (StepResult, error) {
	return x.resolve(phaseID, stepID, StepSkipped, reason, nil)
}

// Run emits the standard running and terminal transitions. fn executes
// outside the mutex so independent steps can progress concurrently.
func (x *Execution) Run(phaseID PhaseID, stepID StepID, fn StepFunc) (StepResult, error) {
	if fn == nil {
		return StepResult{}, fmt.Errorf("phase %q step %q has nil function", phaseID, stepID)
	}
	if err := x.claim(phaseID, stepID); err != nil {
		return StepResult{}, err
	}
	message, err := fn()
	if err != nil {
		return x.Fail(phaseID, stepID, err)
	}
	return x.Done(phaseID, stepID, message)
}

func (x *Execution) claim(phaseID PhaseID, stepID StepID) error {
	x.mu.Lock()
	phase, step, err := x.step(phaseID, stepID)
	if err != nil {
		x.mu.Unlock()
		return err
	}
	if step.status != StepPending {
		x.mu.Unlock()
		return fmt.Errorf("phase %q step %q cannot be claimed from status %d", phaseID, stepID, step.status)
	}
	step.status = StepRunning
	drain := x.queueLocked(Event{
		Phase: phase.id, PhaseLabel: phase.label,
		Step: step.id, StepLabel: step.label,
		Status: StepRunning, Of: step.checkpoints,
	})
	x.mu.Unlock()
	if drain {
		x.drain()
	}
	return nil
}

// SkipPending skips only untouched steps in one phase.
func (x *Execution) SkipPending(phaseID PhaseID, reason string) error {
	x.mu.Lock()
	phase, exists := x.phases[phaseID]
	if !exists {
		x.mu.Unlock()
		return fmt.Errorf("unknown phase ID %q", phaseID)
	}
	drain := false
	for _, stepID := range phase.order {
		step := phase.steps[stepID]
		if step.status == StepPending {
			_, queuedDrain := x.resolveLocked(phase, step, StepSkipped, reason, nil)
			drain = drain || queuedDrain
		}
	}
	x.mu.Unlock()
	if drain {
		x.drain()
	}
	return nil
}

// Finish is the terminal safety net: every unresolved step becomes skipped.
func (x *Execution) Finish(reason string) error {
	x.mu.Lock()
	drain := false
	for _, phaseID := range x.order {
		phase := x.phases[phaseID]
		for _, stepID := range phase.order {
			step := phase.steps[stepID]
			if !terminal(step.status) {
				_, queuedDrain := x.resolveLocked(phase, step, StepSkipped, reason, nil)
				drain = drain || queuedDrain
			}
		}
	}
	x.mu.Unlock()
	if drain {
		x.drain()
	}
	return nil
}

func (x *Execution) resolve(phaseID PhaseID, stepID StepID, status StepStatus, message string, stepErr error) (StepResult, error) {
	x.mu.Lock()
	phase, step, err := x.step(phaseID, stepID)
	if err != nil {
		x.mu.Unlock()
		return StepResult{}, err
	}
	if terminal(step.status) {
		x.mu.Unlock()
		return StepResult{}, fmt.Errorf("phase %q step %q is already terminal", phaseID, stepID)
	}
	result, drain := x.resolveLocked(phase, step, status, message, stepErr)
	x.mu.Unlock()
	if drain {
		x.drain()
	}
	return result, nil
}

func (x *Execution) resolveLocked(phase *executionPhase, step *executionStep, status StepStatus, message string, stepErr error) (StepResult, bool) {
	step.status = status
	step.checkpoint = step.checkpoints
	step.message = message
	step.err = stepErr
	drain := x.queueLocked(Event{
		Phase: phase.id, PhaseLabel: phase.label,
		Step: step.id, StepLabel: step.label,
		Status: status, Of: step.checkpoints, Message: message, Error: stepErr,
	})
	return StepResult{ID: step.id, Name: step.label, Status: status, Message: message, Error: stepErr}, drain
}

func (x *Execution) queueLocked(event Event) bool {
	x.pending = append(x.pending, event)
	if x.emitting {
		return false
	}
	x.emitting = true
	return true
}

func (x *Execution) drain() {
	for {
		x.mu.Lock()
		if len(x.pending) == 0 {
			x.emitting = false
			x.mu.Unlock()
			return
		}
		event := x.pending[0]
		x.pending = x.pending[1:]
		x.mu.Unlock()
		x.emit(event)
	}
}

func (x *Execution) step(phaseID PhaseID, stepID StepID) (*executionPhase, *executionStep, error) {
	phase, exists := x.phases[phaseID]
	if !exists {
		return nil, nil, fmt.Errorf("unknown phase ID %q", phaseID)
	}
	step, exists := phase.steps[stepID]
	if !exists {
		return nil, nil, fmt.Errorf("phase %q has no step ID %q", phaseID, stepID)
	}
	return phase, step, nil
}

func terminal(status StepStatus) bool {
	return status == StepDone || status == StepFailed || status == StepSkipped
}
