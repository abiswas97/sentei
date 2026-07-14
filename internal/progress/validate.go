package progress

import "fmt"

// ValidateStream checks the stable stream contract: a complete
// declaration-and-close prefix must precede work, and every transition must
// target declared identities without mutating terminal state.
func ValidateStream(events []Event) error {
	type declaration struct {
		checkpoints int
		terminal    bool
		reached     int
	}
	declarations := map[string]*declaration{}
	declaredPhases := map[PhaseID]bool{}
	closed := map[PhaseID]bool{}
	prefixStage := 0 // 0: declarations, 1: closes, 2: work
	for i, ev := range events {
		key := ev.Phase + "\x00" + ev.Step
		if ev.Status == StepPending && !ev.Close {
			if prefixStage != 0 {
				return fmt.Errorf("event %d: declaration for step %q appears after declaration prefix", i, ev.Step)
			}
			if ev.Phase == "" || ev.Step == "" {
				return fmt.Errorf("event %d: declaration has empty phase or step ID", i)
			}
			if ev.Of < 1 {
				return fmt.Errorf("event %d: step %q declares invalid checkpoint count %d", i, ev.Step, ev.Of)
			}
			if _, exists := declarations[key]; exists {
				return fmt.Errorf("event %d: duplicate declaration for phase %q step %q", i, ev.Phase, ev.Step)
			}
			declaredPhases[ev.Phase] = true
			declarations[key] = &declaration{checkpoints: ev.Of}
			continue
		}
		if ev.Close {
			if prefixStage == 2 {
				return fmt.Errorf("event %d: phase %q closes after work began", i, ev.Phase)
			}
			if ev.Phase == "" {
				return fmt.Errorf("event %d: close marker has empty phase ID", i)
			}
			prefixStage = 1
			declaredPhases[ev.Phase] = true
			if closed[ev.Phase] {
				return fmt.Errorf("event %d: duplicate close for phase %q", i, ev.Phase)
			}
			closed[ev.Phase] = true
			continue
		}
		prefixStage = 2
		for phaseID := range declaredPhases {
			if !closed[phaseID] {
				return fmt.Errorf("event %d: work began before phase %q closed", i, phaseID)
			}
		}
		decl, exists := declarations[key]
		if !exists {
			return fmt.Errorf("event %d: undeclared phase %q step %q", i, ev.Phase, ev.Step)
		}
		if decl.terminal {
			return fmt.Errorf("event %d: terminal mutation for phase %q step %q", i, ev.Phase, ev.Step)
		}
		if ev.Of != 0 && ev.Of != decl.checkpoints {
			return fmt.Errorf("event %d: step %q checkpoint total changed from %d to %d", i, ev.Step, decl.checkpoints, ev.Of)
		}
		switch ev.Status {
		case StepRunning:
			if ev.Checkpoint < decl.reached {
				return fmt.Errorf("event %d: step %q checkpoint regressed from %d to %d", i, ev.Step, decl.reached, ev.Checkpoint)
			}
			if ev.Checkpoint > decl.checkpoints {
				return fmt.Errorf("event %d: step %q checkpoint %d exceeds declared %d", i, ev.Step, ev.Checkpoint, decl.checkpoints)
			}
			decl.reached = ev.Checkpoint
		case StepDone, StepFailed, StepSkipped:
			decl.terminal = true
		case StepPending:
			return fmt.Errorf("event %d: pending status outside declaration prefix", i)
		default:
			return fmt.Errorf("event %d: invalid step status %d", i, ev.Status)
		}
	}
	for phaseID := range declaredPhases {
		if !closed[phaseID] {
			return fmt.Errorf("declaration prefix: phase %q was not closed", phaseID)
		}
	}
	return nil
}

// ValidateLegacyStream preserves discovery semantics for unconverted
// producers. New code must use ValidateStream.
func ValidateLegacyStream(events []Event) error {
	closed := map[PhaseID]bool{}
	seen := map[string]bool{}
	reached := map[string]int{}
	terminalSteps := map[string]bool{}
	for i, ev := range events {
		if ev.Close {
			closed[ev.Phase] = true
			continue
		}
		key := ev.Phase + "\x00" + ev.Step
		if closed[ev.Phase] && !seen[key] {
			return fmt.Errorf("event %d: step %q added after phase %q closed", i, ev.Step, ev.Phase)
		}
		if terminalSteps[key] {
			return fmt.Errorf("event %d: terminal mutation for phase %q step %q", i, ev.Phase, ev.Step)
		}
		seen[key] = true
		if ev.Status == StepRunning && ev.Checkpoint > 0 {
			if ev.Checkpoint < reached[key] {
				return fmt.Errorf("event %d: step %q checkpoint regressed from %d to %d", i, ev.Step, reached[key], ev.Checkpoint)
			}
			if ev.Of > 0 && ev.Checkpoint > ev.Of {
				return fmt.Errorf("event %d: step %q checkpoint %d exceeds declared %d", i, ev.Step, ev.Checkpoint, ev.Of)
			}
			reached[key] = ev.Checkpoint
		}
		if terminal(ev.Status) {
			terminalSteps[key] = true
		}
	}
	return nil
}
