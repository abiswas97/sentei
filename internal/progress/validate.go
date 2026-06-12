package progress

import "fmt"

// ValidateStream checks the honesty invariants a well-formed stream must
// hold. Close marks a phase's step set as final, so work events for already
// declared steps legitimately follow it; the violation is an event that
// introduces a previously unseen step after the close. The fold itself
// stays forgiving (production never panics on a misbehaving emitter);
// tests call this to flag violations.
func ValidateStream(events []Event) error {
	closed := map[string]bool{}
	seen := map[string]bool{}
	reached := map[string]int{}
	for i, ev := range events {
		if ev.Close {
			closed[ev.Phase] = true
			continue
		}
		key := ev.Phase + "\x00" + ev.Step
		if closed[ev.Phase] && !seen[key] {
			return fmt.Errorf("event %d: step %q added after phase %q closed", i, ev.Step, ev.Phase)
		}
		seen[key] = true
		if ev.Status == StepRunning && ev.Checkpoint > 0 {
			if ev.Checkpoint < reached[key] {
				return fmt.Errorf("event %d: step %q checkpoint regressed from %d to %d", i, ev.Step, reached[key], ev.Checkpoint)
			}
			reached[key] = ev.Checkpoint
		}
	}
	return nil
}
