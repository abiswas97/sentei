package tui

import (
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/repo"
)

func TestCleanupScanning_RendersSpinnerFrame(t *testing.T) {
	m := NewMenuModel(bareDirRunner("/repo"), nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.view = cleanupPreviewView
	m.cleanupScan = nil

	view := stripAnsi(m.viewCleanupPreview())
	if !strings.Contains(view, "Scanning repository") {
		t.Fatalf("scanning text missing:\n%s", view)
	}
	if !strings.Contains(view, stripAnsi(m.spin.View())) {
		t.Errorf("scanning line must carry the spinner frame %q:\n%s", m.spin.View(), view)
	}
}

func TestMenuLoading_RendersSpinner(t *testing.T) {
	m := NewMenuModel(bareDirRunner("/repo"), nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.width = 80

	view := stripAnsi(m.viewMenu())
	if !strings.Contains(view, stripAnsi(m.spin.View())+" loading…") {
		t.Errorf("loading menu item must render spinner + loading…:\n%s", view)
	}
}

func TestMenuLoaded_NoSpinner(t *testing.T) {
	m := NewMenuModel(bareDirRunner("/repo"), nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.width = 80
	updated, _ := m.Update(worktreeContextMsg{worktrees: nil, generation: m.worktreeGeneration})
	m = updated.(Model)

	if strings.Contains(stripAnsi(m.viewMenu()), "loading…") {
		t.Errorf("loaded menu must not show the loading state:\n%s", stripAnsi(m.viewMenu()))
	}
}

func TestSpinnerTick_WhileScanning_SchedulesNext(t *testing.T) {
	m := NewMenuModel(bareDirRunner("/repo"), nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.view = cleanupPreviewView
	m.cleanupScan = nil

	updated, cmd := m.Update(m.spin.Tick())
	if cmd == nil {
		t.Error("tick during an indeterminate wait must schedule the next tick")
	}
	if updated.(Model).spin.View() == m.spin.View() {
		t.Error("tick must advance the spinner frame")
	}
}

func TestSpinnerTick_WhenIdle_Swallowed(t *testing.T) {
	m := NewMenuModel(bareDirRunner("/repo"), nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	updated, _ := m.Update(worktreeContextMsg{worktrees: nil, generation: m.worktreeGeneration})
	m = updated.(Model)

	if _, cmd := m.Update(m.spin.Tick()); cmd != nil {
		t.Error("tick with no indeterminate wait visible must not schedule another tick")
	}
}
