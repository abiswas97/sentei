package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/repo"
)

func makeMigrateConfirmModel(opts *MigrateOpts) Model {
	m := NewMenuModel(nil, nil, "/some/repo", &config.Config{}, repo.ContextNonBareRepo)
	m.width = 80
	m.height = 24
	if opts != nil {
		m.SetMigrateOpts(opts)
	} else {
		m.view = migrateConfirmView
	}
	return m
}

func TestSetMigrateOpts_EntersConfirmView(t *testing.T) {
	m := NewMenuModel(nil, nil, "/some/repo", &config.Config{}, repo.ContextNonBareRepo)

	opts := &MigrateOpts{DeleteBackup: true, RepoPath: "/some/repo"}
	m.SetMigrateOpts(opts)

	if m.view != migrateConfirmView {
		t.Errorf("expected view=migrateConfirmView, got %d", m.view)
	}
	if m.migrateOpts != opts {
		t.Error("expected migrateOpts to be set")
	}
}

func TestSetMigrateOpts_OverridesRepoPath(t *testing.T) {
	m := NewMenuModel(nil, nil, "/original", &config.Config{}, repo.ContextNonBareRepo)

	m.SetMigrateOpts(&MigrateOpts{RepoPath: "/overridden/repo"})

	if m.repoPath != "/overridden/repo" {
		t.Errorf("expected repoPath='/overridden/repo', got %q", m.repoPath)
	}
}

func TestSetMigrateOpts_EmptyRepoPathKeepsOriginal(t *testing.T) {
	m := NewMenuModel(nil, nil, "/original", &config.Config{}, repo.ContextNonBareRepo)

	m.SetMigrateOpts(&MigrateOpts{RepoPath: ""})

	if m.repoPath != "/original" {
		t.Errorf("expected repoPath='/original', got %q", m.repoPath)
	}
}

func TestMigrateConfirmationVM_RepoPathAndDeleteBackup(t *testing.T) {
	m := makeMigrateConfirmModel(&MigrateOpts{DeleteBackup: true, RepoPath: "/my/repo"})
	vm := m.migrateConfirmationVM()

	if vm.Title != "Confirm Migration" {
		t.Errorf("expected title 'Confirm Migration', got %q", vm.Title)
	}

	output := stripAnsi(vm.View())
	if !strings.Contains(output, "/my/repo") {
		t.Errorf("expected repo path in view, got:\n%s", output)
	}
	if !strings.Contains(output, "yes") {
		t.Errorf("expected delete backup 'yes' in view, got:\n%s", output)
	}
	if !strings.Contains(output, "--delete-backup") {
		t.Errorf("expected '--delete-backup' in CLI command, got:\n%s", output)
	}
}

func TestMigrateConfirmationVM_NoDeleteBackup(t *testing.T) {
	m := makeMigrateConfirmModel(&MigrateOpts{DeleteBackup: false, RepoPath: "/my/repo"})
	vm := m.migrateConfirmationVM()

	output := stripAnsi(vm.View())
	if !strings.Contains(output, "no") {
		t.Errorf("expected delete backup 'no' in view, got:\n%s", output)
	}
	if strings.Contains(output, "--delete-backup") {
		t.Errorf("expected no '--delete-backup' in CLI command, got:\n%s", output)
	}
	if !strings.Contains(output, "sentei migrate") {
		t.Errorf("expected 'sentei migrate' in CLI command, got:\n%s", output)
	}
}

func TestMigrateConfirmView_DirectLaunchUsesConfirmationVM(t *testing.T) {
	m := makeMigrateConfirmModel(&MigrateOpts{DeleteBackup: true, RepoPath: "/my/repo"})

	output := stripAnsi(m.viewMigrateConfirm())

	if !strings.Contains(output, "Confirm Migration") {
		t.Error("expected 'Confirm Migration' title in direct-launch view")
	}
	if !strings.Contains(output, "/my/repo") {
		t.Error("expected repo path in direct-launch view")
	}
}

func TestMigrateConfirmView_MenuLaunchUsesOriginalView(t *testing.T) {
	m := makeMigrateConfirmModel(nil)

	output := stripAnsi(m.viewMigrateConfirm())

	if !strings.Contains(output, "Migrate to Bare Repository") {
		t.Errorf("expected 'Migrate to Bare Repository' in menu-launch view, got:\n%s", output)
	}
}

func TestUpdateMigrateConfirm_BackQuitsWhenDirectLaunch(t *testing.T) {
	m := makeMigrateConfirmModel(&MigrateOpts{RepoPath: "/repo"})

	updated, cmd := m.updateMigrateConfirm(ConfirmBackMsg{})
	_ = updated.(Model)

	if cmd == nil {
		t.Fatal("expected quit cmd when launched directly with opts")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestUpdateMigrateConfirm_BackReturnsToMenu(t *testing.T) {
	m := makeMigrateConfirmModel(nil)

	updated, cmd := m.updateMigrateConfirm(ConfirmBackMsg{})
	result := updated.(Model)

	if result.view != menuView {
		t.Errorf("expected view=menuView, got %d", result.view)
	}
	if cmd != nil {
		t.Error("expected nil cmd when going back to menu")
	}
}

func TestUpdateMigrateConfirm_MigrateInfoUpdatesState(t *testing.T) {
	m := makeMigrateConfirmModel(nil)

	updated, _ := m.updateMigrateConfirm(migrateInfoMsg{branch: "main", isDirty: true})
	result := updated.(Model)

	if result.repo.migrateInfo.Branch != "main" {
		t.Errorf("expected branch='main', got %q", result.repo.migrateInfo.Branch)
	}
	if !result.repo.migrateInfo.IsDirty {
		t.Error("expected isDirty=true")
	}
}

func TestUpdateMigrateConfirm_WindowSizeMsg(t *testing.T) {
	m := makeMigrateConfirmModel(nil)

	updated, _ := m.updateMigrateConfirm(tea.WindowSizeMsg{Width: 120, Height: 40})
	result := updated.(Model)

	if result.width != 120 {
		t.Errorf("expected width=120, got %d", result.width)
	}
}

func TestUpdateMigrateConfirm_QuitKey(t *testing.T) {
	m := makeMigrateConfirmModel(nil)

	_, cmd := m.updateMigrateConfirm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit cmd for q key")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}
