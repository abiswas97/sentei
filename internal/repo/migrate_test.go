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
		fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):                            {output: ""},
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
		fmt.Sprintf("%s:[worktree add %s/develop develop]", repoPath, repoPath):                      {output: ""},
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

func TestMigrate_DetachedHead_RejectedBeforeDestruction(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "detached")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {output: ""}, // detached HEAD
	}}

	ec := &eventCollector{}
	result := Migrate(runner, &alwaysOkShell{}, MigrateOptions{RepoPath: repoPath}, ec.emit)

	validate := findPhase(result.Phases, "Validate")
	if validate == nil || !validate.HasFailures() {
		t.Fatal("detached HEAD must fail validation")
	}
	if findPhase(result.Phases, "Backup") != nil || findPhase(result.Phases, "Migrate") != nil {
		t.Error("no destructive phase should run after a validation failure")
	}
	for _, c := range runner.calls {
		if strings.Contains(c, "clone --bare") {
			t.Error("clone --bare must not run for a detached HEAD")
		}
	}
}

func TestMigrate_PreservesOriginURL(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "with-origin")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	barePath := filepath.Join(repoPath, ".bare")
	const originURL = "git@github.com:user/proj.git"

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):                                             {output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath):                                          {output: "main"},
		fmt.Sprintf("%s:[remote get-url origin]", repoPath):                                          {output: originURL},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		fmt.Sprintf("%s:[remote set-url origin %s]", barePath, originURL):                            {output: ""},
		fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):                            {output: ""},
	}}

	ec := &eventCollector{}
	result := Migrate(runner, &alwaysOkShell{}, MigrateOptions{RepoPath: repoPath}, ec.emit)

	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
	restored := false
	for _, c := range runner.calls {
		if strings.Contains(c, fmt.Sprintf("remote set-url origin %s", originURL)) {
			restored = true
		}
	}
	if !restored {
		t.Error("the real origin URL must be restored on the migrated bare repo")
	}
}

func TestMigrate_SlashBranch_ChecksOutExistingBranch(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "slash")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	barePath := filepath.Join(repoPath, ".bare")

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):                                             {output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath):                                          {output: "feature/foo"},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {output: ""},
		fmt.Sprintf("%s:[worktree add %s/feature-foo feature/foo]", repoPath, repoPath):              {output: ""},
	}}

	ec := &eventCollector{}
	result := Migrate(runner, &alwaysOkShell{}, MigrateOptions{RepoPath: repoPath}, ec.emit)

	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
	// The two-arg form (path + existing branch) checks out feature/foo rather
	// than inventing a divergent "foo" branch from the basename, and the
	// worktree directory flattens the slash.
	twoArg := false
	for _, c := range runner.calls {
		if strings.Contains(c, "[worktree add "+repoPath+"/feature-foo feature/foo]") {
			twoArg = true
		}
	}
	if !twoArg {
		t.Error("slash branch must use the two-arg worktree add form with a flattened path")
	}
	if want := filepath.Join(repoPath, "feature-foo"); result.WorktreePath != want {
		t.Errorf("WorktreePath = %q, want %q", result.WorktreePath, want)
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

// failingShell fails every shell command (used to fail the backup cp).
type failingShell struct{}

func (s *failingShell) RunShell(_ string, _ string) (string, error) {
	return "", fmt.Errorf("cp -a: No space left on device")
}

func TestMigrate_BackupFailure_LeavesNoDestructiveRestore(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "proj")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	runner := &mockRunner{responses: map[string]mockResponse{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {output: "main"},
	}}

	ec := &eventCollector{}
	result := Migrate(runner, &failingShell{}, MigrateOptions{RepoPath: repoPath}, ec.emit)

	backup := findPhase(result.Phases, "Backup")
	if backup == nil || !backup.HasFailures() {
		t.Fatal("expected the Backup phase to fail")
	}
	// Critical: with no valid backup, BackupPath must be empty. Both the CLI and
	// TUI gate the (destructive) restore command on BackupPath != "", so an empty
	// path is what prevents telling the user to rm -rf their still-intact repo.
	if result.BackupPath != "" {
		t.Errorf("BackupPath must be empty on backup failure, got %q", result.BackupPath)
	}
	// The Migrate phase must never have run, so the repo root is untouched.
	if findPhase(result.Phases, "Migrate") != nil {
		t.Error("Migrate phase must not run after a backup failure")
	}
}

func TestMigrateResult_RestoreCommand_QuotesPaths(t *testing.T) {
	r := MigrateResult{BareRoot: "/my repo", BackupPath: "/my repo_backup_1"}
	got := r.RestoreCommand()
	want := `rm -rf "/my repo" && mv "/my repo_backup_1" "/my repo"`
	if got != want {
		t.Errorf("RestoreCommand() = %q, want %q", got, want)
	}
}

func TestCopyTree_DoesNotWriteThroughSymlinks(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "precious.txt")
	if err := os.WriteFile(outside, []byte("PRECIOUS"), 0644); err != nil {
		t.Fatal(err)
	}

	// Case 1: a regular-file source must replace a dst symlink, not write through
	// it to the (outside-the-worktree) target.
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	os.MkdirAll(src, 0755)
	os.MkdirAll(dst, 0755)
	os.WriteFile(filepath.Join(src, "f"), []byte("new content"), 0644)
	os.Symlink(outside, filepath.Join(dst, "f")) // checked-out committed symlink

	if err := copyTree(filepath.Join(src, "f"), filepath.Join(dst, "f")); err != nil {
		t.Fatalf("copyTree (file over symlink): %v", err)
	}
	if got, _ := os.ReadFile(outside); string(got) != "PRECIOUS" {
		t.Errorf("write-through corrupted the symlink target: %q", got)
	}

	// Case 2: a symlink source is recreated as a symlink, not dereferenced.
	os.Symlink(outside, filepath.Join(src, "link"))
	if err := copyTree(filepath.Join(src, "link"), filepath.Join(dst, "link")); err != nil {
		t.Fatalf("copyTree (symlink): %v", err)
	}
	info, err := os.Lstat(filepath.Join(dst, "link"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("source symlink should be recreated as a symlink, not dereferenced")
	}
}
