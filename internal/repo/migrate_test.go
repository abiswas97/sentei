package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// alwaysOkShell is a ShellRunner that succeeds for all calls — used to
// satisfy the backup phase without needing to predict the timestamp-based path.
type alwaysOkShell struct{}

func (s *alwaysOkShell) RunShell(_ string, _ string) (string, error) {
	return "", nil
}

func TestMigrate_Successful(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "my-project")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	barePath := filepath.Join(repoPath, ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		// Validate
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {output: "main"},
		// Migrate
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		fmt.Sprintf("%s:[worktree add main]", repoPath):                                              {output: ""},
	}}

	ec := &eventCollector{}
	opts := MigrateOptions{RepoPath: repoPath}
	result := Migrate(runner, &alwaysOkShell{}, opts, ec.emit)

	if result.BareRoot != repoPath {
		t.Errorf("BareRoot = %q, want %q", result.BareRoot, repoPath)
	}
	if result.WorktreePath != filepath.Join(repoPath, "main") {
		t.Errorf("WorktreePath = %q, want %q", result.WorktreePath, filepath.Join(repoPath, "main"))
	}
	if result.BackupPath == "" {
		t.Error("expected BackupPath to be set")
	}
	if !strings.Contains(result.BackupPath, "_backup_") {
		t.Errorf("BackupPath should contain _backup_: %q", result.BackupPath)
	}

	// No phase failures
	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
}

func TestMigrate_DirtyRepo_WarningContinues(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "dirty-project")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	barePath := filepath.Join(repoPath, ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):                                             {output: "M file.txt"},
		fmt.Sprintf("%s:[branch --show-current]", repoPath):                                          {output: "develop"},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		fmt.Sprintf("%s:[worktree add develop]", repoPath):                                           {output: ""},
	}}

	ec := &eventCollector{}
	opts := MigrateOptions{RepoPath: repoPath}
	result := Migrate(runner, &alwaysOkShell{}, opts, ec.emit)

	// Should still succeed — dirty is a warning, not a failure
	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}

	// Check that a warning event was emitted for dirty state
	foundWarning := false
	for _, e := range ec.events {
		if strings.Contains(e.Message, "uncommitted") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("expected warning event about uncommitted changes")
	}
}

func TestMigrate_CloneFailure_ShowsRollbackInfo(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "fail-project")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	barePath := filepath.Join(repoPath, ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {output: "main"},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath): {
			output: "", err: fmt.Errorf("fatal: failed to clone"),
		},
	}}

	ec := &eventCollector{}
	opts := MigrateOptions{RepoPath: repoPath}
	result := Migrate(runner, &alwaysOkShell{}, opts, ec.emit)

	migratePhase := findPhase(result.Phases, "Migrate")
	if migratePhase == nil {
		t.Fatal("expected Migrate phase")
	}
	if !migratePhase.HasFailures() {
		t.Error("expected Migrate phase to have failures")
	}
	// Backup should still exist for rollback
	if result.BackupPath == "" {
		t.Error("backup path should be set even on migration failure")
	}
}

func TestDeleteBackup(t *testing.T) {
	dir := t.TempDir()
	backupDir := filepath.Join(dir, "backup")
	os.MkdirAll(backupDir, 0755)
	os.WriteFile(filepath.Join(backupDir, "file.txt"), []byte("data"), 0644)

	err := DeleteBackup(backupDir)
	if err != nil {
		t.Fatalf("DeleteBackup() error: %v", err)
	}
	if _, err := os.Stat(backupDir); !os.IsNotExist(err) {
		t.Error("backup directory should be deleted")
	}
}

func findPhase(phases []Phase, name string) *Phase {
	for i := range phases {
		if phases[i].Name == name {
			return &phases[i]
		}
	}
	return nil
}
