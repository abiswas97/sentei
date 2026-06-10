package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/pipeline"
	"github.com/abiswas97/sentei/internal/repo"
)

func makeMigrateSummaryModel(result repo.MigrateResult) Model {
	m := NewMenuModel(nil, nil, "/repo", &config.Config{}, repo.ContextNoRepo)
	m.repo.result = result
	m.width = 80
	m.height = 24
	return m
}

func TestViewMigrateSummary_BackupFailure_NoDestructiveRestore(t *testing.T) {
	// Backup failed -> BackupPath empty (the data-loss fix). The destructive
	// restore command must NOT render against the still-intact repo.
	result := repo.MigrateResult{
		BareRoot: "/repo/proj",
		Phases: []pipeline.Phase{
			{Name: "Validate", Steps: []pipeline.StepResult{{Name: "Detect current branch", Status: pipeline.StepDone}}},
			{Name: "Backup", Steps: []pipeline.StepResult{{Name: "Copy repository to backup", Status: pipeline.StepFailed, Error: errors.New("no space left")}}},
		},
		// BackupPath intentionally empty
	}
	out := stripAnsi(makeMigrateSummaryModel(result).viewMigrateSummary())
	if !strings.Contains(out, "Migration failed") {
		t.Errorf("expected a failure header:\n%s", out)
	}
	if strings.Contains(out, "rm -rf") {
		t.Errorf("a backup failure must NOT show a destructive restore command:\n%s", out)
	}
}

func TestViewMigrateSummary_MigrateFailure_ShowsRestore(t *testing.T) {
	result := repo.MigrateResult{
		BareRoot:   "/repo/proj",
		BackupPath: "/repo/proj_backup_1",
		Phases: []pipeline.Phase{
			{Name: "Backup", Steps: []pipeline.StepResult{{Name: "Copy repository to backup", Status: pipeline.StepDone}}},
			{Name: "Migrate", Steps: []pipeline.StepResult{{Name: "Create bare repository", Status: pipeline.StepFailed, Error: errors.New("boom")}}},
		},
	}
	out := stripAnsi(makeMigrateSummaryModel(result).viewMigrateSummary())
	if !strings.Contains(out, "Migration failed") {
		t.Errorf("expected a failure header:\n%s", out)
	}
	if !strings.Contains(out, "rm -rf") || !strings.Contains(out, "/repo/proj_backup_1") {
		t.Errorf("a real backup should yield a restore command:\n%s", out)
	}
}

func TestViewMigrateSummary_Success_OffersDeleteBackup(t *testing.T) {
	result := repo.MigrateResult{
		BareRoot:   "/repo/proj",
		BackupPath: "/repo/proj_backup_1",
		Branch:     "main",
		Phases: []pipeline.Phase{
			{Name: "Migrate", Steps: []pipeline.StepResult{{Name: "Create worktree", Status: pipeline.StepDone}}},
		},
	}
	out := stripAnsi(makeMigrateSummaryModel(result).viewMigrateSummary())
	if !strings.Contains(out, "migrated") {
		t.Errorf("expected a success header:\n%s", out)
	}
	if !strings.Contains(out, "Delete backup?") {
		t.Errorf("expected the delete-backup prompt:\n%s", out)
	}
}

func TestUpdateMigrateSummary_Failure_YIsInert(t *testing.T) {
	result := repo.MigrateResult{
		BareRoot: "/repo/proj",
		Phases: []pipeline.Phase{
			{Name: "Backup", Steps: []pipeline.StepResult{{Name: "x", Status: pipeline.StepFailed, Error: errors.New("no space")}}},
		},
	}
	m := makeMigrateSummaryModel(result)
	// On a failed migration, 'y' (delete backup) must be inert — never delete a
	// backup / advance. Only quit responds.
	_, cmd := m.updateMigrateSummary(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd != nil {
		t.Error("'y' on a failed migration must not delete backup or advance")
	}
}
