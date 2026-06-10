package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/pipeline"
	"github.com/abiswas97/sentei/internal/testutil/mock"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestRunDeps_SingleEcosystem(t *testing.T) {
	runner := &mock.Runner{Responses: map[string]mock.Response{
		"/repo/feature-auth:shell[pnpm install]": {Output: ""},
	}}

	opts := Options{
		Ecosystems: []config.EcosystemConfig{
			{
				Name:    "pnpm",
				Install: config.InstallConfig{Command: "pnpm install"},
			},
		},
	}

	ec := &mock.EventCollector[pipeline.Event]{}
	phase := runDeps(runner, "/repo/feature-auth", opts, ec.Emit)

	if phase.Name != "Dependencies" {
		t.Errorf("phase name = %q, want %q", phase.Name, "Dependencies")
	}
	if len(phase.Steps) != 1 {
		t.Fatalf("step count = %d, want 1", len(phase.Steps))
	}
	if phase.Steps[0].Status != pipeline.StepDone {
		t.Errorf("step status = %v, want pipeline.StepDone", phase.Steps[0].Status)
	}
}

func TestRunDeps_NoEcosystems(t *testing.T) {
	runner := &mock.Runner{Responses: map[string]mock.Response{}}

	opts := Options{Ecosystems: nil}
	ec := &mock.EventCollector[pipeline.Event]{}
	phase := runDeps(runner, "/repo/feature-auth", opts, ec.Emit)

	if len(phase.Steps) != 0 {
		t.Errorf("step count = %d, want 0", len(phase.Steps))
	}
}

func TestRunDeps_InstallFailure(t *testing.T) {
	runner := &mock.Runner{Responses: map[string]mock.Response{
		"/repo/feature-auth:shell[pnpm install]": {Err: fmt.Errorf("ENOENT")},
	}}

	opts := Options{
		Ecosystems: []config.EcosystemConfig{
			{
				Name:    "pnpm",
				Install: config.InstallConfig{Command: "pnpm install"},
			},
		},
	}

	ec := &mock.EventCollector[pipeline.Event]{}
	phase := runDeps(runner, "/repo/feature-auth", opts, ec.Emit)

	if phase.Steps[0].Status != pipeline.StepFailed {
		t.Errorf("step status = %v, want pipeline.StepFailed", phase.Steps[0].Status)
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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		fmt.Sprintf("%s:shell[pnpm install --filter packages/ui]", wtPath):   {Output: ""},
		fmt.Sprintf("%s:shell[pnpm install --filter packages/core]", wtPath): {Output: ""},
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

	ec := &mock.EventCollector[pipeline.Event]{}
	phase := runDeps(runner, wtPath, opts, ec.Emit)

	// 2 workspace installs (root install skipped when workspaces detected)
	if len(phase.Steps) != 2 {
		t.Fatalf("step count = %d, want 2", len(phase.Steps))
	}

	// Verify all steps completed
	for i, step := range phase.Steps {
		if step.Status != pipeline.StepDone {
			t.Errorf("step[%d] %q status = %v, want pipeline.StepDone", i, step.Name, step.Status)
		}
	}

	// Verify events contain "running" and "done" for each
	runningCount := 0
	for _, e := range ec.Events {
		if e.Status == pipeline.StepRunning {
			runningCount++
		}
	}
	if runningCount < 2 {
		t.Errorf("expected at least 2 running events, got %d", runningCount)
	}
}

func TestRunDeps_CommandParsing(t *testing.T) {
	runner := &mock.Runner{Responses: map[string]mock.Response{
		"/wt:shell[go mod download]": {Output: ""},
	}}

	opts := Options{
		Ecosystems: []config.EcosystemConfig{
			{
				Name:    "go",
				Install: config.InstallConfig{Command: "go mod download"},
			},
		},
	}

	ec := &mock.EventCollector[pipeline.Event]{}
	runDeps(runner, "/wt", opts, ec.Emit)

	if len(runner.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(runner.Calls))
	}
	if !strings.Contains(runner.Calls[0], "shell[go mod download]") {
		t.Errorf("call = %q, expected to contain 'shell[go mod download]'", runner.Calls[0])
	}
}
