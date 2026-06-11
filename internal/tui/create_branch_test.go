package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
)

func createBranchModel() Model {
	m := NewMenuModel(bareDirRunner("/repo"), nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.width, m.height = 80, 24
	m.view = createBranchView
	m.remove.worktrees = []git.Worktree{{Path: "/repo/feature-taken", Branch: "refs/heads/feature/taken"}}
	return m
}

func TestUpdateCreateBranch_TypingFillsFocusedField(t *testing.T) {
	m := createBranchModel()

	updated, _ := m.updateCreateBranch(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("feature/new")})
	model := updated.(Model)

	if got := model.create.branchInput.Value(); got != "feature/new" {
		t.Errorf("branch input = %q, want %q", got, "feature/new")
	}
	if model.create.baseInput.Value() != defaultBaseBranch {
		t.Errorf("base input should be untouched, got %q", model.create.baseInput.Value())
	}
}

func TestUpdateCreateBranch_TabSwitchesFields(t *testing.T) {
	m := createBranchModel()

	updated, _ := m.updateCreateBranch(tea.KeyMsg{Type: tea.KeyTab})
	model := updated.(Model)
	if model.create.focusedField != 1 || !model.create.baseInput.Focused() || model.create.branchInput.Focused() {
		t.Fatal("tab should move focus to the base field")
	}

	updated, _ = model.updateCreateBranch(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	if model.create.focusedField != 0 || !model.create.branchInput.Focused() {
		t.Fatal("second tab should move focus back to the branch field")
	}
}

func TestUpdateCreateBranch_EnterValidates(t *testing.T) {
	cases := []struct {
		name     string
		branch   string
		wantView viewState
		wantErr  bool
	}{
		{"valid advances to options", "feature/new", createOptionsView, false},
		{"empty stays with error", "", createBranchView, true},
		{"duplicate worktree stays with error", "feature/taken", createBranchView, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := createBranchModel()
			m.create.branchInput.SetValue(tc.branch)

			updated, _ := m.updateCreateBranch(tea.KeyMsg{Type: tea.KeyEnter})
			model := updated.(Model)

			if model.view != tc.wantView {
				t.Errorf("view = %d, want %d", model.view, tc.wantView)
			}
			if tc.wantErr && model.create.validationErr == "" {
				t.Error("expected a validation error")
			}
			if !tc.wantErr && model.create.validationErr != "" {
				t.Errorf("unexpected validation error %q", model.create.validationErr)
			}
		})
	}
}

func TestUpdateCreateBranch_EscReturnsToMenu(t *testing.T) {
	m := createBranchModel()
	updated, _ := m.updateCreateBranch(tea.KeyMsg{Type: tea.KeyEsc})
	if updated.(Model).view != menuView {
		t.Error("esc should return to the menu")
	}
}

func TestUpdateCreateBranch_TypingClearsValidationError(t *testing.T) {
	m := createBranchModel()
	m.create.validationErr = "branch name is required"

	updated, _ := m.updateCreateBranch(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if got := updated.(Model).create.validationErr; got != "" {
		t.Errorf("typing should clear the validation error, got %q", got)
	}
}

func TestViewCreateBranch_RendersChromeFieldsAndError(t *testing.T) {
	m := createBranchModel()
	m.create.branchInput.SetValue("feat x")
	m.create.validationErr = "branch name cannot contain spaces"

	view := stripANSI(m.viewCreateBranch())
	for _, want := range []string{"sentei ─ Create Worktree", "Branch name", "Base branch", "branch name cannot contain spaces", "enter continue"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}
