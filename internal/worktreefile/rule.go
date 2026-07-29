package worktreefile

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// Rule describes one file copied from the repository container into a worktree.
type Rule struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Overwrite   bool   `yaml:"overwrite,omitempty"`
}

// ValidateRules rejects ambiguous or escaping paths before worktree creation.
func ValidateRules(rules []Rule) error {
	destinations := make(map[string]struct{}, len(rules))
	for i, rule := range rules {
		if err := validateRelativePath(rule.Source); err != nil {
			return fmt.Errorf("worktree_files[%d].source: %w", i, err)
		}
		if err := validateRelativePath(rule.Destination); err != nil {
			return fmt.Errorf("worktree_files[%d].destination: %w", i, err)
		}
		destination := path.Clean(strings.ReplaceAll(rule.Destination, "\\", "/"))
		if _, exists := destinations[destination]; exists {
			return fmt.Errorf("worktree_files[%d].destination: duplicate destination %q", i, rule.Destination)
		}
		destinations[destination] = struct{}{}
	}
	return nil
}

func validateRelativePath(value string) error {
	trimmed := strings.TrimSpace(value)
	portable := strings.ReplaceAll(trimmed, "\\", "/")
	if trimmed == "" || path.Clean(portable) == "." {
		return fmt.Errorf("path must not be empty")
	}
	if filepath.IsAbs(trimmed) || path.IsAbs(portable) {
		return fmt.Errorf("path must be relative")
	}
	for _, component := range strings.Split(portable, "/") {
		if component == ".." {
			return fmt.Errorf("path must not contain .. traversal")
		}
	}
	return nil
}
