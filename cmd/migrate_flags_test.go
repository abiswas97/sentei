package cmd

import (
	"strings"
	"testing"
)

func TestParseMigrateFlags_NoFlags(t *testing.T) {
	opts, err := ParseMigrateFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.DeleteBackup {
		t.Error("expected DeleteBackup=false")
	}
	if opts.RepoPath != "" {
		t.Errorf("expected empty RepoPath, got %q", opts.RepoPath)
	}
}

func TestParseMigrateFlags_WithDeleteBackup(t *testing.T) {
	opts, err := ParseMigrateFlags([]string{"--delete-backup"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.DeleteBackup {
		t.Error("expected DeleteBackup=true")
	}
}

func TestParseMigrateFlags_WithRepoPath(t *testing.T) {
	opts, err := ParseMigrateFlags([]string{"/some/repo/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.RepoPath != "/some/repo/path" {
		t.Errorf("RepoPath = %q, want %q", opts.RepoPath, "/some/repo/path")
	}
}

func TestParseMigrateFlags_WithDeleteBackupAndRepoPath(t *testing.T) {
	opts, err := ParseMigrateFlags([]string{"--delete-backup", "/some/repo/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.DeleteBackup {
		t.Error("expected DeleteBackup=true")
	}
	if opts.RepoPath != "/some/repo/path" {
		t.Errorf("RepoPath = %q, want %q", opts.RepoPath, "/some/repo/path")
	}
}

func TestValidateMigrateForNonInteractive_MissingPath(t *testing.T) {
	opts := &MigrateOptions{}
	err := ValidateMigrateForNonInteractive(opts)
	if err == nil {
		t.Fatal("expected error for missing repo path")
	}
	if !strings.Contains(err.Error(), "repo path required") {
		t.Errorf("error = %q, want message containing %q", err.Error(), "repo path required")
	}
}

func TestValidateMigrateForNonInteractive_Valid(t *testing.T) {
	opts := &MigrateOptions{RepoPath: "/some/repo"}
	err := ValidateMigrateForNonInteractive(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrateCLICommand_RepoPathOnly(t *testing.T) {
	opts := &MigrateOptions{RepoPath: "/some/repo"}
	result := MigrateCLICommand(opts)
	if result != "sentei migrate /some/repo" {
		t.Errorf("MigrateCLICommand = %q, want %q", result, "sentei migrate /some/repo")
	}
}

func TestMigrateCLICommand_WithDeleteBackup(t *testing.T) {
	opts := &MigrateOptions{RepoPath: "/some/repo", DeleteBackup: true}
	result := MigrateCLICommand(opts)
	if !strings.Contains(result, "sentei migrate") {
		t.Errorf("expected 'sentei migrate' prefix, got %q", result)
	}
	if !strings.Contains(result, "--delete-backup") {
		t.Errorf("expected '--delete-backup' flag, got %q", result)
	}
	if !strings.Contains(result, "/some/repo") {
		t.Errorf("expected repo path, got %q", result)
	}
}
