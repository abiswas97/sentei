package progress

// Plan declares a flow's work upfront: phases, their steps, and each step's
// checkpoint count. Declare compiles it into the event stream, so the stream
// stays the single source of truth for totals.
type Plan struct {
	Phases []PlannedPhase
}

// PlannedPhase declares one phase. Open phases may append steps after
// declaration (scan-style discovery) and must be closed explicitly via
// ClosePhase; all other phases are closed by Declare itself.
type PlannedPhase struct {
	Name  string
	Steps []PlannedStep
	Open  bool
}

// PlannedStep declares one step. Checkpoints below 1 declare an atomic step
// (one checkpoint: its resolution).
type PlannedStep struct {
	Name        string
	Checkpoints int
}

// Declare compiles the plan into the stream: one Pending event per planned
// step carrying its checkpoint count, then a close marker for every phase
// not marked Open.
func Declare(plan Plan, emit func(Event)) {
	for _, phase := range plan.Phases {
		for _, step := range phase.Steps {
			emit(Event{Phase: phase.Name, Step: step.Name, Status: StepPending, Of: max(step.Checkpoints, 1)})
		}
		if !phase.Open {
			emit(Event{Phase: phase.Name, Close: true})
		}
	}
}

// ClosePhase emits the close marker for an Open phase once its step set is
// final.
func ClosePhase(name string, emit func(Event)) {
	emit(Event{Phase: name, Close: true})
}
