package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
)

func doneEventsForWorktrees(n int) []integration.ManagerEvent {
	var events []integration.ManagerEvent
	for i := 0; i < n; i++ {
		events = append(events, integration.ManagerEvent{
			Worktree: fmt.Sprintf("/repo/feature-%d", i),
			Step:     "Setup code-review-graph",
			Status:   integration.StatusDone,
		})
	}
	return events
}

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

func TestViewIntegrationSummary_OverflowPeeksAndOffersDetails(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationSummaryView
	m.width, m.height = 80, 40
	m.integ.events = doneEventsForWorktrees(5)

	view := stripAnsi(m.viewIntegrationSummary())

	for _, want := range []string{"feature-0", "feature-1", "feature-2"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected peek to show %q, view:\n%s", want, view)
		}
	}
	for _, hidden := range []string{"feature-3", "feature-4"} {
		if strings.Contains(view, hidden) {
			t.Errorf("expected %q to be hidden behind the portal, view:\n%s", hidden, view)
		}
	}
	if !strings.Contains(view, "and 2 more") {
		t.Errorf("expected an 'and N more' overflow line, view:\n%s", view)
	}
	if !strings.Contains(view, "details") {
		t.Errorf("expected a `?` details hint when outcomes overflow, view:\n%s", view)
	}
}

func TestIntegrationSummaryDetailContent(t *testing.T) {
	tests := []struct {
		name        string
		worktrees   int
		wantContent bool
	}{
		{"within peek has no portal", inlineSummaryPreview, false},
		{"overflow exposes full breakdown", inlineSummaryPreview + 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeIntegrationModel()
			m.view = integrationSummaryView
			m.width, m.height = 80, 40
			m.integ.events = doneEventsForWorktrees(tt.worktrees)

			title, content := m.detailContent()
			if tt.wantContent {
				if content == "" {
					t.Fatal("expected portal content when outcomes overflow the peek")
				}
				if title != "Apply Details" {
					t.Errorf("title = %q, want %q", title, "Apply Details")
				}
				for i := 0; i < tt.worktrees; i++ {
					if !strings.Contains(stripAnsi(content), fmt.Sprintf("feature-%d", i)) {
						t.Errorf("portal must list every worktree; missing feature-%d:\n%s", i, content)
					}
				}
			} else if content != "" {
				t.Errorf("expected no portal content within the peek, got:\n%s", content)
			}
		})
	}
}
