package creator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/abiswas97/sentei/internal/integration"
)

func TestScanArtifacts(t *testing.T) {
	tests := []struct {
		name      string
		dirs      []string
		integs    []integration.Integration
		wantCount int
	}{
		{
			name: "finds code-review-graph artifacts",
			dirs: []string{".code-review-graph"},
			integs: []integration.Integration{
				{
					Name: "code-review-graph",
					Teardown: integration.TeardownSpec{
						Dirs: []string{".code-review-graph/"},
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "no artifacts present",
			dirs: nil,
			integs: []integration.Integration{
				{
					Name: "code-review-graph",
					Teardown: integration.TeardownSpec{
						Dirs: []string{".code-review-graph/"},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "multiple integration artifacts",
			dirs: []string{".code-review-graph", ".cocoindex_code"},
			integs: []integration.Integration{
				{
					Name:     "code-review-graph",
					Teardown: integration.TeardownSpec{Dirs: []string{".code-review-graph/"}},
				},
				{
					Name:     "cocoindex-code",
					Teardown: integration.TeardownSpec{Dirs: []string{".cocoindex_code/"}},
				},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wtDir := t.TempDir()
			for _, d := range tt.dirs {
				os.MkdirAll(filepath.Join(wtDir, d), 0755)
			}

			artifacts := ScanArtifacts(wtDir, tt.integs)
			if len(artifacts) != tt.wantCount {
				t.Errorf("artifact count = %d, want %d", len(artifacts), tt.wantCount)
			}
		})
	}
}

func TestTeardown_WithCommand(t *testing.T) {
	wtDir := t.TempDir()
	os.MkdirAll(filepath.Join(wtDir, ".cocoindex_code"), 0755)

	runner := &mockRunner{responses: map[string]mockResponse{
		wtDir + ":shell[ccc reset --all --force]": {output: "reset"},
	}}

	integs := []integration.Integration{
		{
			Name: "cocoindex-code",
			Teardown: integration.TeardownSpec{
				Command: "ccc reset --all --force",
				Dirs:    []string{".cocoindex_code/"},
			},
		},
	}

	ec := &eventCollector{}
	results := Teardown(runner, wtDir, integs, ec.emit)

	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1", len(results))
	}
	if results[0].Status != StepDone {
		t.Errorf("status = %v, want StepDone", results[0].Status)
	}
}

func TestTeardown_CommandFailsFallsBackToDirDelete(t *testing.T) {
	wtDir := t.TempDir()
	artifactDir := filepath.Join(wtDir, ".cocoindex_code")
	os.MkdirAll(artifactDir, 0755)
	os.WriteFile(filepath.Join(artifactDir, "index.db"), []byte("data"), 0644)

	runner := &mockRunner{responses: map[string]mockResponse{
		wtDir + ":shell[ccc reset --all --force]": {err: fmt.Errorf("command not found")},
	}}

	integs := []integration.Integration{
		{
			Name: "cocoindex-code",
			Teardown: integration.TeardownSpec{
				Command: "ccc reset --all --force",
				Dirs:    []string{".cocoindex_code/"},
			},
		},
	}

	ec := &eventCollector{}
	results := Teardown(runner, wtDir, integs, ec.emit)

	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1", len(results))
	}
	if results[0].Status != StepDone {
		t.Errorf("status = %v, want StepDone (fallback should succeed)", results[0].Status)
	}

	if _, err := os.Stat(artifactDir); !os.IsNotExist(err) {
		t.Error("expected artifact directory to be deleted")
	}
}

func TestTeardown_NoArtifacts(t *testing.T) {
	wtDir := t.TempDir()
	runner := &mockRunner{responses: map[string]mockResponse{}}

	integs := []integration.Integration{
		{
			Name:     "code-review-graph",
			Teardown: integration.TeardownSpec{Dirs: []string{".code-review-graph/"}},
		},
	}

	ec := &eventCollector{}
	results := Teardown(runner, wtDir, integs, ec.emit)

	if len(results) != 0 {
		t.Errorf("result count = %d, want 0 (no artifacts)", len(results))
	}
	if len(ec.events) != 0 {
		t.Errorf("event count = %d, want 0", len(ec.events))
	}
}

func TestTeardown_DirOnlyNoCommand(t *testing.T) {
	wtDir := t.TempDir()
	artifactDir := filepath.Join(wtDir, ".code-review-graph")
	os.MkdirAll(artifactDir, 0755)

	runner := &mockRunner{responses: map[string]mockResponse{}}

	integs := []integration.Integration{
		{
			Name:     "code-review-graph",
			Teardown: integration.TeardownSpec{Dirs: []string{".code-review-graph/"}},
		},
	}

	ec := &eventCollector{}
	results := Teardown(runner, wtDir, integs, ec.emit)

	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1", len(results))
	}
	if results[0].Status != StepDone {
		t.Errorf("status = %v, want StepDone", results[0].Status)
	}

	if _, err := os.Stat(artifactDir); !os.IsNotExist(err) {
		t.Error("expected artifact directory to be deleted")
	}
}
