package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_E2E_FullStack(t *testing.T) {
	xdgDir := t.TempDir()
	senteiConfigDir := filepath.Join(xdgDir, "sentei")
	if err := os.MkdirAll(senteiConfigDir, 0o755); err != nil {
		t.Fatalf("MkdirAll global config dir: %v", err)
	}

	globalConfig := `
ecosystems:
  - name: pnpm
    install:
      command: pnpm install --frozen-lockfile
  - name: custom-tool
    detect:
      files:
        - custom.lock
    install:
      command: custom-tool install
protected_branches:
  - main
  - staging
`
	if err := os.WriteFile(filepath.Join(senteiConfigDir, "config.yaml"), []byte(globalConfig), 0o644); err != nil {
		t.Fatalf("WriteFile global config: %v", err)
	}

	repoDir := t.TempDir()
	repoConfig := `
integrations_enabled:
  - code-review-graph
  - cocoindex-code
protected_branches:
  - main
  - production
`
	if err := os.WriteFile(filepath.Join(repoDir, ".sentei.yaml"), []byte(repoConfig), 0o644); err != nil {
		t.Fatalf("WriteFile repo config: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	cfg, err := LoadConfig(repoDir)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// 16 embedded + 1 custom-tool = 17
	if len(cfg.Ecosystems) != 17 {
		t.Fatalf("expected 17 ecosystems, got %d", len(cfg.Ecosystems))
	}

	// pnpm command should be overridden by global config
	var pnpm *EcosystemConfig
	var customTool *EcosystemConfig
	for i := range cfg.Ecosystems {
		switch cfg.Ecosystems[i].Name {
		case "pnpm":
			pnpm = &cfg.Ecosystems[i]
		case "custom-tool":
			customTool = &cfg.Ecosystems[i]
		}
	}

	if pnpm == nil {
		t.Fatal("pnpm ecosystem not found")
	}
	if pnpm.Install.Command != "pnpm install --frozen-lockfile" {
		t.Errorf("pnpm command: got %q, want %q", pnpm.Install.Command, "pnpm install --frozen-lockfile")
	}
	if pnpm.Source != "global" {
		t.Errorf("pnpm source: got %q, want %q", pnpm.Source, "global")
	}

	if customTool == nil {
		t.Fatal("custom-tool ecosystem not found")
	}
	if customTool.Source != "global" {
		t.Errorf("custom-tool source: got %q, want %q", customTool.Source, "global")
	}

	// per-repo protected_branches replaces global
	wantBranches := []string{"main", "production"}
	if len(cfg.ProtectedBranches) != len(wantBranches) {
		t.Fatalf("ProtectedBranches: got %v, want %v", cfg.ProtectedBranches, wantBranches)
	}
	for i, want := range wantBranches {
		if cfg.ProtectedBranches[i] != want {
			t.Errorf("ProtectedBranches[%d]: got %q, want %q", i, cfg.ProtectedBranches[i], want)
		}
	}

	// integrations from per-repo config
	if len(cfg.IntegrationsEnabled) != 2 {
		t.Fatalf("IntegrationsEnabled: got %d, want 2", len(cfg.IntegrationsEnabled))
	}
	wantIntegrations := []string{"code-review-graph", "cocoindex-code"}
	for i, want := range wantIntegrations {
		if cfg.IntegrationsEnabled[i] != want {
			t.Errorf("IntegrationsEnabled[%d]: got %q, want %q", i, cfg.IntegrationsEnabled[i], want)
		}
	}
}

func TestLoadConfig_E2E_MalformedGlobalErrors(t *testing.T) {
	xdgDir := t.TempDir()
	senteiConfigDir := filepath.Join(xdgDir, "sentei")
	if err := os.MkdirAll(senteiConfigDir, 0o755); err != nil {
		t.Fatalf("MkdirAll global config dir: %v", err)
	}

	malformed := "ecosystems: [\nunclosed bracket"
	if err := os.WriteFile(filepath.Join(senteiConfigDir, "config.yaml"), []byte(malformed), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	repoDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	_, err := LoadConfig(repoDir)
	if err == nil {
		t.Fatal("expected error for malformed global config, got nil")
	}
}

func TestLoadConfig_E2E_MalformedRepoErrors(t *testing.T) {
	xdgDir := t.TempDir()
	repoDir := t.TempDir()

	malformed := "ecosystems: [\nunclosed bracket"
	if err := os.WriteFile(filepath.Join(repoDir, ".sentei.yaml"), []byte(malformed), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	_, err := LoadConfig(repoDir)
	if err == nil {
		t.Fatal("expected error for malformed repo config, got nil")
	}
}
