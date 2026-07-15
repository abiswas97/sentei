package progress

// Plan declares a flow's work upfront: phases, their steps, and each step's
// checkpoint count. Start validates and owns execution of this plan.
type Plan struct {
	Phases []PlannedPhase
}

// Clone returns an independently mutable copy of the plan, including legacy
// compatibility fields retained for producers that have not migrated yet.
func (p Plan) Clone() Plan {
	var clone Plan
	if p.Phases != nil {
		clone.Phases = make([]PlannedPhase, len(p.Phases))
	}
	for i := range p.Phases {
		clone.Phases[i] = p.Phases[i]
		if p.Phases[i].Steps != nil {
			clone.Phases[i].Steps = append([]PlannedStep{}, p.Phases[i].Steps...)
		}
	}
	return clone
}

// PlannedPhase declares one phase. Open phases may append steps after
// declaration (scan-style discovery) and must be closed explicitly via
// ClosePhase; all other phases are closed by Declare itself.
type PlannedPhase struct {
	ID    PhaseID
	Label string
	Steps []PlannedStep

	// Name and Open are temporary compatibility fields for producers that have
	// not moved to Execution. Start deliberately does not accept them as IDs.
	Name string
	Open bool
}

// PlannedStep declares one step. Checkpoints below 1 declare an atomic step
// (one checkpoint: its resolution).
type PlannedStep struct {
	ID          StepID
	Label       string
	Name        string
	Checkpoints int
}

// Declare is a temporary adapter for unconverted producers. New code uses
// Start, which emits stable labels and enforces the complete-prefix contract.
func Declare(plan Plan, emit func(Event)) {
	for _, phase := range plan.Phases {
		phaseID := phase.ID
		if phaseID == "" {
			phaseID = phase.Name
		}
		for _, step := range phase.Steps {
			stepID := step.ID
			if stepID == "" {
				stepID = step.Name
			}
			emit(Event{Phase: phaseID, Step: stepID, Status: StepPending, Of: max(step.Checkpoints, 1)})
		}
		if !phase.Open {
			emit(Event{Phase: phaseID, Close: true})
		}
	}
}

// ClosePhase emits the close marker for an Open phase once its step set is
// final.
func ClosePhase(name string, emit func(Event)) {
	emit(Event{Phase: name, Close: true})
}
