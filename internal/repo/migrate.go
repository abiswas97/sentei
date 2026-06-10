package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abiswas97/sentei/internal/fileutil"
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
	result.WorktreePath = git.WorktreePath(opts.RepoPath, branch)

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
	// A detached HEAD yields an empty branch name. Reject it now, before any
	// destructive phase: otherwise the root is gutted and `worktree add ""` fails,
	// leaving an empty directory with no recovery path.
	if strings.TrimSpace(branch) == "" {
		stepErr := errors.New("cannot migrate a detached HEAD; check out a branch first")
		step := StepResult{Name: "Detect current branch", Status: StepFailed, Error: stepErr}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: stepErr})
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
		// Report NO backup (not the constructed path). cp -a may have left a
		// partial copy; remove it. Otherwise the failure screen would tell the
		// user to `rm -rf <repo> && mv <missing-backup> <repo>` — deleting their
		// still-intact repo (the Migrate phase never ran).
		_ = fileutil.RemoveAllRetry(backupPath)
		step := StepResult{Name: "Copy repository to backup", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase, "", ""
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

	// Capture the real origin URL before cloning. `git clone --bare .git .bare`
	// rewrites origin to the local .git path, which we then delete; without this
	// the migrated repo's origin points at a dead path and push/pull is severed.
	originURL, _ := runner.Run(repoPath, "remote", "get-url", "origin")

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

	// Remove original .git. Retry: it was just read by clone --bare, so on macOS
	// Spotlight may briefly hold its object dir (ENOTEMPTY) — a single RemoveAll
	// would fail the migration after the backup was already taken.
	emit(Event{Phase: phaseName, Step: "Remove original .git", Status: StepRunning})
	gitDir := filepath.Join(repoPath, ".git")
	if err := fileutil.RemoveAllRetry(gitDir); err != nil {
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

	// Restore the real origin URL (clone --bare set it to the local .git path).
	// Best-effort: a local-only repo has no origin to restore.
	if strings.TrimSpace(originURL) != "" {
		emit(Event{Phase: phaseName, Step: "Restore origin remote", Status: StepRunning})
		if _, err := runner.Run(barePath, "remote", "set-url", "origin", originURL); err != nil {
			step := StepResult{Name: "Restore origin remote", Status: StepFailed, Error: err}
			phase.Steps = append(phase.Steps, step)
			emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
			return phase
		}
		phase.Steps = append(phase.Steps, StepResult{Name: "Restore origin remote", Status: StepDone, Message: originURL})
		emit(Event{Phase: phaseName, Step: "Restore origin remote", Status: StepDone, Message: originURL})
	}

	// Remove old working files from root (they'll be in the worktree instead)
	// Keep only .bare, .git (pointer), and directories that are worktrees
	emit(Event{Phase: phaseName, Step: "Clean root directory", Status: StepRunning})
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		step := StepResult{Name: "Clean root directory", Status: StepFailed, Error: err}
		phase.Steps = append(phase.Steps, step)
		emit(Event{Phase: phaseName, Step: step.Name, Status: StepFailed, Error: err})
		return phase
	}
	for _, entry := range entries {
		name := entry.Name()
		// Keep .bare, .git pointer, and any hidden config (.sentei.yaml, etc.)
		if name == ".bare" || name == ".git" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(repoPath, name)); err != nil {
			emit(Event{Phase: phaseName, Step: "Clean root directory", Status: StepRunning,
				Message: fmt.Sprintf("warning: could not remove %s: %v", name, err)})
		}
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Clean root directory", Status: StepDone})
	emit(Event{Phase: phaseName, Step: "Clean root directory", Status: StepDone})

	// Create worktree for current branch. The branch is passed explicitly as
	// the commit-ish: without it, git derives a NEW branch from the path's
	// basename instead of checking out the existing one.
	emit(Event{Phase: phaseName, Step: "Create worktree", Status: StepRunning})
	_, err = runner.Run(repoPath, "worktree", "add", git.WorktreePath(repoPath, branch), branch)
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

func runMigrateCopy(backupPath, worktreePath string, emit func(Event)) Phase {
	phase := Phase{Name: "Copy"}
	phaseName := "Copy"

	emit(Event{Phase: phaseName, Step: "Restore working files", Status: StepRunning})

	// Copy EVERYTHING from the backup working tree (untracked, ignored, and
	// uncommitted-modified files) into the new worktree, which otherwise holds
	// only committed content. This makes the worktree a faithful copy so the
	// backup is genuinely redundant and safe to delete. Skip git internals.
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		// No backup to restore from (e.g. a clean repo with nothing extra).
		phase.Steps = append(phase.Steps, StepResult{Name: "Restore working files", Status: StepDone, Message: "nothing to restore"})
		emit(Event{Phase: phaseName, Step: "Restore working files", Status: StepDone, Message: "nothing to restore"})
		return phase
	}

	copied := 0
	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" || name == ".bare" {
			continue
		}
		src := filepath.Join(backupPath, name)
		dst := filepath.Join(worktreePath, name)
		if copyErr := copyTree(src, dst); copyErr != nil {
			emit(Event{Phase: phaseName, Step: "Restore working files", Status: StepRunning,
				Message: fmt.Sprintf("warning: could not copy %s: %v", name, copyErr)})
			continue
		}
		copied++
	}

	msg := fmt.Sprintf("%d items restored", copied)
	if copied == 0 {
		msg = "nothing to restore"
	}
	phase.Steps = append(phase.Steps, StepResult{Name: "Restore working files", Status: StepDone, Message: msg})
	emit(Event{Phase: phaseName, Step: "Restore working files", Status: StepDone, Message: msg})

	return phase
}

// copyTree recursively copies src to dst. It recreates symlinks rather than
// following them, and replaces an existing dst rather than writing through it,
// so restoring over a checked-out symlink cannot corrupt the link's target
// (which may live outside the worktree).
func copyTree(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	switch {
	case info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		_ = os.Remove(dst)
		return os.Symlink(target, dst)

	case info.IsDir():
		if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil

	default:
		// Replace dst instead of writing through an existing file/symlink.
		_ = os.Remove(dst)
		return fileutil.CopyFile(src, dst)
	}
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

// RestoreCommand returns the shell command that undoes a migration by restoring
// the pre-migration backup over the repo root. Single source of truth for the
// CLI and TUI failure screens.
func (r MigrateResult) RestoreCommand() string {
	// Quote operands: a repo path with spaces would otherwise make `rm -rf`
	// delete the wrong directories.
	return fmt.Sprintf("rm -rf %q && mv %q %q", r.BareRoot, r.BackupPath, r.BareRoot)
}
