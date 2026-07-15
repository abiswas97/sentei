package progress

// Plan declares a flow's work upfront: phases, their steps, and each step's
// checkpoint count. Start validates and owns execution of this plan.
type Plan struct {
	Phases []PlannedPhase
}

// Clone returns an independently mutable copy of the plan.
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

// PlannedPhase declares one phase.
type PlannedPhase struct {
	ID    PhaseID
	Label string
	Steps []PlannedStep
}

// PlannedStep declares one step. Checkpoints below 1 declare an atomic step
// (one checkpoint: its resolution).
type PlannedStep struct {
	ID          StepID
	Label       string
	Checkpoints int
}
