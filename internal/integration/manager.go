package integration

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// appendGitignoreEntries appends missing entries to dir/.gitignore.
func appendGitignoreEntries(dir string, entries []string) error {
	if len(entries) == 0 {
		return nil
	}

	gitignorePath := filepath.Join(dir, ".gitignore")
	existing, err := readGitignoreLines(gitignorePath)
	if err != nil {
		return fmt.Errorf("reading .gitignore: %w", err)
	}
	var toAdd []string
	for _, entry := range entries {
		if !existing[entry] {
			toAdd = append(toAdd, entry)
		}
	}
	if len(toAdd) == 0 {
		return nil
	}
	out, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening .gitignore: %w", err)
	}
	for _, entry := range toAdd {
		if _, err := fmt.Fprintln(out, entry); err != nil {
			_ = out.Close()
			return fmt.Errorf("writing .gitignore: %w", err)
		}
	}
	return out.Close()
}

func readGitignoreLines(path string) (map[string]bool, error) {
	lines := make(map[string]bool)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return lines, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines[scanner.Text()] = true
	}
	return lines, scanner.Err()
}
