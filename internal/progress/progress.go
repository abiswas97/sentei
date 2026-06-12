// Package progress is the single vocabulary for multi-phase operation
// progress (worktree removal, integration apply, teardown, worktree
// creation, repo create/clone/migrate, cleanup): step statuses, per-step
// results grouped into phases, the events emitted while a phase runs, and
// the fold from an event stream to per-phase display state.
package progress

type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepDone
	StepFailed
	StepSkipped
)

// StepResult is the recorded outcome of a single step within a phase.
type StepResult struct {
	Name    string
	Status  StepStatus
	Message string
	Error   error
}

// Phase is a named group of step results.
type Phase struct {
	Name  string
	Steps []StepResult
}

// Event is a progress notification emitted while a flow runs.
type Event struct {
	Phase   string
	Step    string
	Status  StepStatus
	Message string
	Error   error
}

// HasFailures reports whether any step in the phase failed.
func (p *Phase) HasFailures() bool {
	for _, s := range p.Steps {
		if s.Status == StepFailed {
			return true
		}
	}
	return false
}

// PhasesHaveFailures reports whether any step across the phases failed.
func PhasesHaveFailures(phases []Phase) bool {
	for i := range phases {
		if phases[i].HasFailures() {
			return true
		}
	}
	return false
}

// FirstFailure returns the first failed step across the phases, along with
// the name of the phase it belongs to.
func FirstFailure(phases []Phase) (phaseName string, step StepResult, ok bool) {
	for _, p := range phases {
		for _, s := range p.Steps {
			if s.Status == StepFailed {
				return p.Name, s, true
			}
		}
	}
	return "", StepResult{}, false
}
