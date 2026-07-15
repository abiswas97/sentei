package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/repo"
)

func TestUpdateSummary_MenuLaunch_KeysReturnToMenu(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{"enter", tea.KeyPressMsg{Code: tea.KeyEnter}},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEsc}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMenuModel(nil, nil, "/repo", &config.Config{}, repo.ContextBareRepo)
			m.view = summaryView

			updated, cmd := m.updateSummary(tc.msg)

			if updated.(Model).view != menuView {
				t.Errorf("view = %d, want menuView", updated.(Model).view)
			}
			if cmd != nil {
				t.Error("returning to menu should not emit a command")
			}
		})
	}
}

func TestUpdateSummary_MenuLaunch_QuitKeyMatchesFooter(t *testing.T) {
	m := NewMenuModel(nil, nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.view = summaryView

	updated, cmd := m.updateSummary(keyMsg("q"))
	if updated.(Model).view != summaryView {
		t.Fatalf("quit key changed view to %d before exit", updated.(Model).view)
	}
	if cmd == nil {
		t.Fatal("footer advertises q quit, but q emitted no quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("q emitted %T, want tea.QuitMsg", cmd())
	}
}

func TestUpdateSummary_DirectLaunch_KeysQuit(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{"enter", tea.KeyPressMsg{Code: tea.KeyEnter}},
		{"quit key", keyMsg("q")},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEsc}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(nil, nil, "/repo")
			m.view = summaryView

			_, cmd := m.updateSummary(tc.msg)

			if cmd == nil {
				t.Fatal("expected a quit command")
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Errorf("expected tea.QuitMsg, got %T", cmd())
			}
		})
	}
}

func TestUpdateSummary_OtherKeysIgnored(t *testing.T) {
	m := NewModel(nil, nil, "/repo")
	m.view = summaryView

	updated, cmd := m.updateSummary(keyMsg("x"))

	if updated.(Model).view != summaryView || cmd != nil {
		t.Error("unhandled keys should leave the model untouched")
	}
}
