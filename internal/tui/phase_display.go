package tui

import "github.com/abiswas97/sentei/internal/pipeline"

type phaseDisplay struct {
	name   string
	steps  []stepDisplay
	total  int
	done   int
	failed int
}

type stepDisplay struct {
	name   string
	status pipeline.StepStatus
}

// buildPhaseDisplays folds a pipeline event stream into per-phase display
// state, preserving the order phases first appeared in.
func buildPhaseDisplays(events []pipeline.Event) []phaseDisplay {
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
			case pipeline.StepDone, pipeline.StepSkipped:
				// A skipped step is resolved (non-failing); count it as done so a
				// phase with a best-effort skip still reaches 100%.
				pd.done++
			case pipeline.StepFailed:
				pd.failed++
				pd.done++
			}
		}
		result = append(result, *pd)
	}
	return result
}
