package integration

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
)

// ManagerStatus indicates the outcome of a manager operation step.
type ManagerStatus int

const (
	StatusRunning ManagerStatus = iota
	StatusDone
	StatusFailed
)

// ManagerEvent is emitted by EnableIntegration and DisableIntegration to report
// progress for each worktree operation step.
type ManagerEvent struct {
	Worktree string
	Step     string
	Status   ManagerStatus
	Error    error
}

// EnableIntegration installs and sets up integ in every worktree listed in
// wtPaths. For each worktree it:
//  1. Optionally seeds the index directory from mainWTPath.
//  2. Detects whether the tool is already installed; installs it if not.
//  3. Runs the setup command.
//  4. Appends gitignore entries.
//
// Progress is reported via emit.
func EnableIntegration(
	shell git.ShellRunner,
	repoPath string,
	mainWTPath string,
	wtPaths []string,
	integ Integration,
	emit func(ManagerEvent),
) {
	for _, wtPath := range wtPaths {
		// Step 1: seed index from main worktree when applicable.
		if integ.IndexCopyDir != "" && mainWTPath != "" && wtPath != mainWTPath {
			src := filepath.Join(mainWTPath, integ.IndexCopyDir)
			dst := filepath.Join(wtPath, integ.IndexCopyDir)
			// Only copy when source exists; ignore copy errors (non-fatal optimisation).
			if _, err := os.Stat(src); err == nil {
				_ = os.RemoveAll(dst)
				_ = fileutil.CopyDir(src, dst)
			}
		}

		// Step 2: detect and install.
		if !detectTool(shell, wtPath, integ) {
			if err := installTool(shell, wtPath, integ, emit); err != nil {
				// installTool already emitted the failure event.
				continue
			}
		}

		// Step 3: run setup.
		stepName := "Setup " + integ.Name
		emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusRunning})

		workDir := wtPath
		if integ.Setup.WorkingDir == "repo" {
			workDir = repoPath
		}
		cmd := strings.ReplaceAll(integ.Setup.Command, "{path}", wtPath)

		if _, err := shell.RunShell(workDir, cmd); err != nil {
			emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusFailed, Error: err})
			continue
		}
		emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusDone})

		// Step 4: append gitignore entries (best-effort; non-fatal).
		_ = appendGitignoreEntries(wtPath, integ.GitignoreEntries)
	}
}

// DisableIntegration tears down integ in every worktree listed in wtPaths.
// For each worktree it runs the teardown command (if any) then removes artifact
// directories.  Progress is reported via emit.
func DisableIntegration(
	shell git.ShellRunner,
	wtPaths []string,
	integ Integration,
	emit func(ManagerEvent),
) {
	for _, wtPath := range wtPaths {
		// Step 1: run teardown command.
		if integ.Teardown.Command != "" {
			stepName := "Teardown " + integ.Name
			emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusRunning})
			if _, err := shell.RunShell(wtPath, integ.Teardown.Command); err != nil {
				emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusFailed, Error: err})
				// Continue to directory removal even if command fails.
			} else {
				emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusDone})
			}
		}

		// Step 2: remove artifact directories.
		for _, dir := range integ.Teardown.Dirs {
			dirName := strings.TrimSuffix(dir, "/")
			fullPath := filepath.Join(wtPath, dirName)
			stepName := fmt.Sprintf("Remove %s in %s", dirName, filepath.Base(wtPath))
			emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusRunning})
			if err := os.RemoveAll(fullPath); err != nil {
				emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusFailed, Error: err})
			} else {
				emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusDone})
			}
		}
	}
}

// detectTool returns true if the integration's binary/command is available.
func detectTool(shell git.ShellRunner, wtPath string, integ Integration) bool {
	cmd := integ.Detect.Command
	if cmd == "" && integ.Detect.BinaryName != "" {
		cmd = integ.Detect.BinaryName + " --version"
	}
	if cmd == "" {
		return false
	}
	_, err := shell.RunShell(wtPath, cmd)
	return err == nil
}

// installTool checks dependencies and installs the integration tool.
// It emits events and returns an error on failure.
func installTool(shell git.ShellRunner, wtPath string, integ Integration, emit func(ManagerEvent)) error {
	// Check and install dependencies.
	for _, dep := range integ.Dependencies {
		_, err := shell.RunShell(wtPath, dep.Detect)
		if err != nil && dep.Install != "" {
			stepName := "Install dependency " + dep.Name
			emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusRunning})
			if _, err2 := shell.RunShell(wtPath, dep.Install); err2 != nil {
				emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusFailed, Error: err2})
				return err2
			}
			emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusDone})
		}
	}

	// Install the tool itself.
	stepName := "Install " + integ.Name
	emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusRunning})
	if _, err := shell.RunShell(wtPath, integ.Install.Command); err != nil {
		emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusFailed, Error: err})
		return err
	}
	emit(ManagerEvent{Worktree: wtPath, Step: stepName, Status: StatusDone})
	return nil
}

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

// readGitignoreLines returns the set of lines in the .gitignore file at path.
// Returns an empty set if the file does not exist.
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
