package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/testutil/mock"
)

func TestUpdateIntegrationList_LoadedMsg_SetsState(t *testing.T) {
	m := makeIntegrationModel()
	all := integration.All()
	depStatus := map[string]bool{"python3.10+": true, "pipx": false}
	current := map[string]bool{"code-review-graph": false, "cocoindex-code": true}

	msg := integrationStateLoadedMsg{
		integrations: all,
		current:      current,
		depStatus:    depStatus,
		enabled:      nil,
	}

	updated, _ := m.updateIntegrationList(msg)
	m = updated.(Model)

	if len(m.integ.integrations) != len(all) {
		t.Errorf("expected %d integrations, got %d", len(all), len(m.integ.integrations))
	}
	if m.integ.current["code-review-graph"] != false {
		t.Error("expected current[code-review-graph]=false")
	}
	if m.integ.current["cocoindex-code"] != true {
		t.Error("expected current[cocoindex-code]=true")
	}
	if m.integ.depStatus["python3.10+"] != true {
		t.Error("expected depStatus[python3.10+]=true")
	}
	if m.integ.depStatus["pipx"] != false {
		t.Error("expected depStatus[pipx]=false")
	}
	// staged must mirror current when enabled is nil
	for _, integ := range all {
		if m.integ.staged[integ.Name] != current[integ.Name] {
			t.Errorf("staged[%s] should equal current[%s]", integ.Name, integ.Name)
		}
	}
}

func TestUpdateIntegrationList_LoadedMsg_OverlaysEnabled(t *testing.T) {
	m := makeIntegrationModel()
	all := integration.All()
	current := map[string]bool{"code-review-graph": true, "cocoindex-code": false}

	msg := integrationStateLoadedMsg{
		integrations: all,
		current:      current,
		depStatus:    map[string]bool{},
		enabled:      []string{"cocoindex-code"},
	}

	updated, _ := m.updateIntegrationList(msg)
	m = updated.(Model)

	if !m.integ.staged["cocoindex-code"] {
		t.Error("staged[cocoindex-code] should be true (enabled overlay)")
	}
	if m.integ.staged["code-review-graph"] {
		t.Error("staged[code-review-graph] should be false (enabled overrides disk)")
	}
}

func TestUpdateIntegrationList_LoadedMsg_CorruptState(t *testing.T) {
	m := makeIntegrationModel()
	all := integration.All()
	current := map[string]bool{"code-review-graph": true, "cocoindex-code": false}

	msg := integrationStateLoadedMsg{
		integrations: all,
		current:      current,
		depStatus:    map[string]bool{},
		err:          fmt.Errorf("corrupt"),
	}

	updated, _ := m.updateIntegrationList(msg)
	m = updated.(Model)

	// When err is set, staged should fall back to current (enabled overlay not applied)
	for _, integ := range all {
		if m.integ.staged[integ.Name] != current[integ.Name] {
			t.Errorf("on corrupt state, staged[%s] should equal current[%s]", integ.Name, integ.Name)
		}
	}
}

func TestUpdateIntegrationList_Toggle(t *testing.T) {
	m := makeIntegrationModel()
	// cursor=0 → code-review-graph, currently staged=true
	m.integ.cursor = 0

	before := m.integ.staged[m.integ.integrations[0].Name]

	updated, _ := m.updateIntegrationList(keyMsg(" "))
	m = updated.(Model)

	after := m.integ.staged[m.integ.integrations[0].Name]
	if after == before {
		t.Errorf("toggle should flip staged value: was %v, still %v", before, after)
	}
}

