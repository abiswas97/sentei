package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed defaults/ecosystems.yaml
var defaultEcosystemsYAML []byte

func loadEmbeddedDefaults() (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(defaultEcosystemsYAML, &cfg); err != nil {
		return nil, fmt.Errorf("parsing embedded defaults: %w", err)
	}
	return &cfg, nil
}

// loadFile reads and parses a YAML config file. Returns (nil, nil) if the file
// does not exist.
func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return &cfg, nil
}

// mergeEcosystems performs a keyed merge of overlay into base, using name as
// the key. Existing entries are updated field-by-field (only non-zero overlay
// fields replace base values). New entries are appended. Source is set on all
// changed or new entries.
func mergeEcosystems(base, overlay []EcosystemConfig, overlaySource string) []EcosystemConfig {
	if len(overlay) == 0 {
		return base
	}

	result := make([]EcosystemConfig, len(base))
	copy(result, base)

	index := make(map[string]int, len(result))
	for i, e := range result {
		index[e.Name] = i
	}

	for _, over := range overlay {
		i, exists := index[over.Name]
		if !exists {
			over.Source = overlaySource
			result = append(result, over)
			index[over.Name] = len(result) - 1
			continue
		}

		e := result[i]
		if over.Enabled != nil {
			e.Enabled = over.Enabled
		}
		if len(over.Detect.Files) > 0 {
			e.Detect.Files = over.Detect.Files
		}
		if over.Install.Command != "" {
			e.Install.Command = over.Install.Command
		}
		if over.Install.WorkspaceDetect != "" {
			e.Install.WorkspaceDetect = over.Install.WorkspaceDetect
		}
		if over.Install.WorkspaceInstall != "" {
			e.Install.WorkspaceInstall = over.Install.WorkspaceInstall
		}
		if over.Install.Parallel != nil {
			e.Install.Parallel = over.Install.Parallel
		}
		if len(over.EnvFiles) > 0 {
			e.EnvFiles = over.EnvFiles
		}
		if len(over.PostInstall) > 0 {
			e.PostInstall = over.PostInstall
		}
		e.Source = overlaySource
		result[i] = e
	}

	return result
}

// mergeConfigs merges overlay on top of base, returning a new Config. Scalar
// lists (ProtectedBranches, IntegrationsEnabled) are replaced entirely when
// the overlay provides them.
func mergeConfigs(base, overlay *Config, overlaySource string) *Config {
	result := &Config{
		Ecosystems:          mergeEcosystems(base.Ecosystems, overlay.Ecosystems, overlaySource),
		ProtectedBranches:   base.ProtectedBranches,
		IntegrationsEnabled: base.IntegrationsEnabled,
	}
	if len(overlay.ProtectedBranches) > 0 {
		result.ProtectedBranches = overlay.ProtectedBranches
	}
	if len(overlay.IntegrationsEnabled) > 0 {
		result.IntegrationsEnabled = overlay.IntegrationsEnabled
	}
	return result
}

// knownIntegrations is the set of recognised integration names.
var knownIntegrations = map[string]struct{}{
	"code-review-graph": {},
	"cocoindex-code":    {},
}

// validate checks the config for structural errors and warns about unknown
// integration names.
func validate(cfg *Config) error {
	for i, e := range cfg.Ecosystems {
		if !e.IsEnabled() {
			continue
		}
		if e.Name == "" {
			return fmt.Errorf("ecosystems[%d]: name is required", i)
		}
		if len(e.Detect.Files) == 0 {
			return fmt.Errorf("ecosystem %q: detect.files must not be empty", e.Name)
		}
	}
	for _, name := range cfg.IntegrationsEnabled {
		if _, ok := knownIntegrations[name]; !ok {
			fmt.Fprintf(os.Stderr, "warning: unknown integration %q in integrations_enabled\n", name)
		}
	}
	return nil
}

// globalConfigPath returns the path to the global sentei config file, honouring
// XDG_CONFIG_HOME and defaulting to ~/.config.
func globalConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "sentei", "config.yaml")
}

// resolveRepoRoot resolves a git working directory to the root of the bare
// repository (the directory containing the worktrees). It falls back to
// repoPath on any error.
func resolveRepoRoot(repoPath string) string {
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return repoPath
	}
	commonDir := strings.TrimSpace(string(out))
	if filepath.IsAbs(commonDir) {
		return filepath.Dir(commonDir)
	}
	return filepath.Dir(filepath.Join(repoPath, commonDir))
}

// LoadConfig is the public API for loading sentei configuration. It:
//  1. Loads the embedded defaults (Source="embedded").
//  2. Merges the global config (~/.config/sentei/config.yaml).
//  3. Resolves the repo root and merges .sentei.yaml from it.
//  4. Validates the result.
func LoadConfig(repoPath string) (*Config, error) {
	cfg, err := loadEmbeddedDefaults()
	if err != nil {
		return nil, err
	}
	for i := range cfg.Ecosystems {
		cfg.Ecosystems[i].Source = "embedded"
	}

	globalPath := globalConfigPath()
	globalCfg, err := loadFile(globalPath)
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}
	if globalCfg != nil {
		cfg = mergeConfigs(cfg, globalCfg, "global")
	}

	repoRoot := resolveRepoRoot(repoPath)
	repoCfg, err := loadFile(filepath.Join(repoRoot, ".sentei.yaml"))
	if err != nil {
		return nil, fmt.Errorf("loading repo config: %w", err)
	}
	if repoCfg != nil {
		cfg = mergeConfigs(cfg, repoCfg, "per-repo")
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Config is the top-level configuration for sentei.
type Config struct {
	Ecosystems          []EcosystemConfig `yaml:"ecosystems"`
	ProtectedBranches   []string          `yaml:"protected_branches"`
	IntegrationsEnabled []string          `yaml:"integrations_enabled"`
}

// EcosystemConfig describes how to detect and install a language/tool ecosystem.
type EcosystemConfig struct {
	Name        string        `yaml:"name"`
	Enabled     *bool         `yaml:"enabled,omitempty"`
	Detect      DetectConfig  `yaml:"detect"`
	Install     InstallConfig `yaml:"install"`
	EnvFiles    []string      `yaml:"env_files"`
	PostInstall []string      `yaml:"post_install"`
	Source      string        `yaml:"-"` // "embedded", "global", or "per-repo"
}

// IsEnabled reports whether the ecosystem is active. An absent Enabled field
// is treated as true.
func (e *EcosystemConfig) IsEnabled() bool {
	return e.Enabled == nil || *e.Enabled
}

// DetectConfig holds the file patterns used to detect an ecosystem.
type DetectConfig struct {
	Files []string `yaml:"files"`
}

// InstallConfig describes how to install an ecosystem's dependencies.
type InstallConfig struct {
	Command          string `yaml:"command"`
	WorkspaceDetect  string `yaml:"workspace_detect,omitempty"`
	WorkspaceInstall string `yaml:"workspace_install,omitempty"`
	Parallel         *bool  `yaml:"parallel,omitempty"`
}

// IsParallel reports whether installation should run in parallel. An absent
// Parallel field is treated as false.
func (i *InstallConfig) IsParallel() bool {
	return i.Parallel != nil && *i.Parallel
}
