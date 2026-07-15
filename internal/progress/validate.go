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
	phaseLabels := map[PhaseID]string{}
	stepLabels := map[string]string{}
	prefixStage := 0 // 0: declarations, 1: closes, 2: work
	for i, ev := range events {
		key := ev.Phase + "\x00" + ev.Step
		if ev.PhaseLabel != "" {
			if label := phaseLabels[ev.Phase]; label != "" && label != ev.PhaseLabel {
				return fmt.Errorf("event %d: phase %q label changed from %q to %q", i, ev.Phase, label, ev.PhaseLabel)
			}
			phaseLabels[ev.Phase] = ev.PhaseLabel
		}
		if ev.StepLabel != "" {
			if label := stepLabels[key]; label != "" && label != ev.StepLabel {
				return fmt.Errorf("event %d: phase %q step %q label changed from %q to %q", i, ev.Phase, ev.Step, label, ev.StepLabel)
			}
			stepLabels[key] = ev.StepLabel
		}
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
		case StepDone:
			decl.terminal = true
		case StepFailed:
			if ev.Error == nil {
				return fmt.Errorf("event %d: failed step %q has no error", i, ev.Step)
			}
			decl.terminal = true
		case StepSkipped:
			if ev.Message == "" {
				return fmt.Errorf("event %d: skipped step %q has no reason", i, ev.Step)
			}
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

// ValidateCompletedStream applies the strict stream contract and additionally
// requires a completed execution.
func ValidateCompletedStream(events []Event) error {
	if err := ValidateStream(events); err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	declared := map[string]Event{}
	phaseSteps := map[PhaseID]int{}
	terminalSteps := map[string]bool{}
	for _, ev := range events {
		key := ev.Phase + "\x00" + ev.Step
		switch {
		case ev.Status == StepPending && !ev.Close:
			declared[key] = ev
			phaseSteps[ev.Phase]++
		case !ev.Close && terminal(ev.Status):
			terminalSteps[key] = true
		}
	}
	for key, declaration := range declared {
		if !terminalSteps[key] {
			return fmt.Errorf("phase %q step %q is not terminal", declaration.Phase, declaration.Step)
		}
	}
	for _, ev := range events {
		if ev.Close && phaseSteps[ev.Phase] == 0 {
			return fmt.Errorf("phase %q has no declared steps", ev.Phase)
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
