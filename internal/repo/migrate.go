package repo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abiswas97/sentei/internal/git"
)

type MigrateOptions struct {
	RepoPath string
}

type MigrateResult struct {
	BareRoot     string
	WorktreePath string
	BackupPath   string
	BackupSize   string
	Branch       string
	IsDirty      bool
	Phases       []Phase
}

func Migrate(runner git.CommandRunner, shell git.ShellRunner, opts MigrateOptions, emit func(Event)) MigrateResult {
	result := MigrateResult{BareRoot: opts.RepoPath}

	// Phase 1: Validate
	validatePhase, branch, isDirty := runMigrateValidate(runner, opts.RepoPath, emit)
	result.Phases = append(result.Phases, validatePhase)
	result.Branch = branch
	result.IsDirty = isDirty
	if validatePhase.HasFailures() {
		return result
	}

	// Phase 2: Backup
	backupPhase, backupPath, backupSize := runMigrateBackup(shell, opts.RepoPath, emit)
	result.Phases = append(result.Phases, backupPhase)
	result.BackupPath = backupPath
	result.BackupSize = backupSize
	if backupPhase.HasFailures() {
		return result
	}

	// Phase 3: Migrate
	migratePhase := runMigrateBare(runner, opts.RepoPath, branch, emit)
	result.Phases = append(result.Phases, migratePhase)
	if migratePhase.HasFailures() {
		return result
	}
	result.WorktreePath = filepath.Join(opts.RepoPath, branch)

	// Phase 4: Copy (best-effort)
	copyPhase := runMigrateCopy(backupPath, result.WorktreePath, emit)
	result.Phases = append(result.Phases, copyPhase)

	return result
}

