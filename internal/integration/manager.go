package integration

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiswas97/sentei/internal/fileutil"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
)

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
	emit func(progress.Event),
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
		} else {
			// Tool already installed — skip install and dep steps.
			for _, dep := range integ.Dependencies {
				emit(progress.Event{Phase: wtPath, Step: InstallDependencyStepName(dep), Status: progress.StepSkipped, Message: "already installed"})
			}
			emit(progress.Event{Phase: wtPath, Step: InstallStepName(integ), Status: progress.StepSkipped, Message: "already installed"})
		}

		// Step 3: run setup.
		stepName := SetupStepName(integ)
		emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepRunning})

		workDir := wtPath
		if integ.Setup.WorkingDir == "repo" {
			workDir = repoPath
		}
		// Quote the worktree path (it embeds the branch name) before it enters a
		// command run via sh -c, so a branch like "a&&rm -rf x" cannot inject.
		cmd := strings.ReplaceAll(integ.Setup.Command, "{path}", git.ShellQuote(wtPath))

		if _, err := shell.RunShell(workDir, cmd); err != nil {
			emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepFailed, Error: err})
			continue
		}
		emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepDone})

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
	emit func(progress.Event),
) {
	for _, wtPath := range wtPaths {
		// Step 1: run teardown command.
		if integ.Teardown.Command != "" {
			stepName := TeardownStepName(integ)
			emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepRunning})
			if _, err := shell.RunShell(wtPath, integ.Teardown.Command); err != nil {
				emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepFailed, Error: err})
				// Continue to directory removal even if command fails.
			} else {
				emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepDone})
			}
		}

		// Step 2: remove artifact directories.
		for _, dir := range integ.Teardown.Dirs {
			dirName := strings.TrimSuffix(dir, "/")
			fullPath := filepath.Join(wtPath, dirName)
			stepName := RemoveDirStepName(dir, wtPath)
			emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepRunning})
			if err := os.RemoveAll(fullPath); err != nil {
				emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepFailed, Error: err})
			} else {
				emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepDone})
			}
		}
	}
}

// detectTool returns true if the integration's binary/command is available.
func detectTool(shell git.ShellRunner, wtPath string, integ Integration) bool {
	cmd := integ.Detect.Command
	if cmd == "" && integ.Detect.BinaryName != "" {
		// Presence, not flags: not every CLI implements --version (ccc
		// exits 2), and a tool installed by any manager must be detected.
		cmd = "command -v " + integ.Detect.BinaryName
	}
	if cmd == "" {
		return false
	}
	_, err := shell.RunShell(wtPath, cmd)
	return err == nil
}

// installTool checks dependencies and installs the integration tool.
// It emits events and returns an error on failure.
func installTool(shell git.ShellRunner, wtPath string, integ Integration, emit func(progress.Event)) error {
	// Check and install dependencies.
	for _, dep := range integ.Dependencies {
		stepName := InstallDependencyStepName(dep)
		_, err := shell.RunShell(wtPath, dep.Detect)
		if err != nil && dep.Install != "" {
			emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepRunning})
			if _, err2 := shell.RunShell(wtPath, dep.Install); err2 != nil {
				emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepFailed, Error: err2})
				return err2
			}
			emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepDone})
		} else {
			emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepSkipped})
		}
	}

	// Install the tool itself.
	stepName := InstallStepName(integ)
	emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepRunning})
	if _, err := shell.RunShell(wtPath, integ.Install.Command); err != nil {
		emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepFailed, Error: err})
		return err
	}
	emit(progress.Event{Phase: wtPath, Step: stepName, Status: progress.StepDone})
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
