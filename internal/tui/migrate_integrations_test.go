package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/integration"
)

func TestUpdateMigrateIntegrations_DetectedMsg(t *testing.T) {
	m := makeIntegrationModel()
	m.view = migrateIntegrationsView

	detected := map[string]bool{
		"code-review-graph": true,
		"cocoindex-code":    false,
	}
	updated, _ := m.updateMigrateIntegrations(migrateIntegrationDetectedMsg{
		integrations: integration.All(),
		detected:     detected,
	})
	m = updated.(Model)

	if !m.integ.staged["code-review-graph"] {
		t.Error("expected code-review-graph to be pre-checked (detected=true)")
	}
	if m.integ.staged["cocoindex-code"] {
		t.Error("expected cocoindex-code to be unchecked (detected=false)")
	}
	if !m.integ.detected["code-review-graph"] {
		t.Error("expected detected map to be stored")
	}
}

func TestUpdateMigrateIntegrations_Toggle(t *testing.T) {
	m := makeIntegrationModel()
	m.view = migrateIntegrationsView
	m.integ.integrations = integration.All()
	m.integ.cursor = 0
	firstName := m.integ.integrations[0].Name
	initialState := m.integ.staged[firstName]

	updated, _ := m.updateMigrateIntegrations(keyMsg(" "))
	m = updated.(Model)

	if m.integ.staged[firstName] == initialState {
		t.Errorf("expected staged[%q] to be toggled from %v", firstName, initialState)
	}
}

func TestUpdateMigrateIntegrations_ConfirmNoSelections(t *testing.T) {
	m := makeIntegrationModel()
	m.view = migrateIntegrationsView
	m.integ.integrations = integration.All()
	// Set all staged to false
	for _, integ := range m.integ.integrations {
		m.integ.staged[integ.Name] = false
	}

	updated, _ := m.updateMigrateIntegrations(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.view != migrateNextView {
		t.Errorf("expected migrateNextView when no selections, got %d", m.view)
	}
}

func TestUpdateMigrateIntegrations_Skip(t *testing.T) {
	m := makeIntegrationModel()
	m.view = migrateIntegrationsView
	m.integ.integrations = integration.All()

	updated, _ := m.updateMigrateIntegrations(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.view != migrateNextView {
		t.Errorf("expected migrateNextView after esc, got %d", m.view)
	}
}

func TestViewMigrateIntegrations_ShowsDetectedHint(t *testing.T) {
	m := makeIntegrationModel()
	m.view = migrateIntegrationsView
	m.integ.integrations = integration.All()
	m.integ.detected = map[string]bool{
		"code-review-graph": true,
		"cocoindex-code":    false,
	}

	output := stripAnsi(m.viewMigrateIntegrations())

	if !strings.Contains(output, "detected") {
		t.Errorf("expected output to contain 'detected' hint, got:\n%s", output)
	}
}

func TestViewMigrateIntegrations_ShowsIntroText(t *testing.T) {
	m := makeIntegrationModel()
	m.view = migrateIntegrationsView
	m.integ.integrations = integration.All()

	output := stripAnsi(m.viewMigrateIntegrations())

	if !strings.Contains(output, "dev tools") {
		t.Errorf("expected output to contain 'dev tools', got:\n%s", output)
	}
}
