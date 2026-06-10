package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
)

func TestUpdateIntegrationProgress_Finalized_TransitionsToSummary(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.returnView = integrationListView
	genBefore := m.worktreeGeneration

	updated, cmd := m.updateIntegrationProgress(integrationFinalizedMsg{err: nil})
	m = updated.(Model)

	if m.view != integrationSummaryView {
		t.Errorf("expected integrationSummaryView, got %d", m.view)
	}
	if m.integ.saveErr != nil {
		t.Errorf("expected nil saveErr, got %v", m.integ.saveErr)
	}
	if m.worktreeGeneration != genBefore+1 {
		t.Error("expected eager worktree reload generation bump on successful apply")
	}
	if cmd == nil {
		t.Error("expected reload Cmd alongside summary transition")
	}
}

func TestUpdateIntegrationProgress_Finalized_SaveError_TransitionsToSummary(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.returnView = integrationListView
	genBefore := m.worktreeGeneration

	saveErr := errors.New("save failed")
	updated, _ := m.updateIntegrationProgress(integrationFinalizedMsg{err: saveErr})
	m = updated.(Model)

	if m.view != integrationSummaryView {
		t.Errorf("expected integrationSummaryView on save failure, got %d", m.view)
	}
	if !errors.Is(m.integ.saveErr, saveErr) {
		t.Errorf("expected saveErr stored, got %v", m.integ.saveErr)
	}
	if m.worktreeGeneration != genBefore {
		t.Error("worktreeGeneration must not advance on a failed save")
	}
	if m.integ.staged["cocoindex-code"] {
		t.Error("a failed save must not mutate staged state")
	}
}

func TestUpdateIntegrationSummary_Enter_ReturnsToListAndReloads(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationSummaryView

	updated, cmd := m.updateIntegrationSummary(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.view != integrationListView {
		t.Errorf("expected integrationListView, got %d", model.view)
	}
	if cmd == nil {
		t.Error("expected loadIntegrationState Cmd so staged markers reconcile from disk")
	}
}

func TestViewIntegrationSummary_AllSucceeded(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationSummaryView
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/feature-a", Step: "Setup code-review-graph", Status: integration.StatusRunning},
		{Worktree: "/repo/feature-a", Step: "Setup code-review-graph", Status: integration.StatusDone},
		{Worktree: "/repo/feature-b", Step: "Setup code-review-graph", Status: integration.StatusDone},
	}

	view := m.viewIntegrationSummary()
	for _, want := range []string{"feature-a", "feature-b", indicatorDone, "2 steps applied", "enter integrations"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected summary to contain %q, view:\n%s", want, view)
		}
	}
	if strings.Contains(view, "failed") {
		t.Errorf("fully successful apply must not mention failures, view:\n%s", view)
	}
}

func TestViewIntegrationSummary_PartialFailure(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationSummaryView
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/feature-a", Step: "Setup code-review-graph", Status: integration.StatusDone},
		{Worktree: "/repo/feature-b", Step: "Install dependency pipx", Status: integration.StatusFailed, Error: errors.New("brew install pipx: exit 1")},
	}

	view := m.viewIntegrationSummary()
	for _, want := range []string{"1 step applied", "1 failed", indicatorFailed, "brew install pipx: exit 1"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected summary to contain %q, view:\n%s", want, view)
		}
	}
}

func TestViewIntegrationSummary_SaveError(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationSummaryView
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/feature-a", Step: "Setup code-review-graph", Status: integration.StatusDone},
	}
	m.integ.saveErr = errors.New("disk full")

	view := m.viewIntegrationSummary()
	if !strings.Contains(view, "not saved") {
		t.Errorf("expected prominent not-saved warning, view:\n%s", view)
	}
	if !strings.Contains(view, "disk full") {
		t.Errorf("expected save error detail, view:\n%s", view)
	}
}
