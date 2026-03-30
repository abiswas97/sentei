package tui

import (
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/repo"
)

func makeIntegrationModel() Model {
	m := NewMenuModel(nil, nil, "/repo", &config.Config{}, repo.ContextBareRepo)
	m.integ.integrations = integration.All()
	m.integ.current = map[string]bool{"code-review-graph": true, "cocoindex-code": false}
	m.integ.staged = map[string]bool{"code-review-graph": true, "cocoindex-code": false}
	m.integ.depStatus = map[string]bool{"python3.10+": true, "pipx": true, "python3.11+": true, "uv": true}
	m.width = 80
	m.height = 24
	return m
}

func TestUpdateIntegrationProgress_EventMsg(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	ch := make(chan integration.ManagerEvent, 1)
	doneCh := make(chan struct{}, 1)
	m.integ.eventCh = ch
	m.integ.doneCh = doneCh

	ev := integration.ManagerEvent{
		Worktree: "/repo/main",
		Step:     "Install code-review-graph",
		Status:   integration.StatusRunning,
	}
	updated, _ := m.updateIntegrationProgress(integrationEventMsg{Event: ev})
	m = updated.(Model)

	if len(m.integ.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(m.integ.events))
	}
	if m.integ.events[0].Step != "Install code-review-graph" {
		t.Errorf("unexpected event step: %q", m.integ.events[0].Step)
	}
}

func TestUpdateIntegrationProgress_FinalizedMsg(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.returnView = integrationListView

	newCurrent := map[string]bool{
		"code-review-graph": true,
		"cocoindex-code":    true,
	}
	updated, _ := m.updateIntegrationProgress(integrationFinalizedMsg{current: newCurrent, err: nil})
	m = updated.(Model)

	if m.view != integrationListView {
		t.Errorf("expected integrationListView, got %d", m.view)
	}
	if !m.integ.current["code-review-graph"] {
		t.Error("expected code-review-graph to be current")
	}
	if !m.integ.current["cocoindex-code"] {
		t.Error("expected cocoindex-code to be updated in current")
	}
}

func TestUpdateIntegrationProgress_FinalizedMsg_Migration(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.returnView = migrateNextView
	// current starts as code-review-graph=true, cocoindex-code=false
	originalCurrent := map[string]bool{
		"code-review-graph": true,
		"cocoindex-code":    false,
	}
	m.integ.current = originalCurrent

	newCurrent := map[string]bool{
		"code-review-graph": true,
		"cocoindex-code":    true,
	}
	updated, _ := m.updateIntegrationProgress(integrationFinalizedMsg{current: newCurrent, err: nil})
	m = updated.(Model)

	if m.view != migrateNextView {
		t.Errorf("expected migrateNextView, got %d", m.view)
	}
	// current should NOT be updated for migration flow
	if m.integ.current["cocoindex-code"] {
		t.Error("cocoindex-code should not be updated in current for migration flow")
	}
}

func TestViewIntegrationProgress_GroupsByWorktree(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/main", Step: "Install step", Status: integration.StatusDone},
		{Worktree: "/repo/feature", Step: "Install step", Status: integration.StatusRunning},
	}

	output := stripAnsi(m.viewIntegrationProgress())

	if !strings.Contains(output, "main") {
		t.Error("expected output to contain 'main' worktree name")
	}
	if !strings.Contains(output, "feature") {
		t.Error("expected output to contain 'feature' worktree name")
	}
}

func TestViewIntegrationProgress_ShowsProgressBar(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/main", Step: "step1", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "step2", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "step3", Status: integration.StatusRunning},
	}

	output := stripAnsi(m.viewIntegrationProgress())

	if !strings.Contains(output, "2/3") {
		t.Errorf("expected output to contain progress '2/3', got:\n%s", output)
	}
}

func TestViewIntegrationProgress_ProgressCountsUniqueSteps(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	// Each step emits Running then Done — 6 raw events but only 3 unique steps.
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/main", Step: "Setup code-review-graph", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "Setup code-review-graph", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "Install cocoindex-code", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "Install cocoindex-code", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "Setup cocoindex-code", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "Setup cocoindex-code", Status: integration.StatusDone},
	}

	output := stripAnsi(m.viewIntegrationProgress())

	// Should show 3/3 (3 unique steps, all done), not 6/6.
	if !strings.Contains(output, "3/3") {
		t.Errorf("expected progress '3/3' (unique steps), got:\n%s", output)
	}
}

func TestViewIntegrationProgress_Loading(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.events = nil

	output := stripAnsi(m.viewIntegrationProgress())

	if !strings.Contains(output, "Applying Integration Changes") {
		t.Errorf("expected title 'Applying Integration Changes', got:\n%s", output)
	}
}
