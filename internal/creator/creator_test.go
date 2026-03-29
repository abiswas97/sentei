package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/integration"
)

func TestRun_FullPipeline(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:[worktree add /repo/feature-auth -b feature/auth main]": {output: ""},
		"/repo/feature-auth:[merge main --no-edit]":                    {output: ""},
		"/repo/feature-auth:[go mod download]":                         {output: ""},
		"/repo/feature-auth:[code-review-graph --version]":             {output: "1.0"},
		"/repo:[code-review-graph build --repo /repo/feature-auth]":    {output: ""},
	}}

	opts := Options{
		BranchName:     "feature/auth",
		BaseBranch:     "main",
		RepoPath:       "/repo",
		SourceWorktree: "/repo/main",
		MergeBase:      true,
		CopyEnvFiles:   false,
		Ecosystems: []config.EcosystemConfig{
			{
				Name:    "go",
				Install: config.InstallConfig{Command: "go mod download"},
			},
		},
		Integrations: []integration.Integration{
			{
				Name: "code-review-graph",
				Detect: integration.DetectSpec{
					Command: "code-review-graph --version",
				},
				Setup: integration.SetupSpec{
					Command:    "code-review-graph build --repo {path}",
					WorkingDir: "repo",
				},
				GitignoreEntries: []string{".code-review-graph/"},
			},
		},
	}

	ec := &eventCollector{}
	result := Run(runner, opts, ec.emit)

	if result.WorktreePath != "/repo/feature-auth" {
		t.Errorf("WorktreePath = %q, want %q", result.WorktreePath, "/repo/feature-auth")
	}
	if len(result.Phases) != 3 {
		t.Fatalf("phase count = %d, want 3", len(result.Phases))
	}
	if result.Phases[0].Name != "Setup" {
		t.Errorf("phase[0] = %q, want Setup", result.Phases[0].Name)
	}
	if result.Phases[1].Name != "Dependencies" {
		t.Errorf("phase[1] = %q, want Dependencies", result.Phases[1].Name)
	}
	if result.Phases[2].Name != "Integrations" {
		t.Errorf("phase[2] = %q, want Integrations", result.Phases[2].Name)
	}
	if result.HasFailures() {
		t.Error("expected no failures in full pipeline")
	}
	if len(ec.events) == 0 {
		t.Error("expected events to be emitted")
	}
}

func TestRun_CreateWorktreeFails_AbortsEarly(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:[worktree add /repo/feature-dup -b feature/dup main]": {
			err: fmt.Errorf("fatal: branch already exists"),
		},
	}}

	opts := Options{
		BranchName:     "feature/dup",
		BaseBranch:     "main",
		RepoPath:       "/repo",
		SourceWorktree: "/repo/main",
		MergeBase:      true,
		CopyEnvFiles:   false,
		Ecosystems: []config.EcosystemConfig{
			{Name: "go", Install: config.InstallConfig{Command: "go mod download"}},
		},
	}

	ec := &eventCollector{}
	result := Run(runner, opts, ec.emit)

	if len(result.Phases) != 1 {
		t.Fatalf("phase count = %d, want 1 (abort after setup)", len(result.Phases))
	}
	if result.WorktreePath != "" {
		t.Errorf("WorktreePath = %q, want empty on failure", result.WorktreePath)
	}
}

func TestRun_MergeFailsContinues(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:[worktree add /repo/feature-conflict -b feature/conflict main]": {output: ""},
		"/repo/feature-conflict:[merge main --no-edit]":                        {err: fmt.Errorf("conflict")},
	}}

	opts := Options{
		BranchName:     "feature/conflict",
		BaseBranch:     "main",
		RepoPath:       "/repo",
		SourceWorktree: "/repo/main",
		MergeBase:      true,
		CopyEnvFiles:   false,
	}

	ec := &eventCollector{}
	result := Run(runner, opts, ec.emit)

	if len(result.Phases) != 3 {
		t.Fatalf("phase count = %d, want 3 (continues despite merge failure)", len(result.Phases))
	}
	if result.WorktreePath == "" {
		t.Error("WorktreePath should be set even with merge failure")
	}
	if !result.HasFailures() {
		t.Error("expected HasFailures to return true")
	}
}

func TestRun_CopyEnvFiles(t *testing.T) {
	srcDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, ".env"), []byte("KEY=val"), 0644)

	repoDir := t.TempDir()
	wtPath := filepath.Join(repoDir, "feature-env")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[worktree add %s -b feature/env main]", repoDir, wtPath): {output: ""},
	}}

	// Create the worktree dir since the mock doesn't actually create it
	os.MkdirAll(wtPath, 0755)

	opts := Options{
		BranchName:     "feature/env",
		BaseBranch:     "main",
		RepoPath:       repoDir,
		SourceWorktree: srcDir,
		MergeBase:      false,
		CopyEnvFiles:   true,
		Ecosystems: []config.EcosystemConfig{
			{
				Name:     "node",
				EnvFiles: []string{".env"},
				Install:  config.InstallConfig{Command: ""},
			},
		},
	}

	ec := &eventCollector{}
	result := Run(runner, opts, ec.emit)

	// Verify env file was copied
	envDst := filepath.Join(wtPath, ".env")
	data, err := os.ReadFile(envDst)
	if err != nil {
		t.Fatalf("failed to read copied env file: %v", err)
	}
	if string(data) != "KEY=val" {
		t.Errorf("env file content = %q, want %q", string(data), "KEY=val")
	}

	_ = result
}

func TestResult_HasFailures(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name: "no failures",
			result: Result{
				Phases: []Phase{
					{Steps: []StepResult{{Status: StepDone}, {Status: StepSkipped}}},
				},
			},
			want: false,
		},
		{
			name: "has failure",
			result: Result{
				Phases: []Phase{
					{Steps: []StepResult{{Status: StepDone}, {Status: StepFailed}}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasFailures(); got != tt.want {
				t.Errorf("HasFailures() = %v, want %v", got, tt.want)
			}
		})
	}
}
