package tui

import (
	"strings"
	"testing"
)

func TestDetailContent_IntegrationList_AllSections(t *testing.T) {
	m := makeIntegrationModel()
	m.portal = m.portal.SetSize(m.width, m.height)
	m.view = integrationListView

	title, content := m.detailContent()
	if title != "Integration details" {
		t.Errorf("title = %q, want Integration details", title)
	}
	plain := stripAnsi(content)
	for _, integ := range m.integ.integrations {
		if !strings.Contains(plain, integ.Name) {
			t.Errorf("portal page missing section for %q:\n%s", integ.Name, plain)
		}
	}
	if !strings.Contains(plain, "installed") {
		t.Errorf("portal page missing dependency status rows:\n%s", plain)
	}
}

func TestDetailContent_MigrateIntegrations_SamePage(t *testing.T) {
	m := makeIntegrationModel()
	m.view = migrateIntegrationsView

	title, content := m.detailContent()
	if title != "Integration details" || content == "" {
		t.Errorf("migrate-integrations view must serve the same portal page, got title %q", title)
	}
}

func TestDetailContent_NoIntegrations_Empty(t *testing.T) {
	m := makeIntegrationModel()
	m.view = integrationListView
	m.integ.integrations = nil

	if _, content := m.detailContent(); content != "" {
		t.Errorf("no integrations must yield no portal page, got %q", content)
	}
}

func TestRenderIntegrationsDetail_WrapsToPortalWidth(t *testing.T) {
	m := makeIntegrationModel()
	m.portal = m.portal.SetSize(80, 24)

	maxWidth := m.portal.contentWidth()
	for i, line := range strings.Split(stripAnsi(m.renderIntegrationsDetail()), "\n") {
		if n := len([]rune(line)); n > maxWidth {
			t.Errorf("line %d exceeds portal width (%d > %d): %q", i, n, maxWidth, line)
		}
	}
}

func TestPortal_HorizontalKeysInert(t *testing.T) {
	m := makeIntegrationModel()
	m.portal = m.portal.SetSize(80, 24)
	m.portal = m.portal.Open(portalDetails, "T", strings.Repeat("x", 300))

	before := m.portal.View(strings.Repeat(" ", 80))
	updated, _ := m.updatePortalKeys(keyMsg("l"))
	after := updated.(Model).portal.View(strings.Repeat(" ", 80))
	if before != after {
		t.Error("l must not horizontally scroll the portal; it scrolls vertically only")
	}
}
