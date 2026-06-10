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
	"github.com/abiswas97/sentei/internal/pipeline"
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
	Phases       []pipeline.Phase
}

func Migrate(runner git.CommandRunner, shell git.ShellRunner, opts MigrateOptions, emit func(pipeline.Event)) MigrateResult {
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

func runMigrateValidate(runner git.CommandRunner, repoPath string, emit func(pipeline.Event)) (pipeline.Phase, string, bool) {
	rec := pipeline.NewPhaseRecorder("Validate", emit)

	isDirty := false
	ok := rec.Step("Check repository status", func() (string, error) {
		statusOutput, err := runner.Run(repoPath, "status", "--porcelain")
		if err != nil {
			return "", err
		}
		isDirty = strings.TrimSpace(statusOutput) != ""
		if isDirty {
			return "uncommitted changes detected", nil
		}
		return "clean", nil
	})
	if !ok {
		return rec.Phase(), "", false
	}

	var branch string
	rec.Step("Detect current branch", func() (string, error) {
		var err error
		branch, err = runner.Run(repoPath, "branch", "--show-current")
		if err != nil {
			return "", err
		}
		// A detached HEAD yields an empty branch name. Reject it now, before any
		// destructive phase: otherwise the root is gutted and `worktree add ""` fails,
		// leaving an empty directory with no recovery path.
		if strings.TrimSpace(branch) == "" {
			return "", errors.New("cannot migrate a detached HEAD; check out a branch first")
		}
		return branch, nil
	})

	return rec.Phase(), branch, isDirty
}

func runMigrateBackup(shell git.ShellRunner, repoPath string, emit func(pipeline.Event)) (pipeline.Phase, string, string) {
	rec := pipeline.NewPhaseRecorder("Backup", emit)

	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s_backup_%s", repoPath, timestamp)

	ok := rec.Step("Copy repository to backup", func() (string, error) {
		cpCmd := fmt.Sprintf("cp -a %q %q", repoPath, backupPath)
		if _, err := shell.RunShell(filepath.Dir(repoPath), cpCmd); err != nil {
			// cp -a may have left a partial copy; remove it. Otherwise the failure
			// screen would tell the user to `rm -rf <repo> && mv <missing-backup>
			// <repo>` — deleting their still-intact repo (the Migrate phase never
			// ran).
			_ = fileutil.RemoveAllRetry(backupPath)
			return "", err
		}
		return backupPath, nil
	})
	if !ok {
		// Report NO backup (not the constructed path).
		return rec.Phase(), "", ""
	}

	var size string
	rec.Step("Calculate backup size", func() (string, error) {
		size = calculateDirSize(backupPath)
		return size, nil
	})

	return rec.Phase(), backupPath, size
}

func runMigrateBare(runner git.CommandRunner, repoPath, branch string, emit func(pipeline.Event)) pipeline.Phase {
	rec := pipeline.NewPhaseRecorder("Migrate", emit)
	barePath := filepath.Join(repoPath, ".bare")

	// Capture the real origin URL before cloning. `git clone --bare .git .bare`
	// rewrites origin to the local .git path, which we then delete; without this
	// the migrated repo's origin points at a dead path and push/pull is severed.
	originURL, _ := runner.Run(repoPath, "remote", "get-url", "origin")

	ok := rec.Step("Create bare repository", func() (string, error) {
		_, err := runner.Run(repoPath, "clone", "--bare", ".git", barePath)
		return "", err
	})
	if !ok {
		return rec.Phase()
	}

	// Retry: .git was just read by clone --bare, so on macOS Spotlight may
	// briefly hold its object dir (ENOTEMPTY) — a single RemoveAll would fail
	// the migration after the backup was already taken.
	ok = rec.Step("Remove original .git", func() (string, error) {
		return "", fileutil.RemoveAllRetry(filepath.Join(repoPath, ".git"))
	})
	if !ok {
		return rec.Phase()
	}

	ok = rec.Step("Create .git pointer", func() (string, error) {
		return "", os.WriteFile(filepath.Join(repoPath, ".git"), []byte("gitdir: .bare\n"), 0644)
	})
	if !ok {
		return rec.Phase()
	}

	ok = rec.Step("Configure refspec", func() (string, error) {
		_, err := runner.Run(barePath, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
		return "", err
	})
	if !ok {
		return rec.Phase()
	}

	// Restore the real origin URL (clone --bare set it to the local .git path).
	// Best-effort: a local-only repo has no origin to restore.
	if strings.TrimSpace(originURL) != "" {
		ok = rec.Step("Restore origin remote", func() (string, error) {
			if _, err := runner.Run(barePath, "remote", "set-url", "origin", originURL); err != nil {
				return "", err
			}
			return originURL, nil
		})
		if !ok {
			return rec.Phase()
		}
	}

	// Remove old working files from root (they'll be in the worktree instead).
	// Keep only .bare and the .git pointer.
	ok = rec.Step("Clean root directory", func() (string, error) {
		entries, err := os.ReadDir(repoPath)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			name := entry.Name()
			if name == ".bare" || name == ".git" {
				continue
			}
			if err := os.RemoveAll(filepath.Join(repoPath, name)); err != nil {
				rec.Emit("Clean root directory", pipeline.StepRunning,
					fmt.Sprintf("warning: could not remove %s: %v", name, err))
			}
		}
		return "", nil
	})
	if !ok {
		return rec.Phase()
	}

	// Create worktree for current branch. The branch is passed explicitly as
	// the commit-ish: without it, git derives a NEW branch from the path's
	// basename instead of checking out the existing one.
	rec.Step("Create worktree", func() (string, error) {
		_, err := runner.Run(repoPath, "worktree", "add", git.WorktreePath(repoPath, branch), branch)
		return "", err
	})

	return rec.Phase()
}

func runMigrateCopy(backupPath, worktreePath string, emit func(pipeline.Event)) pipeline.Phase {
	rec := pipeline.NewPhaseRecorder("Copy", emit)

	// Copy EVERYTHING from the backup working tree (untracked, ignored, and
	// uncommitted-modified files) into the new worktree, which otherwise holds
	// only committed content. This makes the worktree a faithful copy so the
	// backup is genuinely redundant and safe to delete. Skip git internals.
	rec.Step("Restore working files", func() (string, error) {
		entries, err := os.ReadDir(backupPath)
		if err != nil {
			// No backup to restore from (e.g. a clean repo with nothing extra).
			return "nothing to restore", nil
		}

		copied := 0
		for _, entry := range entries {
			name := entry.Name()
			if name == ".git" || name == ".bare" {
				continue
			}
			if copyErr := copyTree(filepath.Join(backupPath, name), filepath.Join(worktreePath, name)); copyErr != nil {
				rec.Emit("Restore working files", pipeline.StepRunning,
					fmt.Sprintf("warning: could not copy %s: %v", name, copyErr))
				continue
			}
			copied++
		}

		if copied == 0 {
			return "nothing to restore", nil
		}
		return fmt.Sprintf("%d items restored", copied), nil
	})

	return rec.Phase()
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
