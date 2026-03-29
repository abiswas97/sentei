package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/config"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestRunDeps_SingleEcosystem(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo/feature-auth:shell[pnpm install]": {output: ""},
	}}

	opts := Options{
		Ecosystems: []config.EcosystemConfig{
			{
				Name:    "pnpm",
				Install: config.InstallConfig{Command: "pnpm install"},
			},
		},
	}

	ec := &eventCollector{}
	phase := runDeps(runner, "/repo/feature-auth", opts, ec.emit)

	if phase.Name != "Dependencies" {
		t.Errorf("phase name = %q, want %q", phase.Name, "Dependencies")
	}
	if len(phase.Steps) != 1 {
		t.Fatalf("step count = %d, want 1", len(phase.Steps))
	}
	if phase.Steps[0].Status != StepDone {
		t.Errorf("step status = %v, want StepDone", phase.Steps[0].Status)
	}
}

func TestRunDeps_NoEcosystems(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{}}

	opts := Options{Ecosystems: nil}
	ec := &eventCollector{}
	phase := runDeps(runner, "/repo/feature-auth", opts, ec.emit)

	if len(phase.Steps) != 0 {
		t.Errorf("step count = %d, want 0", len(phase.Steps))
	}
}

func TestRunDeps_InstallFailure(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo/feature-auth:shell[pnpm install]": {err: fmt.Errorf("ENOENT")},
	}}

	opts := Options{
		Ecosystems: []config.EcosystemConfig{
			{
				Name:    "pnpm",
				Install: config.InstallConfig{Command: "pnpm install"},
			},
		},
	}

	ec := &eventCollector{}
	phase := runDeps(runner, "/repo/feature-auth", opts, ec.emit)

	if phase.Steps[0].Status != StepFailed {
		t.Errorf("step status = %v, want StepFailed", phase.Steps[0].Status)
	}
}

func TestRunDeps_ParallelWorkspaces(t *testing.T) {
	tmpDir := t.TempDir()
	wtPath := filepath.Join(tmpDir, "feature-auth")
	os.MkdirAll(wtPath, 0755)

	pkgsUI := filepath.Join(wtPath, "packages", "ui")
	pkgsCore := filepath.Join(wtPath, "packages", "core")
	os.MkdirAll(pkgsUI, 0755)
	os.MkdirAll(pkgsCore, 0755)

	wsYaml := "packages:\n  - packages/*\n"
	os.WriteFile(filepath.Join(wtPath, "pnpm-workspace.yaml"), []byte(wsYaml), 0644)

	runner := &mockRunner{responses: map[string]mockResponse{
		wtPath + ":shell[pnpm install]":                                      {output: ""},
		fmt.Sprintf("%s:shell[pnpm install --filter packages/ui]", wtPath):   {output: ""},
		fmt.Sprintf("%s:shell[pnpm install --filter packages/core]", wtPath): {output: ""},
	}}

	opts := Options{
		Ecosystems: []config.EcosystemConfig{
			{
				Name: "pnpm",
				Install: config.InstallConfig{
					Command:          "pnpm install",
					WorkspaceDetect:  "pnpm-workspace.yaml",
					WorkspaceInstall: "pnpm install --filter {dir}",
					Parallel:         boolPtr(true),
				},
			},
		},
	}

	ec := &eventCollector{}
	phase := runDeps(runner, wtPath, opts, ec.emit)

	// Root install + 2 workspace installs = 3 steps
	if len(phase.Steps) != 3 {
		t.Fatalf("step count = %d, want 3", len(phase.Steps))
	}

	// Verify all steps completed
	for i, step := range phase.Steps {
		if step.Status != StepDone {
			t.Errorf("step[%d] %q status = %v, want StepDone", i, step.Name, step.Status)
		}
	}

	// Verify events contain "running" and "done" for each
	runningCount := 0
	for _, e := range ec.events {
		if e.Status == StepRunning {
			runningCount++
		}
	}
	if runningCount < 3 {
		t.Errorf("expected at least 3 running events, got %d", runningCount)
	}
}

func TestRunDeps_CommandParsing(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/wt:shell[go mod download]": {output: ""},
	}}

	opts := Options{
		Ecosystems: []config.EcosystemConfig{
			{
				Name:    "go",
				Install: config.InstallConfig{Command: "go mod download"},
			},
		},
	}

	ec := &eventCollector{}
	runDeps(runner, "/wt", opts, ec.emit)

	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.calls))
	}
	if !strings.Contains(runner.calls[0], "shell[go mod download]") {
		t.Errorf("call = %q, expected to contain 'shell[go mod download]'", runner.calls[0])
	}
}
