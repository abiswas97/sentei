package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/config"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/state"
)

func TestUpdateIntegrationProgress_FinalizedMsg_SaveError_DoesNotApply(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.returnView = integrationListView
	genBefore := m.worktreeGeneration

	// The apply tried to flip cocoindex-code on, but state.Save failed.
	updated, _ := m.updateIntegrationProgress(integrationFinalizedMsg{err: errors.New("save failed")})
	m = updated.(Model)

	// In-memory state must stay consistent with disk: the unsaved change is not applied.
	if m.integ.current["cocoindex-code"] {
		t.Error("a failed save must not mutate in-memory current state")
	}
	if m.integ.staged["cocoindex-code"] {
		t.Error("a failed save must not mutate staged state")
	}
	// worktreeGeneration must not advance (no reload of unsaved state).
	if m.worktreeGeneration != genBefore {
		t.Errorf("worktreeGeneration must not advance on a failed save: %d -> %d", genBefore, m.worktreeGeneration)
	}
}

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

func TestUpdateIntegrationProgress_FinalizedMsg_DoesNotMutateInMemory(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.returnView = integrationListView

	updated, _ := m.updateIntegrationProgress(integrationFinalizedMsg{err: nil})
	m = updated.(Model)

	if m.view != integrationSummaryView {
		t.Errorf("expected integrationSummaryView, got %d", m.view)
	}
	// current/staged are reconciled from disk when the summary is dismissed,
	// never mutated in-memory at finalize time.
	if m.integ.current["cocoindex-code"] {
		t.Error("finalize must not mutate in-memory current state")
	}
	if m.integ.staged["cocoindex-code"] {
		t.Error("finalize must not mutate staged state")
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

	updated, _ := m.updateIntegrationProgress(integrationFinalizedMsg{err: nil})
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
	m.integ.totalSteps = 3 // Known upfront.
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/main", Step: "step1", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "step1", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "step2", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "step2", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "step3", Status: integration.StatusRunning},
	}

	output := stripAnsi(m.viewIntegrationProgress())

	// 2 done out of 3 total → 66%.
	if !strings.Contains(output, "66%") {
		t.Errorf("expected output to contain progress '66%%', got:\n%s", output)
	}
}

func TestViewIntegrationProgress_ProgressCountsUniqueSteps(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.totalSteps = 3
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/main", Step: "Setup code-review-graph", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "Setup code-review-graph", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "Install cocoindex-code", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "Install cocoindex-code", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "Setup cocoindex-code", Status: integration.StatusRunning},
		{Worktree: "/repo/main", Step: "Setup cocoindex-code", Status: integration.StatusDone},
	}

	output := stripAnsi(m.viewIntegrationProgress())

	// 3/3 done → 100%.
	if !strings.Contains(output, "100%") {
		t.Errorf("expected progress '100%%', got:\n%s", output)
	}
}

func TestViewIntegrationProgress_TotalKnownUpfront(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationProgressView
	m.integ.totalSteps = 9
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/main", Step: "Setup crg", Status: integration.StatusDone},
		{Worktree: "/repo/main", Step: "Install ccc", Status: integration.StatusRunning},
	}

	output := stripAnsi(m.viewIntegrationProgress())

	// 1 done out of 9 → 11%.
	if !strings.Contains(output, "11%") {
		t.Errorf("expected progress '11%%' (upfront total), got:\n%s", output)
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

func TestFinalizeIntegrationApply_SavesStagedIntegrations(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, ".bare"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := makeIntegrationModel()
	m.repoPath = tmp
	m.integ.returnView = integrationListView
	m.integ.staged = map[string]bool{"code-review-graph": true, "cocoindex-code": false}

	msg := m.finalizeIntegrationApply()()

	finalized, ok := msg.(integrationFinalizedMsg)
	if !ok {
		t.Fatalf("expected integrationFinalizedMsg, got %T", msg)
	}
	if finalized.err != nil {
		t.Fatalf("unexpected save error: %v", finalized.err)
	}

	st, err := state.Load(filepath.Join(tmp, ".bare"))
	if err != nil {
		t.Fatal(err)
	}
	if !st.HasIntegration("code-review-graph") || st.HasIntegration("cocoindex-code") {
		t.Errorf("persisted integrations = %v, want only code-review-graph", st.Integrations)
	}
}

func TestFinalizeIntegrationApply_MigrateSavesUnderBareRoot(t *testing.T) {
	bareRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(bareRoot, ".bare"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := makeIntegrationModel()
	m.repoPath = "/elsewhere"
	m.integ.returnView = migrateNextView
	m.repo.result = repo.MigrateResult{BareRoot: bareRoot, Branch: "main"}
	m.integ.staged = map[string]bool{"code-review-graph": true}

	msg := m.finalizeIntegrationApply()()

	if err := msg.(integrationFinalizedMsg).err; err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	st, err := state.Load(filepath.Join(bareRoot, ".bare"))
	if err != nil {
		t.Fatal(err)
	}
	if !st.HasIntegration("code-review-graph") {
		t.Errorf("persisted integrations = %v, want code-review-graph under the migrated bare root", st.Integrations)
	}
}

func TestFinalizeIntegrationApply_SaveFailureReturnsError(t *testing.T) {
	m := makeIntegrationModel()
	m.repoPath = filepath.Join(t.TempDir(), "missing") // .bare dir does not exist
	m.integ.returnView = integrationListView

	msg := m.finalizeIntegrationApply()()

	if msg.(integrationFinalizedMsg).err == nil {
		t.Error("expected a save error when the .bare dir is missing")
	}
}
