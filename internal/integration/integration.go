package integration

// Integration describes a dev tool that can be installed and managed per-worktree.
type Integration struct {
	Name             string
	Description      string
	URL              string
	Dependencies     []Dependency
	Detect           DetectSpec
	Install          InstallSpec
	Setup            SetupSpec
	Teardown         TeardownSpec
	GitignoreEntries []string
}

// Dependency is a prerequisite tool needed before an integration can be installed.
type Dependency struct {
	Name    string
	Detect  string // shell command to check presence
	Install string // shell command to install (may be empty)
}

// DetectSpec describes how to check whether an integration is already installed.
type DetectSpec struct {
	Command    string
	BinaryName string
}

// InstallSpec describes how to install an integration.
type InstallSpec struct {
	Command      string
	FirstRunNote string
}

// SetupSpec describes how to initialise an integration inside a worktree or repo.
type SetupSpec struct {
	Command    string
	WorkingDir string // "repo" or "worktree"
}

// TeardownSpec describes how to remove an integration's artefacts.
type TeardownSpec struct {
	Command string
	Dirs    []string
}

var registry []Integration

func register(i Integration) {
	registry = append(registry, i)
}

// All returns every registered integration.
func All() []Integration {
	return registry
}

// Get returns the integration with the given name, or nil if not found.
func Get(name string) *Integration {
	for i := range registry {
		if registry[i].Name == name {
			return &registry[i]
		}
	}
	return nil
}
