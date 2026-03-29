package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// cargoMembersRe matches quoted strings in a TOML array, used to parse Cargo.toml members.
var cargoMembersRe = regexp.MustCompile(`"([^"]+)"`)

// DetectWorkspaces parses the given configFile inside rootDir and returns
// relative directory paths for each workspace member. Returns nil if the file
// doesn't exist or has no workspaces defined.
func DetectWorkspaces(rootDir, configFile string) ([]string, error) {
	path := filepath.Join(rootDir, configFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	switch configFile {
	case "pnpm-workspace.yaml":
		return parsePnpmWorkspace(rootDir, data)
	case "package.json":
		return parseNpmWorkspace(rootDir, data)
	case "go.work":
		return parseGoWork(rootDir, data)
	case "Cargo.toml":
		return parseCargoWorkspace(rootDir, data)
	default:
		return nil, fmt.Errorf("unsupported workspace config file: %s", configFile)
	}
}

func parsePnpmWorkspace(rootDir string, data []byte) ([]string, error) {
	var cfg struct {
		Packages []string `yaml:"packages"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, nil
	}
	if len(cfg.Packages) == 0 {
		return nil, nil
	}
	return resolveGlobs(rootDir, cfg.Packages)
}

func parseNpmWorkspace(rootDir string, data []byte) ([]string, error) {
	var pkg struct {
		Workspaces []string `json:"workspaces"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, nil
	}
	if pkg.Workspaces == nil {
		return nil, nil
	}
	if len(pkg.Workspaces) == 0 {
		return nil, nil
	}
	return resolveGlobs(rootDir, pkg.Workspaces)
}

func parseGoWork(rootDir string, data []byte) ([]string, error) {
	content := string(data)
	var dirs []string

	i := 0
	for i < len(content) {
		// Find next "use" keyword
		idx := strings.Index(content[i:], "use")
		if idx == -1 {
			break
		}
		pos := i + idx
		i = pos + 3

		// Skip whitespace after "use"
		j := i
		for j < len(content) && (content[j] == ' ' || content[j] == '\t') {
			j++
		}
		if j >= len(content) {
			break
		}

		if content[j] == '(' {
			// Block form: use ( ... )
			end := strings.Index(content[j:], ")")
			if end == -1 {
				break
			}
			block := content[j+1 : j+end]
			for _, line := range strings.Split(block, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				dir := strings.TrimPrefix(line, "./")
				dirs = append(dirs, dir)
			}
			i = j + end + 1
		} else {
			// Single-line form: use ./dir
			end := strings.IndexAny(content[j:], "\n\r")
			var token string
			if end == -1 {
				token = strings.TrimSpace(content[j:])
				i = len(content)
			} else {
				token = strings.TrimSpace(content[j : j+end])
				i = j + end + 1
			}
			if token != "" {
				dir := strings.TrimPrefix(token, "./")
				dirs = append(dirs, dir)
			}
		}
	}

	return filterExistingDirs(rootDir, dirs), nil
}

func parseCargoWorkspace(rootDir string, data []byte) ([]string, error) {
	content := string(data)

	// Find [workspace] section
	wsIdx := strings.Index(content, "[workspace]")
	if wsIdx == -1 {
		return nil, nil
	}
	section := content[wsIdx:]

	// Find the next section header (a line starting with "[") to bound the workspace section.
	// We look for "\n[" to avoid matching array values like members = ["..."].
	nextSection := strings.Index(section[1:], "\n[")
	if nextSection != -1 {
		section = section[:nextSection+2]
	}

	// Find "members" key
	membersIdx := strings.Index(section, "members")
	if membersIdx == -1 {
		return nil, nil
	}

	// Find the opening bracket of the members array
	arrStart := strings.Index(section[membersIdx:], "[")
	if arrStart == -1 {
		return nil, nil
	}
	arrEnd := strings.Index(section[membersIdx+arrStart:], "]")
	if arrEnd == -1 {
		return nil, nil
	}

	arrayContent := section[membersIdx+arrStart : membersIdx+arrStart+arrEnd+1]
	matches := cargoMembersRe.FindAllStringSubmatch(arrayContent, -1)

	var patterns []string
	for _, m := range matches {
		patterns = append(patterns, m[1])
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	return resolveGlobs(rootDir, patterns)
}

// resolveGlobs expands glob patterns relative to rootDir and returns only
// directories as deduplicated relative paths.
func resolveGlobs(rootDir string, patterns []string) ([]string, error) {
	seen := make(map[string]struct{})
	var result []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(rootDir, pattern))
		if err != nil {
			return nil, fmt.Errorf("expand glob %q: %w", pattern, err)
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}
			rel, err := filepath.Rel(rootDir, match)
			if err != nil {
				continue
			}
			if _, exists := seen[rel]; !exists {
				seen[rel] = struct{}{}
				result = append(result, rel)
			}
		}
	}

	return result, nil
}

// filterExistingDirs returns only the dirs (relative to rootDir) that exist.
func filterExistingDirs(rootDir string, dirs []string) []string {
	var result []string
	for _, dir := range dirs {
		info, err := os.Stat(filepath.Join(rootDir, dir))
		if err == nil && info.IsDir() {
			result = append(result, dir)
		}
	}
	return result
}
