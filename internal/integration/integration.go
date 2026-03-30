package integration

// Integration describes a dev tool that can be installed and managed per-worktree.
type Integration struct {
	Name             string
	ShortDescription string // one-line tagline for list views
	Description      string // full description for info dialog
	URL              string
	Dependencies     []Dependency
	Detect           DetectSpec
	Install          InstallSpec
	Setup            SetupSpec
	Teardown         TeardownSpec
	GitignoreEntries []string
	// IndexCopyDir is the directory name (relative to worktree root) that can be
	// copied from one worktree to another to seed an incremental index. Empty means
	// the integration's index cannot be shared across worktrees (e.g., absolute paths).
	IndexCopyDir string
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

// All returns every known integration.
func All() []Integration {
	return []Integration{
		codeReviewGraph(),
		cocoindexCode(),
	}
}

// Get returns the integration with the given name, or nil if not found.
func Get(name string) *Integration {
	all := All()
	for i := range all {
		if all[i].Name == name {
			return &all[i]
		}
	}
	return nil
}

// Names returns the names of all known integrations.
func Names() []string {
	all := All()
	names := make([]string, len(all))
	for i, integ := range all {
		names[i] = integ.Name
	}
	return names
}
