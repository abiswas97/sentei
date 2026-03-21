package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DedupConfig(configPath string, opts Options, emit func(Event)) (ConfigResult, error) {
	emit(Event{Step: "dedup-config", Message: "Deduplicating git config...", Level: LevelStep})

	data, err := os.ReadFile(configPath)
	if err != nil {
		return ConfigResult{}, fmt.Errorf("reading config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	before := len(lines)

	var out []string
	seen := make(map[string]bool)

	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			seen = make(map[string]bool)
			out = append(out, line)
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			out = append(out, line)
			continue
		}

		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, line)
	}

	after := len(out)
	removed := before - after
	result := ConfigResult{Before: before, After: after, Removed: removed}

	if removed == 0 {
		emit(Event{Step: "dedup-config", Message: fmt.Sprintf("Config already clean (%d lines)", before), Level: LevelInfo})
		return result, nil
	}

	if opts.DryRun {
		emit(Event{Step: "dedup-config", Message: fmt.Sprintf("Would remove %d duplicate lines (%d → %d)", removed, before, after), Level: LevelDetail})
		return result, nil
	}

	if err := atomicWriteConfig(configPath, strings.Join(out, "\n")); err != nil {
		return result, fmt.Errorf("writing deduped config: %w", err)
	}

	emit(Event{Step: "dedup-config", Message: fmt.Sprintf("Deduplicated config: removed %d lines (%d → %d)", removed, before, after), Level: LevelInfo})
	return result, nil
}

func atomicWriteConfig(configPath string, content string) error {
	bakPath := configPath + ".bak"

	info, err := os.Stat(configPath)
	var perm os.FileMode = 0644
	if err == nil {
		perm = info.Mode().Perm()
	}

	original, err := os.ReadFile(configPath)
	if err == nil {
		if err := os.WriteFile(bakPath, original, perm); err != nil {
			return fmt.Errorf("creating backup: %w", err)
		}
	}

	dir := filepath.Dir(configPath)
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("setting permissions: %w", err)
	}

	if err := os.Rename(tmpPath, configPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
