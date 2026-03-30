package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
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

func TestUpdateIntegrationList_InfoOpen(t *testing.T) {
	m := makeIntegrationModel()
	m.integ.cursor = 1

	updated, _ := m.updateIntegrationList(keyMsg("?"))
	m = updated.(Model)

	if !m.integ.showInfo {
		t.Error("showInfo should be true after ?")
	}
	if m.integ.infoCursor != 1 {
		t.Errorf("infoCursor should match cursor=1, got %d", m.integ.infoCursor)
	}
}

func TestUpdateIntegrationList_InfoClose(t *testing.T) {
	m := makeIntegrationModel()
	m.integ.showInfo = true
	m.integ.infoCursor = 0

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := m.updateIntegrationList(escMsg)
	m = updated.(Model)

	if m.integ.showInfo {
		t.Error("showInfo should be false after esc")
	}
}

func TestUpdateIntegrationList_InfoCarousel(t *testing.T) {
	m := makeIntegrationModel()
	m.integ.showInfo = true
	m.integ.infoCursor = 0
	total := len(m.integ.integrations)

	// Move right
	updated, _ := m.updateIntegrationList(keyMsg("l"))
	m = updated.(Model)
	if m.integ.infoCursor != 1 {
		t.Errorf("infoCursor should be 1 after l, got %d", m.integ.infoCursor)
	}

	// Wrap around at end
	m.integ.infoCursor = total - 1
	updated, _ = m.updateIntegrationList(keyMsg("l"))
	m = updated.(Model)
	if m.integ.infoCursor != 0 {
		t.Errorf("infoCursor should wrap to 0, got %d", m.integ.infoCursor)
	}
}

func TestUpdateIntegrationList_Back(t *testing.T) {
	m := makeIntegrationModel()
	// Stage a change so staged != current
	m.integ.staged["cocoindex-code"] = true // current is false

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
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

func TestRenderIntegrationInfo(t *testing.T) {
	m := makeIntegrationModel()
	m.integ.showInfo = true
	m.integ.infoCursor = 0

	integ := m.integ.integrations[0]
	out := stripAnsi(m.renderIntegrationInfo())

	if !strings.Contains(out, integ.Name) {
		t.Errorf("info should contain integration name %q", integ.Name)
	}
	// Description is word-wrapped by lipgloss; check for a meaningful prefix
	descPrefix := integ.Description[:30]
	if !strings.Contains(out, descPrefix) {
		t.Errorf("info should contain description prefix %q", descPrefix)
	}
	if integ.URL != "" && !strings.Contains(out, integ.URL) {
		t.Errorf("info should contain URL %q", integ.URL)
	}
	if len(integ.Dependencies) > 0 && !strings.Contains(out, "Dependencies") {
		t.Errorf("info should contain 'Dependencies' header")
	}
}

func TestRenderIntegrationInfo_DepStatus(t *testing.T) {
	m := makeIntegrationModel()
	m.integ.showInfo = true
	m.integ.infoCursor = 0 // code-review-graph has python3.10+ and pipx deps

	// python3.10+ is installed
	m.integ.depStatus["python3.10+"] = true
	out := stripAnsi(m.renderIntegrationInfo())
	if !strings.Contains(out, "installed") {
		t.Errorf("info should show 'installed' for present dep, got:\n%s", out)
	}

	// pipx is not installed
	m.integ.depStatus["pipx"] = false
	out = stripAnsi(m.renderIntegrationInfo())
	if !strings.Contains(out, "will be installed") {
		t.Errorf("info should show 'will be installed' for missing dep, got:\n%s", out)
	}
}
