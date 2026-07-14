package progress

// PhaseState is the folded display state of one phase: its steps in
// first-appearance order, the derived counts, and whether the phase's step
// set is final (Closed).
type PhaseState struct {
	ID     PhaseID
	Name   string
	Steps  []StepState
	Total  int
	Done   int
	Failed int
	Closed bool
}

// StepState is the folded display state of one step. Declared is the step's
// checkpoint count (1 for atomic steps); Reached is how many checkpoints
// have been crossed, monotonic and clamped to Declared. Resolution (Done,
// Failed, Skipped) reaches the final checkpoint.
type StepState struct {
	ID       StepID
	Name     string
	Status   StepStatus
	Message  string
	Error    error
	Reached  int
	Declared int
}

// Settled reports whether the phase may render done treatment: its step set
// is final and every declared step is resolved. A phase with no steps is
// never settled (it renders pending or skipped, not done).
func (p PhaseState) Settled() bool {
	return p.Closed && p.Total > 0 && p.Done == p.Total
}

// CheckpointProgress sums reached and declared checkpoints across phases,
// the overall bar's fill source. Undeclared steps count as one checkpoint
// reached on resolution, so streams without declarations yield the same
// ratio as step counting.
func CheckpointProgress(states []PhaseState) (reached, declared int) {
	for _, p := range states {
		for _, s := range p.Steps {
			reached += s.Reached
			declared += s.Declared
		}
	}
	return reached, declared
}

// Snapshot folds an event stream into per-phase display state, preserving
// the order phases and steps first appeared in. Done, Skipped, and Failed
// steps all count as resolved (a phase with a best-effort skip still
// reaches completion); a later event for a step supersedes its status.
// Declaration events (Pending bursts, close markers) establish totals,
// checkpoint counts, and the Closed flag; the fold is forgiving toward
// undeclared streams, which keep discovery semantics.
func Snapshot(events []Event) []PhaseState {
	phases := map[PhaseID]*PhaseState{}
	var order []PhaseID

	phaseFor := func(id PhaseID, label string) *PhaseState {
		ps, exists := phases[id]
		if !exists {
			if label == "" {
				label = id
			}
			ps = &PhaseState{ID: id, Name: label}
			phases[id] = ps
			order = append(order, id)
		} else if ps.Name == "" && label != "" {
			ps.Name = label
		}
		return ps
	}

	for _, ev := range events {
		ps := phaseFor(ev.Phase, ev.PhaseLabel)
		if ev.Close {
			ps.Closed = true
			continue
		}

		idx := -1
		for i := range ps.Steps {
			if ps.Steps[i].ID == ev.Step {
				idx = i
				break
			}
		}
		if idx == -1 {
			label := ev.StepLabel
			if label == "" {
				label = ev.Step
			}
			ps.Steps = append(ps.Steps, StepState{ID: ev.Step, Name: label, Declared: 1})
			idx = len(ps.Steps) - 1
		}
		step := &ps.Steps[idx]
		if step.Name == "" && ev.StepLabel != "" {
			step.Name = ev.StepLabel
		}

		step.Status = ev.Status
		if ev.Message != "" {
			step.Message = ev.Message
		}
		if ev.Error != nil {
			step.Error = ev.Error
		}
		step.Declared = max(step.Declared, ev.Of)
		switch ev.Status {
		case StepRunning:
			step.Reached = max(step.Reached, min(ev.Checkpoint, step.Declared))
		case StepDone, StepFailed, StepSkipped:
			step.Reached = step.Declared
		}
	}

	var result []PhaseState
	for _, id := range order {
		ps := phases[id]
		ps.Total = len(ps.Steps)
		for _, s := range ps.Steps {
			switch s.Status {
			case StepDone, StepSkipped:
				ps.Done++
			case StepFailed:
				ps.Failed++
				ps.Done++
			}
		}
		result = append(result, *ps)
	}
	return result
}

// WithPendingPhases returns states reordered onto the canonical phase
// sequence, inserting an empty (pending) PhaseState for any canonical phase
// that has not emitted events yet. Phases outside the canonical list keep
// their discovery order at the end.
func WithPendingPhases(states []PhaseState, names ...string) []PhaseState {
	byName := make(map[string]PhaseState, len(states))
	for _, ps := range states {
		byName[ps.Name] = ps
	}
	result := make([]PhaseState, 0, len(names)+len(states))
	canonical := make(map[string]bool, len(names))
	for _, name := range names {
		canonical[name] = true
		if ps, ok := byName[name]; ok {
			result = append(result, ps)
		} else {
			result = append(result, PhaseState{Name: name})
		}
	}
	for _, ps := range states {
		if !canonical[ps.Name] {
			result = append(result, ps)
		}
	}
	return result
}
