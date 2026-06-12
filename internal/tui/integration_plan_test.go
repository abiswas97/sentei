package tui

import (
	"errors"
	"testing"

	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/progress"
)

// Simulates the worktree-outer apply event stream for a 2x2 apply (two
// staged integrations across two worktrees) and asserts the spec scenario
// "Apply phases never reopen": each worktree phase declares its full total
// upfront, progresses monotonically, and settles exactly once.
func TestBuildIntegrationPhases_PhasesNeverReopen(t *testing.T) {
	integA := integration.Integration{Name: "alpha"}
	integB := integration.Integration{Name: "beta"}
	worktrees := []string{"/wt/feat-1", "/wt/feat-2"}

	permutations := []struct {
		name string
		emit func(wt string, integ integration.Integration, send func(progress.Event))
	}{
		{"success", func(wt string, integ integration.Integration, send func(progress.Event)) {
			send(progress.Event{Phase: wt, Step: integration.SetupStepName(integ), Status: progress.StepRunning})
			send(progress.Event{Phase: wt, Step: integration.SetupStepName(integ), Status: progress.StepDone})
		}},
		{"failure", func(wt string, integ integration.Integration, send func(progress.Event)) {
			send(progress.Event{Phase: wt, Step: integration.SetupStepName(integ), Status: progress.StepRunning})
			send(progress.Event{Phase: wt, Step: integration.SetupStepName(integ), Status: progress.StepFailed, Error: errors.New("boom")})
		}},
		{"skip-install", func(wt string, integ integration.Integration, send func(progress.Event)) {
			send(progress.Event{Phase: wt, Step: integration.InstallStepName(integ), Status: progress.StepSkipped})
			send(progress.Event{Phase: wt, Step: integration.SetupStepName(integ), Status: progress.StepRunning})
			send(progress.Event{Phase: wt, Step: integration.SetupStepName(integ), Status: progress.StepDone})
		}},
	}

	for _, perm := range permutations {
		t.Run(perm.name, func(t *testing.T) {
			var stream []progress.Event
			send := func(e progress.Event) { stream = append(stream, e) }

			progress.Declare(integration.ApplyPlan([]integration.Integration{integA, integB}, nil, worktrees), send)
			for _, wt := range worktrees {
				perm.emit(wt, integA, send)
				perm.emit(wt, integB, send)
				progress.ClosePhase(wt, send)
			}
			if err := progress.ValidateStream(stream); err != nil {
				t.Fatalf("driver emitted an invalid stream: %v", err)
			}

			declared := 0
			for _, e := range stream {
				if e.Status == progress.StepPending && !e.Close {
					declared++
				}
			}

			m := NewModel(nil, nil, "/repo")
			settledAt := map[string]int{}
			maxTotal := map[string]int{}
			for i := 1; i <= len(stream); i++ {
				m.integ.events = stream[:i]
				phases := m.buildIntegrationPhases()
				for _, p := range phases {
					if p.Total < maxTotal[p.Name] {
						t.Fatalf("prefix %d: phase %s total regressed %d -> %d", i, p.Name, maxTotal[p.Name], p.Total)
					}
					maxTotal[p.Name] = p.Total
					if i > declared && p.Total != 2 {
						// Declaration burst fully folded: every phase must
						// carry its full upfront total from then on.
						t.Fatalf("prefix %d: phase %s total = %d, want 2 declared upfront", i, p.Name, p.Total)
					}
					if p.Settled() {
						if settledAt[p.Name] == 0 {
							settledAt[p.Name] = i
						}
					} else if settledAt[p.Name] != 0 {
						t.Fatalf("prefix %d: phase %s reopened after settling at prefix %d", i, p.Name, settledAt[p.Name])
					}
				}
			}
			if len(settledAt) != 2 {
				t.Fatalf("expected both phases to settle, got %v", settledAt)
			}
		})
	}
}
