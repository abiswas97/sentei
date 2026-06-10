package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/pipeline"
)

func TestPrintEvent(t *testing.T) {
	tests := []struct {
		name  string
		event cleanup.Event
		want  []string
	}{
		{"step", cleanup.Event{Level: cleanup.LevelStep, Message: "Checking refs"}, []string{"→", "Checking refs"}},
		{"info", cleanup.Event{Level: cleanup.LevelInfo, Message: "Pruned 2 refs"}, []string{"✓", "Pruned 2 refs"}},
		{"warn", cleanup.Event{Level: cleanup.LevelWarn, Message: "skipped branch"}, []string{"⚠", "skipped branch"}},
		{"detail", cleanup.Event{Level: cleanup.LevelDetail, Message: "3 branches remain"}, []string{"3 branches remain"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printEvent(tt.event) })
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("output %q missing %q", out, want)
				}
			}
		})
	}
}

func TestPrintCloneEvent(t *testing.T) {
	tests := []struct {
		name  string
		event pipeline.Event
		want  []string
	}{
		{"running", pipeline.Event{Phase: "clone", Step: "fetch", Status: pipeline.StepRunning}, []string{"→", "[clone]", "fetch"}},
		{"done", pipeline.Event{Phase: "clone", Step: "fetch", Status: pipeline.StepDone}, []string{"✓", "[clone]", "fetch"}},
		{"done with message", pipeline.Event{Phase: "clone", Step: "fetch", Status: pipeline.StepDone, Message: "fast"}, []string{"✓", "(fast)"}},
		{"failed", pipeline.Event{Phase: "clone", Step: "fetch", Status: pipeline.StepFailed, Error: errors.New("boom")}, []string{"✗", "boom"}},
		{"skipped", pipeline.Event{Phase: "clone", Step: "fetch", Status: pipeline.StepSkipped}, []string{"⊘", "fetch"}},
		{"skipped with message", pipeline.Event{Phase: "clone", Step: "fetch", Status: pipeline.StepSkipped, Message: "no-op"}, []string{"⊘", "(no-op)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printCloneEvent(tt.event) })
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("output %q missing %q", out, want)
				}
			}
		})
	}
}

func TestPrintCreateEvent(t *testing.T) {
	tests := []struct {
		name  string
		event pipeline.Event
		want  []string
	}{
		{"running", pipeline.Event{Phase: "create", Step: "branch", Status: pipeline.StepRunning}, []string{"→", "create", "branch"}},
		{"done", pipeline.Event{Step: "branch", Status: pipeline.StepDone}, []string{"✓", "branch"}},
		{"done with message", pipeline.Event{Step: "branch", Status: pipeline.StepDone, Message: "created"}, []string{"✓", "— created"}},
		{"failed", pipeline.Event{Step: "branch", Status: pipeline.StepFailed, Error: errors.New("boom")}, []string{"✗", "— boom"}},
		{"failed without error", pipeline.Event{Step: "branch", Status: pipeline.StepFailed}, []string{"✗", "branch"}},
		{"skipped", pipeline.Event{Step: "branch", Status: pipeline.StepSkipped}, []string{"branch (skipped)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printCreateEvent(tt.event) })
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("output %q missing %q", out, want)
				}
			}
		})
	}
}

func TestPrintMigrateEvent(t *testing.T) {
	tests := []struct {
		name  string
		event pipeline.Event
		want  []string
	}{
		{"running", pipeline.Event{Phase: "backup", Step: "copy", Status: pipeline.StepRunning}, []string{"→", "[backup]", "copy"}},
		{"running with message", pipeline.Event{Phase: "backup", Step: "copy", Status: pipeline.StepRunning, Message: "1.2MB"}, []string{"→", "— 1.2MB"}},
		{"done", pipeline.Event{Phase: "backup", Step: "copy", Status: pipeline.StepDone}, []string{"✓", "[backup]", "copy"}},
		{"done with message", pipeline.Event{Phase: "backup", Step: "copy", Status: pipeline.StepDone, Message: "ok"}, []string{"✓", "(ok)"}},
		{"failed", pipeline.Event{Phase: "backup", Step: "copy", Status: pipeline.StepFailed, Error: errors.New("boom")}, []string{"✗", "boom"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printMigrateEvent(tt.event) })
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("output %q missing %q", out, want)
				}
			}
		})
	}
}
