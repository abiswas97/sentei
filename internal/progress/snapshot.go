package progress

// PhaseState is the folded display state of one phase: its steps in
// first-appearance order and the derived counts.
type PhaseState struct {
	Name   string
	Steps  []StepState
	Total  int
	Done   int
	Failed int
}

// StepState is the folded display state of one step.
type StepState struct {
	Name   string
	Status StepStatus
}

// Snapshot folds an event stream into per-phase display state, preserving
// the order phases and steps first appeared in. Done, Skipped, and Failed
// steps all count as resolved (a phase with a best-effort skip still
// reaches completion); a later event for a step supersedes its status.
func Snapshot(events []Event) []PhaseState {
	phases := map[string]*PhaseState{}
	var order []string

	for _, ev := range events {
		ps, exists := phases[ev.Phase]
		if !exists {
			ps = &PhaseState{Name: ev.Phase}
			phases[ev.Phase] = ps
			order = append(order, ev.Phase)
		}

		found := false
		for i := range ps.Steps {
			if ps.Steps[i].Name == ev.Step {
				ps.Steps[i].Status = ev.Status
				found = true
				break
			}
		}
		if !found {
			ps.Steps = append(ps.Steps, StepState{Name: ev.Step, Status: ev.Status})
		}
	}

	var result []PhaseState
	for _, name := range order {
		ps := phases[name]
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
