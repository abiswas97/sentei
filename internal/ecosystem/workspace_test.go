package ecosystem

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func mkdirs(t *testing.T, base string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			t.Fatalf("mkdirAll %s: %v", d, err)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func sortedCopy(s []string) []string {
	c := make([]string, len(s))
	copy(c, s)
	sort.Strings(c)
	return c
}

func TestDetectWorkspaces_Pnpm(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pnpm-workspace.yaml"), `packages:
  - "packages/*"
  - "apps/*"
`)
	mkdirs(t, root, "packages/core", "packages/utils", "apps/web")

	got, err := DetectWorkspaces(root, "pnpm-workspace.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 dirs, got %d: %v", len(got), got)
	}
	got = sortedCopy(got)
	want := []string{"apps/web", "packages/core", "packages/utils"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("got[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestDetectWorkspaces_Npm(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"workspaces": ["packages/*"]}`)
	mkdirs(t, root, "packages/lib")

	got, err := DetectWorkspaces(root, "package.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "packages/lib" {
		t.Errorf("expected [packages/lib], got %v", got)
	}
}

func TestDetectWorkspaces_PackageJsonNoWorkspaces(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"name": "my-pkg", "version": "1.0.0"}`)

	got, err := DetectWorkspaces(root, "package.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestDetectWorkspaces_GoWork(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), `go 1.21

use (
	./cmd/api
	./pkg/core
)
`)
	mkdirs(t, root, "cmd/api", "pkg/core")

	got, err := DetectWorkspaces(root, "go.work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 dirs, got %d: %v", len(got), got)
	}
	got = sortedCopy(got)
	want := []string{"cmd/api", "pkg/core"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("got[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestDetectWorkspaces_CargoToml(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Cargo.toml"), `[workspace]
members = ["crates/lib", "crates/cli"]
`)
	mkdirs(t, root, "crates/lib", "crates/cli")

	got, err := DetectWorkspaces(root, "Cargo.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 dirs, got %d: %v", len(got), got)
	}
	got = sortedCopy(got)
	want := []string{"crates/cli", "crates/lib"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("got[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestDetectWorkspaces_MissingFile(t *testing.T) {
	root := t.TempDir()

	got, err := DetectWorkspaces(root, "pnpm-workspace.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestDetectWorkspaces_GlobNoMatches(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pnpm-workspace.yaml"), `packages:
  - "packages/*"
`)
	// No packages/ dir created

	got, err := DetectWorkspaces(root, "pnpm-workspace.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestDetectWorkspaces_MalformedPnpmYaml(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pnpm-workspace.yaml"), `{this is: [not valid yaml`)

	got, err := DetectWorkspaces(root, "pnpm-workspace.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}
