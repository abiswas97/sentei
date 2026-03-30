package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/integration"
)

func TestRunIntegrations_NoIntegrations(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{}}
	opts := Options{Integrations: nil}
	ec := &eventCollector{}

	phase := runIntegrations(runner, "/wt", opts, ec.emit)

	if len(phase.Steps) != 0 {
		t.Errorf("step count = %d, want 0", len(phase.Steps))
	}
}

func TestRunIntegrations_AlreadyInstalled(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/wt:shell[code-review-graph --version]":          {output: "1.0.0"},
		"/repo:shell[code-review-graph build --repo /wt]": {output: "built"},
	}}

	opts := Options{
		RepoPath: "/repo",
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

	phase := runIntegrations(runner, "/wt", opts, ec.emit)

	// Should have steps: detect + setup
	hasSetup := false
	for _, s := range phase.Steps {
		if strings.Contains(s.Name, "setup") || strings.Contains(s.Name, "Setup") {
			hasSetup = true
		}
	}
	if !hasSetup {
		t.Error("expected setup step to be present")
	}
}

func TestRunIntegrations_InstallRequired(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		// Detect fails first time
		"/wt:shell[code-review-graph --version]": {err: fmt.Errorf("not found")},
		// Dependency checks
		`/wt:shell[python3 -c "import sys; assert sys.version_info >= (3,10)"]`: {output: ""},
		"/wt:shell[pipx --version]": {output: "1.0"},
		// Install
		"/wt:shell[pipx install code-review-graph]": {output: "installed"},
		// Setup (working dir = repo, so runs from opts.RepoPath)
		"/repo:shell[code-review-graph build --repo /wt]": {output: "built"},
	}}

	opts := Options{
		RepoPath: "/repo",
		Integrations: []integration.Integration{
			{
				Name: "code-review-graph",
				Dependencies: []integration.Dependency{
					{
						Name:   "python3.10+",
						Detect: `python3 -c "import sys; assert sys.version_info >= (3,10)"`,
					},
					{
						Name:    "pipx",
						Detect:  "pipx --version",
						Install: "brew install pipx",
					},
				},
				Detect: integration.DetectSpec{
					Command: "code-review-graph --version",
				},
				Install: integration.InstallSpec{
					Command: "pipx install code-review-graph",
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

	phase := runIntegrations(runner, "/wt", opts, ec.emit)

	hasFailed := false
	for _, s := range phase.Steps {
		if s.Status == StepFailed {
			hasFailed = true
		}
	}
	if hasFailed {
		t.Error("expected no failures when install + setup succeed")
	}
}

func TestRunIntegrations_SetupFailure(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/wt:shell[ccc --version]": {output: "1.0"},
		"/wt:shell[ccc init]":      {err: fmt.Errorf("init failed")},
	}}

	opts := Options{
		RepoPath: "/repo",
		Integrations: []integration.Integration{
			{
				Name: "cocoindex-code",
				Detect: integration.DetectSpec{
					Command: "ccc --version",
				},
				Setup: integration.SetupSpec{
					Command:    "ccc init",
					WorkingDir: "worktree",
				},
			},
		},
	}

	ec := &eventCollector{}
	phase := runIntegrations(runner, "/wt", opts, ec.emit)

	hasFailed := false
	for _, s := range phase.Steps {
		if s.Status == StepFailed {
			hasFailed = true
		}
	}
	if !hasFailed {
		t.Error("expected a failure when setup command fails")
	}
}

func TestCopyIntegrationIndex_CopiesFromSource(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	srcIndex := filepath.Join(sourceDir, ".cocoindex_code")
	os.MkdirAll(srcIndex, 0755)
	os.WriteFile(filepath.Join(srcIndex, "settings.yml"), []byte("test"), 0644)

	err := copyIntegrationIndex(sourceDir, targetDir, ".cocoindex_code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	targetIndex := filepath.Join(targetDir, ".cocoindex_code", "settings.yml")
	if _, err := os.Stat(targetIndex); os.IsNotExist(err) {
		t.Error("expected settings.yml to be copied")
	}
}

func TestCopyIntegrationIndex_NoSourceIndex_ReturnsError(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	err := copyIntegrationIndex(sourceDir, targetDir, ".cocoindex_code")
	if err == nil {
		t.Error("expected error when source index doesn't exist")
	}
}

func TestAppendGitignore(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		entries  []string
		want     string
	}{
		{
			name:     "adds new entries",
			existing: "node_modules/\n",
			entries:  []string{".code-review-graph/"},
			want:     "node_modules/\n.code-review-graph/\n",
		},
		{
			name:     "skips existing entries",
			existing: ".code-review-graph/\n",
			entries:  []string{".code-review-graph/"},
			want:     ".code-review-graph/\n",
		},
		{
			name:     "creates file if absent",
			existing: "",
			entries:  []string{".code-review-graph/", ".cocoindex_code/"},
			want:     ".code-review-graph/\n.cocoindex_code/\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			gitignorePath := filepath.Join(dir, ".gitignore")

			if tt.existing != "" {
				os.WriteFile(gitignorePath, []byte(tt.existing), 0644)
			}

			appendGitignore(dir, tt.entries)

			got, _ := os.ReadFile(gitignorePath)
			if string(got) != tt.want {
				t.Errorf("gitignore content:\ngot:  %q\nwant: %q", string(got), tt.want)
			}
		})
	}
}
