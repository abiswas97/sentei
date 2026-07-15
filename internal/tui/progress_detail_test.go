package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/abiswas97/sentei/internal/progress"
)

func progressDetailModel() Model {
	m := NewModel(nil, nil, "/repo")
	m.view = progressView
	m.width, m.height, m.windowHeight = 50, 18, 24
	m.portal = m.portal.SetSize(50, 24)
	m.remove.run.events = []progress.Event{
		{Phase: "phase-id", PhaseLabel: "Readable phase", Step: "skip-id", StepLabel: "界 skipped step", Status: progress.StepPending, Of: 1},
		{Phase: "phase-id", PhaseLabel: "Readable phase", Step: "fail-id", StepLabel: "failed step", Status: progress.StepPending, Of: 1},
		{Phase: "phase-id", PhaseLabel: "Readable phase", Close: true},
		{Phase: "phase-id", Step: "skip-id", Status: progress.StepSkipped, Message: "already installed with a very long reason"},
		{Phase: "phase-id", Step: "fail-id", Status: progress.StepFailed, Error: errors.New("\x1b[31mfull failure detail\x1b[0m")},
	}
	return m
}

func TestProgressDetail_IncludesStructuredFullText(t *testing.T) {
	m := progressDetailModel()
	title, content := m.detailContent()
	plain := stripANSI(content)
	if title != "Progress details" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{"phase-id", "Readable phase", "skip-id", "界 skipped step", "skipped", "already installed with a very long reason", "fail-id", "failed", "full failure detail"} {
		if !strings.Contains(plain, want) {
			t.Errorf("detail missing %q:\n%s", want, plain)
		}
	}
	for i, line := range strings.Split(m.viewProgress(), "\n") {
		if width := lipgloss.Width(line); width > 50 {
			t.Fatalf("main line %d width = %d", i+1, width)
		}
	}
}

func TestProgressDetail_OfferedOnlyForFailureOmissionOrTopError(t *testing.T) {
	m := NewModel(nil, nil, "/repo")
	m.view = progressView
	m.width, m.height, m.windowHeight = 80, 18, 24
	m.remove.run.events = []progress.Event{{Phase: "p", PhaseLabel: "P", Step: "s", StepLabel: "S", Status: progress.StepPending, Of: 1}, {Phase: "p", Close: true}}
	if _, content := m.detailContent(); content != "" {
		t.Fatalf("simple pending progress unexpectedly offered details: %q", content)
	}
	m.remove.run.result.Err = errors.New("delivery failed")
	if _, content := m.detailContent(); !strings.Contains(content, "delivery failed") {
		t.Fatalf("top-level error did not offer details: %q", content)
	}
	if footer := stripANSI(viewFooter(80, m.removalLayout().Hints)); !strings.Contains(footer, "? details") {
		t.Fatalf("details hint missing: %q", footer)
	}
}

func TestProgressPortal_StaysOpenAcrossBackgroundEventAndResize(t *testing.T) {
	m := progressDetailModel()
	title, content := m.detailContent()
	m.portal = m.portal.Open(portalDetails, title, content)
	updated, _ := m.Update(removalEventMsg{event: progress.Event{Phase: "phase-id", Step: "skip-id", Status: progress.StepSkipped, Message: "updated"}})
	m = updated.(Model)
	if !m.portal.Visible() {
		t.Fatal("background progress event closed portal")
	}
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = updated.(Model)
	if !m.portal.Visible() || m.portal.viewport.Width() != m.portal.contentWidth() {
		t.Fatal("resize did not preserve/refit portal")
	}
}