func TestUpdateIntegrationList_Navigate(t *testing.T) {
	m := makeIntegrationModel()
	m.integ.cursor = 0

	// Move down
	updated, _ := m.updateIntegrationList(keyMsg("j"))
	m = updated.(Model)
	if m.integ.cursor != 1 {
		t.Errorf("cursor should be 1 after j, got %d", m.integ.cursor)
	}

	// Move up
	updated, _ = m.updateIntegrationList(keyMsg("k"))
	m = updated.(Model)
	if m.integ.cursor != 0 {
		t.Errorf("cursor should be 0 after k, got %d", m.integ.cursor)
	}

	// Clamping at top
	updated, _ = m.updateIntegrationList(keyMsg("k"))
	m = updated.(Model)
	if m.integ.cursor != 0 {
		t.Errorf("cursor should clamp at 0, got %d", m.integ.cursor)
	}

	// Clamping at bottom
	m.integ.cursor = len(m.integ.integrations) - 1
	updated, _ = m.updateIntegrationList(keyMsg("j"))
	m = updated.(Model)
	if m.integ.cursor != len(m.integ.integrations)-1 {
		t.Errorf("cursor should clamp at %d, got %d", len(m.integ.integrations)-1, m.integ.cursor)
	}
}

func TestUpdateIntegrationList_Back(t *testing.T) {
	m := makeIntegrationModel()
	// Stage a change so staged != current
	m.integ.staged["cocoindex-code"] = true // current is false

	escMsg := tea.KeyPressMsg{Code: tea.KeyEsc}
	updated, _ := m.updateIntegrationList(escMsg)
	m = updated.(Model)

	// staged should be reset to current
	for _, integ := range m.integ.integrations {
		if m.integ.staged[integ.Name] != m.integ.current[integ.Name] {
			t.Errorf("staged[%s] should be reset to current on esc back", integ.Name)
		}
	}
	if m.view != menuView {
		t.Errorf("view should be menuView after esc back, got %v", m.view)
	}
}

func TestIntegrationHasPendingChanges(t *testing.T) {
	m := makeIntegrationModel()

	// staged == current → no pending changes
	if m.integrationHasPendingChanges() {
		t.Error("expected no pending changes when staged == current")
	}

	// Flip cocoindex-code: staged=true, current=false → pending
	m.integ.staged["cocoindex-code"] = true
	if !m.integrationHasPendingChanges() {
		t.Error("expected pending changes after staging cocoindex-code=true")
	}
}

func TestPendingChangeCount(t *testing.T) {
	m := makeIntegrationModel()

	// No changes
	if count := m.pendingChangeCount(); count != 0 {
		t.Errorf("expected 0 pending changes, got %d", count)
	}

	// Change both: cocoindex-code true (current=false), code-review-graph false (current=true)
	m.integ.staged["cocoindex-code"] = true
	m.integ.staged["code-review-graph"] = false

	if count := m.pendingChangeCount(); count != 2 {
		t.Errorf("expected 2 pending changes, got %d", count)
	}
}

func TestViewIntegrationList_ShowsIntegrations(t *testing.T) {
	m := makeIntegrationModel()
	out := stripAnsi(m.viewIntegrationList())

	for _, integ := range m.integ.integrations {
		if !strings.Contains(out, integ.Name) {
			t.Errorf("view should contain integration name %q", integ.Name)
		}
		if !strings.Contains(out, integ.ShortDescription) {
			t.Errorf("view should contain short description for %q", integ.Name)
		}
	}
}

func TestViewIntegrationList_CheckboxStates(t *testing.T) {
	m := makeIntegrationModel()
	// cocoindex-code: staged=true, current=false → [+]
	m.integ.staged["cocoindex-code"] = true
	m.integ.current["cocoindex-code"] = false

	out := stripAnsi(m.viewIntegrationList())
	if !strings.Contains(out, "[+]") {
		t.Errorf("view should show [+] for newly staged addition, got:\n%s", out)
	}
}

func TestViewIntegrationList_PendingCount(t *testing.T) {
	m := makeIntegrationModel()
	m.integ.staged["cocoindex-code"] = true // current=false → 1 change

	out := stripAnsi(m.viewIntegrationList())
	if !strings.Contains(out, "1 change pending") {
		t.Errorf("view should show '1 change pending', got:\n%s", out)
	}
}

func TestViewIntegrationList_Legend(t *testing.T) {
	m := makeIntegrationModel()
	out := stripAnsi(m.viewIntegrationList())

	for _, want := range []string{"active", "inactive", "adding", "removing"} {
		if !strings.Contains(out, want) {
			t.Errorf("legend should contain %q", want)
		}
	}
}

