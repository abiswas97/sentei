package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/pipeline"
	"github.com/abiswas97/sentei/internal/testutil/mock"
)

func runningLayout() ProgressLayout {
	return ProgressLayout{
		Title: "T", Width: 80, Height: 30,
		Phases: []phaseDisplay{
			{name: "Removing worktrees", total: 3, done: 1, steps: []stepDisplay{
				{name: "done-step", status: pipeline.StepDone},
				{name: "active-step", status: pipeline.StepRunning},
				{name: "pending-step", status: pipeline.StepPending},
			}},
		},
	}
}

func TestProgressLayout_StaticFallbackUsesMidpointDot(t *testing.T) {
	view := stripANSI(runningLayout().View())

	if strings.Contains(view, "◐") {
		t.Errorf("static layouts must not render the retired ◐, view:\n%s", view)
	}
	if !strings.Contains(view, "∙ Removing worktrees") {
		t.Errorf("expected midpoint-dot fallback on the active phase, view:\n%s", view)
	}
	if !strings.Contains(view, "    ∙ active-step") {
		t.Errorf("expected midpoint-dot fallback on the running step, view:\n%s", view)
	}
}

func TestRenderProgressLayout_InjectsBreathFrame(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView

	out := stripAnsi(m.renderProgressLayout(runningLayout()))
	if !strings.Contains(out, "· Removing worktrees") {
		t.Errorf("expected initial breath frame · on the active phase, view:\n%s", out)
	}

	for range 2 {
		updated, _ := m.Update(spinner.TickMsg{ID: m.breath.ID()})
		m = updated.(Model)
	}
	out = stripAnsi(m.renderProgressLayout(runningLayout()))
	if !strings.Contains(out, "● Removing worktrees") {
		t.Errorf("expected third breath frame ● after two ticks, view:\n%s", out)
	}
}

func TestBreathTicks_GatedOutsideProgressViews(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = listView

	if _, cmd := m.Update(spinner.TickMsg{ID: m.breath.ID()}); cmd != nil {
		t.Error("breath ticks outside progress views must be swallowed")
	}

	m.view = progressView
	if _, cmd := m.Update(spinner.TickMsg{ID: m.breath.ID()}); cmd == nil {
		t.Error("breath ticks in a progress view must continue the animation")
	}
}

func TestBreathTick_StartsOnFlowEntry(t *testing.T) {
	m := NewModel([]git.Worktree{{Path: "/work/a", Branch: "refs/heads/a"}}, &mock.Runner{}, "/repo")
	m.view = confirmView
	m.remove.selected = map[string]bool{"/work/a": true}

	updated, cmd := m.Update(keyMsg("y"))
	model := updated.(Model)
	if model.view != progressView {
		t.Fatalf("expected progressView after confirm, got %d", model.view)
	}
	if !containsBreathTick(cmd, model.breath.ID()) {
		t.Error("entering a progress view must start the breath tick chain")
	}
}

// containsBreathTick walks a command tree looking for the breath spinner's
// own TickMsg, expanding batches without executing flow side effects twice.
func containsBreathTick(cmd tea.Cmd, id int) bool {
	if cmd == nil {
		return false
	}
	switch msg := cmd().(type) {
	case spinner.TickMsg:
		return msg.ID == id
	case tea.BatchMsg:
		for _, sub := range msg {
			if containsBreathTick(sub, id) {
				return true
			}
		}
	}
	return false
}

func TestCleanupRunningLine_UsesBreathFrame(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = cleanupResultView

	out := stripAnsi(m.viewCleanupResult())
	if strings.Contains(out, "◐") {
		t.Errorf("cleanup running line must not use the retired ◐, view:\n%s", out)
	}
	if !strings.Contains(out, "· Running cleanup…") {
		t.Errorf("expected breath frame on the running line, view:\n%s", out)
	}
}

func TestViewStatLine_UsesProvidedActiveGlyph(t *testing.T) {
	line := stripAnsi(viewStatLine(WindowStats{Done: 1, Active: 2, Pending: 3, Showing: 4, Total: 6}, "●"))
	if !strings.Contains(line, "● 2 active") {
		t.Errorf("stat line must render the provided glyph, got %q", line)
	}
}

func TestOverallBar_FillsWidthMinusMeta(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	if got := m.bar.Width(); got != 120-2-progressBarElapsedReserve {
		t.Errorf("bar width = %d, want %d", got, 120-2-progressBarElapsedReserve)
	}

	updated, _ = m.Update(tea.WindowSizeMsg{Width: 30, Height: 40})
	m = updated.(Model)
	if got := m.bar.Width(); got != minProgressBarWidth {
		t.Errorf("bar width below floor: %d, want %d", got, minProgressBarWidth)
	}
}

func TestStaticBar_FollowsLayoutWidth(t *testing.T) {
	l := runningLayout()
	l.Width = 120
	wide := stripANSI(l.View())
	l.Width = 60
	narrow := stripANSI(l.View())

	wideBar := barLineLength(t, wide)
	narrowBar := barLineLength(t, narrow)
	if wideBar <= narrowBar {
		t.Errorf("static bar must widen with the layout: %d (120 cols) vs %d (60 cols)", wideBar, narrowBar)
	}
}

func barLineLength(t *testing.T, view string) int {
	t.Helper()
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "█") || strings.Contains(line, "░") {
			return len([]rune(strings.TrimRight(line, " ")))
		}
	}
	t.Fatalf("no bar line in view:\n%s", view)
	return 0
}

func TestOverallBar_GradientSpansFill(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")

	full := m.bar.ViewAs(1.0)
	cells := strings.SplitAfter(full, "█")
	if len(cells) < 3 {
		t.Fatalf("expected multiple filled cells, got %q", full)
	}
	if !strings.Contains(full, "38;2;") {
		t.Errorf("expected truecolor gradient sequences in the fill, got %q", full)
	}
	first := ansiPrefix(cells[0])
	last := ansiPrefix(cells[len(cells)-2])
	if first == last {
		t.Errorf("gradient endpoints must differ: first cell %q == last cell %q", first, last)
	}
}

// ansiPrefix returns the escape sequence immediately preceding the cell's rune.
func ansiPrefix(cell string) string {
	if i := strings.LastIndex(cell, "\x1b["); i >= 0 {
		return cell[i:]
	}
	return ""
}

func TestPalettes_DefineGradientTokens(t *testing.T) {
	for name, p := range map[string]palette{"dark": darkPalette, "light": lightPalette} {
		if p.barStart == nil || p.barEnd == nil {
			t.Errorf("%s palette missing gradient tokens", name)
		}
	}
}
