package integration

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectPresent checks whether an integration's artifacts exist in the given directory.
// It checks for the presence of any directory listed in GitignoreEntries.
func DetectPresent(dir string, integ Integration) bool {
	for _, entry := range integ.GitignoreEntries {
		name := strings.TrimSuffix(entry, "/")
		info, err := os.Stat(filepath.Join(dir, name))
		if err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// DetectAllPresent checks all integrations for artifact presence in the given
// directory. Returns a map of integration name → present.
func DetectAllPresent(dir string, integrations []Integration) map[string]bool {
	result := make(map[string]bool, len(integrations))
	for _, integ := range integrations {
		result[integ.Name] = DetectPresent(dir, integ)
	}
	return result
}

// DetectDeps checks whether each dependency for all integrations is present.
// Returns a map of dep name → installed.
func DetectDeps(shell DepDetector, integrations []Integration) map[string]bool {
	result := make(map[string]bool)
	for _, integ := range integrations {
		for _, dep := range integ.Dependencies {
			if _, checked := result[dep.Name]; checked {
				continue
			}
			_, err := shell.RunShell(".", dep.Detect)
			result[dep.Name] = err == nil
		}
	}
	return result
}

// DepDetector is the subset of git.ShellRunner needed for dep detection.
type DepDetector interface {
	RunShell(dir string, command string) (string, error)
}