func runMigrateValidate(runner git.CommandRunner, repoPath string, emit func(Event)) (Phase, string, bool) {
	phase := Phase{Name: "Validate"}
	phaseName := "Validate"

	// Check for uncommitted changes
	emit(Event{Phase: phaseName, Step: "Check repository status", Status: StepRunning})
	statusOutput, err := runner.Run(repoPath, "status", "--porcelain")
	if err != nil {
		step := StepResult{Name: "Check repository status", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase, "", false
	}
	isDirty := strings.TrimSpace(statusOutput) != ""
	if isDirty {
		emit(Event{Phase: phaseName, Step: "Check repository status", Status: StepDone, Message: "uncommitted changes detected"})
	} else {
		emit(Event{Phase: phaseName, Step: "Check repository status", Status: StepDone, Message: "clean"})
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Check repository status", Status: StepDone})

	// Detect current branch
	emit(Event{Phase: phaseName, Step: "Detect current branch", Status: StepRunning})
	branch, err := runner.Run(repoPath, "branch", "--show-current")
	if err != nil {
		step := StepResult{Name: "Detect current branch", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase, "", isDirty
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Detect current branch", Status: StepDone, Message: branch})
	emit(Event{Phase: phaseName, Step: "Detect current branch", Status: StepDone, Message: branch})

	return phase, branch, isDirty
}

func runMigrateBackup(shell git.ShellRunner, repoPath string, emit func(Event)) (Phase, string, string) {
	phase := Phase{Name: "Backup"}
	phaseName := "Backup"

	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s_backup_%s", repoPath, timestamp)

	emit(Event{Phase: phaseName, Step: "Copy repository to backup", Status: StepRunning})
	parentDir := filepath.Dir(repoPath)
	cpCmd := fmt.Sprintf("cp -a %q %q", repoPath, backupPath)
	_, err := shell.RunShell(parentDir, cpCmd)
	if err != nil {
		step := StepResult{Name: "Copy repository to backup", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase, backupPath, ""
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Copy repository to backup", Status: StepDone, Message: backupPath})
	emit(Event{Phase: phaseName, Step: "Copy repository to backup", Status: StepDone, Message: backupPath})

	// Calculate backup size
	emit(Event{Phase: phaseName, Step: "Calculate backup size", Status: StepRunning})
	size := calculateDirSize(backupPath)
	phase.Steps = append(phase.Steps, StepResult{Name: "Calculate backup size", Status: StepDone, Message: size})
	emit(Event{Phase: phaseName, Step: "Calculate backup size", Status: StepDone, Message: size})

	return phase, backupPath, size
}

func runMigrateBare(runner git.CommandRunner, repoPath, branch string, emit func(Event)) Phase {
	phase := Phase{Name: "Migrate"}
	phaseName := "Migrate"
	barePath := filepath.Join(repoPath, ".bare")

	// Clone bare
	emit(Event{Phase: phaseName, Step: "Create bare repository", Status: StepRunning})
	_, err := runner.Run(repoPath, "clone", "--bare", ".git", barePath)
	if err != nil {
		step := StepResult{Name: "Create bare repository", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create bare repository", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create bare repository", Status: StepDone})

	// Remove original .git
	emit(Event{Phase: phaseName, Step: "Remove original .git", Status: StepRunning})
	gitDir := filepath.Join(repoPath, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		step := StepResult{Name: "Remove original .git", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Remove original .git", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Remove original .git", Status: StepDone})

	// Create .git pointer
	emit(Event{Phase: phaseName, Step: "Create .git pointer", Status: StepRunning})
	gitPointerPath := filepath.Join(repoPath, ".git")
	if err := os.WriteFile(gitPointerPath, []byte("gitdir: .bare\n"), 0644); err != nil {
		step := StepResult{Name: "Create .git pointer", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create .git pointer", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create .git pointer", Status: StepDone})

	// Configure refspec
	emit(Event{Phase: phaseName, Step: "Configure refspec", Status: StepRunning})
	_, err = runner.Run(barePath, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	if err != nil {
		step := StepResult{Name: "Configure refspec", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Configure refspec", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Configure refspec", Status: StepDone})

	// Create worktree for current branch
	emit(Event{Phase: phaseName, Step: "Create worktree", Status: StepRunning})
	_, err = runner.Run(repoPath, "worktree", "add", branch)
	if err != nil {
		step := StepResult{Name: "Create worktree", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Create worktree", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Create worktree", Status: StepDone})

	return phase
}

// copyPatterns defines files/directories to copy from backup to new worktree.
var copyPatterns = []string{
	".env*",
	"node_modules",
	"vendor",
	"build",
	"dist",
	".vscode",
	".idea",
}

func runMigrateCopy(backupPath, worktreePath string, emit func(Event)) Phase {
	phase := Phase{Name: "Copy"}
	phaseName := "Copy"

	emit(Event{Phase: phaseName, Step: "Copy development files", Status: StepRunning})
	copied := 0
	for _, pattern := range copyPatterns {
		matches, err := filepath.Glob(filepath.Join(backupPath, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			name := filepath.Base(match)
			dst := filepath.Join(worktreePath, name)
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.IsDir() {
				if err := copyDir(match, dst); err != nil {
					emit(Event{Phase: phaseName, Step: "Copy development files", Status: StepRunning,
						Message: fmt.Sprintf("warning: could not copy %s: %v", name, err)})
					continue
				}
			} else {
				if err := copyFile(match, dst); err != nil {
					emit(Event{Phase: phaseName, Step: "Copy development files", Status: StepRunning,
						Message: fmt.Sprintf("warning: could not copy %s: %v", name, err)})
					continue
				}
			}
			copied++
		}
	}

	msg := fmt.Sprintf("%d items copied", copied)
	if copied == 0 {
		msg = "no development files found to copy"
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Copy development files", Status: StepDone, Message: msg})
	emit(Event{Phase: phaseName, Step: "Copy development files", Status: StepDone, Message: msg})

	return phase
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		return copyFile(path, dstPath)
	})
}

func calculateDirSize(path string) string {
	var totalSize int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})
	return formatSize(totalSize)
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// DeleteBackup removes the backup directory.
func DeleteBackup(backupPath string) error {
	return os.RemoveAll(backupPath)
}
