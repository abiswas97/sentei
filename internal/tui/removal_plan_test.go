package tui

import (
	"testing"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/progress"
)

// Spec "Teardown counts are real": the running Teardown phase displays the
// planned total from its first frame, not a placeholder.
func TestBuildRemovalPhases_TeardownDeclaresPlannedTotal(t *testing.T) {
	m := NewModel([]git.Worktree{{Path: "/wt/a"}, {Path: "/wt/b"}}, nil, "/repo")
	m.remove.run = newRemovalRun(m.remove.worktrees)
	m.remove.run.teardownRunning = true
	m.remove.run.teardownPlanned = []string{
		"Teardown code-review-graph", "Teardown cocoindex-code",
		"Teardown code-review-graph", "Teardown cocoindex-code",
	}

	phases := m.buildRemovalPhases()
	td := phases[0]
	if td.Name != "Teardown" || td.Total != 4 {
		t.Fatalf("running teardown = %s %d, want Teardown with total 4 from the plan", td.Name, td.Total)
	}
	if td.Done != 0 || td.Settled() {
		t.Error("running teardown must not report completion before results land")
	}
}

// Spec "Parallel removal moves the bar at start": started-but-unfinished
// removals reach their first checkpoint while headers still count 0 done.
func TestBuildRemovalPhases_StartCheckpointsMoveTheBar(t *testing.T) {
	wts := []git.Worktree{{Path: "/wt/a"}, {Path: "/wt/b"}, {Path: "/wt/c"}}
	m := NewModel(wts, nil, "/repo")
	m.remove.run = newRemovalRun(wts)
	for _, wt := range wts {
		m.remove.run.statuses[wt.Path] = statusRemoving
	}

	phases := m.buildRemovalPhases()
	var removing progress.PhaseState
	for _, p := range phases {
		if p.Name == "Removing worktrees" {
			removing = p
		}
	}
	if removing.Done != 0 {
		t.Fatalf("headers must count 0 done while all removals run, got %d", removing.Done)
	}
	reached, declared := progress.CheckpointProgress([]progress.PhaseState{removing})
	if reached != 3 || declared != 6 {
		t.Errorf("checkpoints = %d/%d, want 3/6 (start credit for three parallel removals)", reached, declared)
	}
}
