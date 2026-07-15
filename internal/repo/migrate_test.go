package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/testutil/mock"
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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		// Validate
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {Output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {Output: "main"},
		// Migrate
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {Output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {Output: ""},
		fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):                            {Output: ""},
	}}

	ec := &mock.EventCollector[progress.Event]{}
	opts := MigrateOptions{RepoPath: repoPath}
	result := Migrate(runner, &alwaysOkShell{}, opts, ec.Emit)

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
			if step.Status == progress.StepFailed {
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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):                                             {Output: "M file.txt"},
		fmt.Sprintf("%s:[branch --show-current]", repoPath):                                          {Output: "develop"},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {Output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {Output: ""},
		fmt.Sprintf("%s:[worktree add %s/develop develop]", repoPath, repoPath):                      {Output: ""},
	}}

	ec := &mock.EventCollector[progress.Event]{}
	opts := MigrateOptions{RepoPath: repoPath}
	result := Migrate(runner, &alwaysOkShell{}, opts, ec.Emit)

	// Should still succeed — dirty is a warning, not a failure
	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == progress.StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}

	// Check that a warning event was emitted for dirty state
	foundWarning := false
	for _, e := range ec.Events {
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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {Output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {Output: "main"},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath): {
			Output: "", Err: fmt.Errorf("fatal: failed to clone"),
		},
	}}

	ec := &mock.EventCollector[progress.Event]{}
	opts := MigrateOptions{RepoPath: repoPath}
	result := Migrate(runner, &alwaysOkShell{}, opts, ec.Emit)

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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {Output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {Output: ""}, // detached HEAD
	}}

	ec := &mock.EventCollector[progress.Event]{}
	result := Migrate(runner, &alwaysOkShell{}, MigrateOptions{RepoPath: repoPath}, ec.Emit)

	validate := findPhase(result.Phases, "Validate")
	if validate == nil || !validate.HasFailures() {
		t.Fatal("detached HEAD must fail validation")
	}
	for _, phaseName := range []string{"Backup", "Migrate", "Copy"} {
		phase := findPhase(result.Phases, phaseName)
		if phase == nil {
			t.Fatalf("prepared phase %q missing", phaseName)
		}
		for _, step := range phase.Steps {
			if step.Status != progress.StepSkipped {
				t.Errorf("%s/%s status = %v, want skipped", phaseName, step.Name, step.Status)
			}
		}
	}
	for _, c := range runner.Calls {
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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):                                             {Output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath):                                          {Output: "main"},
		fmt.Sprintf("%s:[remote get-url origin]", repoPath):                                          {Output: originURL},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {Output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {Output: ""},
		fmt.Sprintf("%s:[remote set-url origin %s]", barePath, originURL):                            {Output: ""},
		fmt.Sprintf("%s:[worktree add %s/main main]", repoPath, repoPath):                            {Output: ""},
	}}

	ec := &mock.EventCollector[progress.Event]{}
	result := Migrate(runner, &alwaysOkShell{}, MigrateOptions{RepoPath: repoPath}, ec.Emit)

	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == progress.StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
	restored := false
	for _, c := range runner.Calls {
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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):                                             {Output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath):                                          {Output: "feature/foo"},
		fmt.Sprintf("%s:[clone --bare .git %s]", repoPath, barePath):                                 {Output: ""},
		fmt.Sprintf("%s:[config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*]", barePath): {Output: ""},
		fmt.Sprintf("%s:[worktree add %s/feature-foo feature/foo]", repoPath, repoPath):              {Output: ""},
	}}

	ec := &mock.EventCollector[progress.Event]{}
	result := Migrate(runner, &alwaysOkShell{}, MigrateOptions{RepoPath: repoPath}, ec.Emit)

	for _, phase := range result.Phases {
		for _, step := range phase.Steps {
			if step.Status == progress.StepFailed {
				t.Errorf("step %q failed: %v", step.Name, step.Error)
			}
		}
	}
	// The two-arg form (path + existing branch) checks out feature/foo rather
	// than inventing a divergent "foo" branch from the basename, and the
	// worktree directory flattens the slash.
	twoArg := false
	for _, c := range runner.Calls {
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

func findPhase(phases []progress.Phase, name string) *progress.Phase {
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

	runner := &mock.Runner{Responses: map[string]mock.Response{
		fmt.Sprintf("%s:[status --porcelain]", repoPath):    {Output: ""},
		fmt.Sprintf("%s:[branch --show-current]", repoPath): {Output: "main"},
	}}

	ec := &mock.EventCollector[progress.Event]{}
	result := Migrate(runner, &failingShell{}, MigrateOptions{RepoPath: repoPath}, ec.Emit)

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
	// The Migrate phase is declared up front but every operation remains skipped.
	migrate := findPhase(result.Phases, "Migrate")
	if migrate == nil {
		t.Fatal("prepared Migrate phase missing")
	}
	for _, step := range migrate.Steps {
		if step.Status != progress.StepSkipped {
			t.Errorf("Migrate/%s status = %v, want skipped", step.Name, step.Status)
		}
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
