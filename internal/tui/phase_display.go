package tui

import "github.com/abiswas97/sentei/internal/progress"

type phaseDisplay struct {
	name   string
	steps  []stepDisplay
	total  int
	done   int
	failed int
}

type stepDisplay struct {
	name   string
	status progress.StepStatus
}

// buildPhaseDisplays folds a pipeline event stream into per-phase display
// state, preserving the order phases first appeared in.
func buildPhaseDisplays(events []progress.Event) []phaseDisplay {
	phases := map[string]*phaseDisplay{}
	var order []string

	for _, ev := range events {
		pd, exists := phases[ev.Phase]
		if !exists {
			pd = &phaseDisplay{name: ev.Phase}
			phases[ev.Phase] = pd
			order = append(order, ev.Phase)
		}

		found := false
		for i := range pd.steps {
			if pd.steps[i].name == ev.Step {
				pd.steps[i].status = ev.Status
				found = true
				break
			}
		}
		if !found {
			pd.steps = append(pd.steps, stepDisplay{name: ev.Step, status: ev.Status})
		}
	}

	var result []phaseDisplay
	for _, name := range order {
		pd := phases[name]
		pd.total = len(pd.steps)
		for _, s := range pd.steps {
			switch s.status {
			case progress.StepDone, progress.StepSkipped:
				// A skipped step is resolved (non-failing); count it as done so a
				// phase with a best-effort skip still reaches 100%.
				pd.done++
			case progress.StepFailed:
				pd.failed++
				pd.done++
			}
		}
		result = append(result, *pd)
	}
	return result
}

// withPendingPhases returns displays reordered onto the canonical phase
// sequence, inserting an empty (pending) phaseDisplay for any canonical
// phase that has not emitted events yet. Phases outside the canonical list
// keep their discovery order at the end.
func withPendingPhases(displays []phaseDisplay, names ...string) []phaseDisplay {
	byName := make(map[string]phaseDisplay, len(displays))
	for _, pd := range displays {
		byName[pd.name] = pd
	}
	result := make([]phaseDisplay, 0, len(names)+len(displays))
	canonical := make(map[string]bool, len(names))
	for _, name := range names {
		canonical[name] = true
		if pd, ok := byName[name]; ok {
			result = append(result, pd)
		} else {
			result = append(result, phaseDisplay{name: name})
		}
	}
	for _, pd := range displays {
		if !canonical[pd.name] {
			result = append(result, pd)
		}
	}
	return result
}
