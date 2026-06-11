package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/progress"

	"github.com/abiswas97/sentei/internal/git"
)

func TestProgressLayout_Overall(t *testing.T) {
	cases := []struct {
		name                string
		layout              ProgressLayout
		wantDone, wantTotal int
	}{
		{"override wins", ProgressLayout{OverallDone: 3, OverallTotal: 10, Phases: []phaseDisplay{{done: 1, total: 1}}}, 3, 10},
		{"discovered phases summed", ProgressLayout{Phases: []phaseDisplay{{done: 2, total: 4}, {done: 1, total: 2}}}, 3, 6},
		{"undiscovered phase counts as outstanding", ProgressLayout{Phases: []phaseDisplay{{done: 2, total: 2}, {total: 0}}}, 2, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			done, total := tc.layout.overall()
			if done != tc.wantDone || total != tc.wantTotal {
				t.Errorf("overall() = %d/%d, want %d/%d", done, total, tc.wantDone, tc.wantTotal)
			}
		})
	}
}

func TestProgressFrame_GatedToProgressViews(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")

	cmd := m.bar.SetPercent(0.5)
	frame, ok := cmd().(progress.FrameMsg)
	if !ok {
		t.Fatal("SetPercent must yield a FrameMsg")
	}

	m.view = listView
	if _, cmd := m.Update(frame); cmd != nil {
		t.Error("frames outside progress views must be swallowed")
	}

	m.view = progressView
	if _, cmd := m.Update(frame); cmd == nil {
		t.Error("frames in a progress view must continue the animation")
	}
}

func TestSyncProgressBar_TargetsActiveFlowOnly(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = menuView
	if cmd := m.syncProgressBar(); cmd != nil {
		t.Error("no spring target outside progress views")
	}

	m.view = progressView
	if cmd := m.syncProgressBar(); cmd == nil {
		t.Error("progress view must produce a spring target (and stopwatch start)")
	}
}

func TestRenderProgressLayout_InjectsBarAndElapsed(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView

	out := stripAnsi(m.renderProgressLayout(m.removalLayout()))
	if !strings.Contains(out, "elapsed") {
		t.Errorf("rendered layout missing elapsed readout:\n%s", out)
	}
	if !strings.Contains(out, "%") {
		t.Errorf("rendered layout missing percentage text:\n%s", out)
	}
	if !strings.Contains(out, "░") {
		t.Errorf("rendered layout missing the bar track:\n%s", out)
	}
}