func TestRenderIntegrationsDetail(t *testing.T) {
	m := makeIntegrationModel()
	m.portal = m.portal.SetSize(m.width, m.height)

	out := stripAnsi(m.renderIntegrationsDetail())
	for _, integ := range m.integ.integrations {
		if !strings.Contains(out, integ.Name) {
			t.Errorf("detail page should contain integration name %q", integ.Name)
		}
		if !strings.Contains(out, integ.Description[:30]) {
			t.Errorf("detail page should contain description prefix for %q", integ.Name)
		}
		if integ.URL != "" && !strings.Contains(out, integ.URL) {
			t.Errorf("detail page should contain URL %q", integ.URL)
		}
		if len(integ.Dependencies) > 0 && !strings.Contains(out, "Dependencies") {
			t.Errorf("detail page should contain 'Dependencies' header")
		}
	}
}

func TestRenderIntegrationsDetail_DepStatus(t *testing.T) {
	m := makeIntegrationModel()
	m.portal = m.portal.SetSize(m.width, m.height)

	m.integ.depStatus["python3.10+"] = true
	out := stripAnsi(m.renderIntegrationsDetail())
	if !strings.Contains(out, "installed") {
		t.Errorf("detail page should show 'installed' for present dep, got:\n%s", out)
	}

	m.integ.depStatus["pipx"] = false
	out = stripAnsi(m.renderIntegrationsDetail())
	if !strings.Contains(out, "will be installed") {
		t.Errorf("detail page should show 'will be installed' for missing dep, got:\n%s", out)
	}
}

func TestStartIntegrationApply_ComputesPlanAndAppliesChanges(t *testing.T) {
	wtPath := t.TempDir()
	m := makeIntegrationModel()
	m.remove.worktrees = []git.Worktree{{Path: wtPath, Branch: "refs/heads/main"}}
	m.integ.integrations = []integration.Integration{
		{
			Name:         "enable-me",
			Dependencies: []integration.Dependency{{Name: "dep1", Detect: "dep1 --version"}},
			Detect:       integration.DetectSpec{Command: "enable-me --version"},
			Install:      integration.InstallSpec{Command: "install enable-me"},
			Setup:        integration.SetupSpec{Command: "enable-me init"},
		},
		{
			Name:     "disable-me",
			Teardown: integration.TeardownSpec{Command: "disable-me clean", Dirs: []string{".disable-a/", ".disable-b/"}},
		},
	}
	m.integ.current = map[string]bool{"enable-me": false, "disable-me": true}
	m.integ.staged = map[string]bool{"enable-me": true, "disable-me": false}
	m.shell = &mock.Runner{Responses: map[string]mock.Response{
		wtPath + ":shell[enable-me --version]": {Output: "1.0"}, // already installed
		wtPath + ":shell[enable-me init]":      {Output: "ok"},
		wtPath + ":shell[disable-me clean]":    {Output: "ok"},
	}}

	updated, cmd := m.startIntegrationApply()

	// enable: 1 setup + 1 dep + 1 install; disable: 1 teardown + 2 dir removals.
	if updated.integ.totalSteps != 6 {
		t.Errorf("totalSteps = %d, want 6", updated.integ.totalSteps)
	}
	if len(updated.integ.targetWorktrees) != 1 || updated.integ.targetWorktrees[0] != wtPath {
		t.Errorf("targetWorktrees = %v, want [%s]", updated.integ.targetWorktrees, wtPath)
	}
	if cmd == nil {
		t.Fatal("expected a wait command")
	}

	events := drainIntegrationApply(t, updated)
	var sawSetup, sawTeardown bool
	for _, ev := range events {
		if ev.Step == "Setup enable-me" && ev.Status == integration.StatusDone {
			sawSetup = true
		}
		if ev.Step == "Teardown disable-me" && ev.Status == integration.StatusDone {
			sawTeardown = true
		}
	}
	if !sawSetup {
		t.Errorf("expected enable-me setup to run, events: %v", events)
	}
	if !sawTeardown {
		t.Errorf("expected disable-me teardown to run, events: %v", events)
	}
}
