package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/abiswas97/sentei/internal/cleanup"
	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/integration"
	"github.com/abiswas97/sentei/internal/progress"
	"github.com/abiswas97/sentei/internal/repo"
)

func TestViewProgress_NoPurpleBadge_HasChromeAndBar(t *testing.T) {
	m := NewModel([]git.Worktree{{Path: "/work/a", Branch: "refs/heads/a"}}, nil, "/repo")
	m.width, m.height = 80, 30
	m.remove.run = newRemovalRun([]git.Worktree{{Path: "/work/a", Branch: "refs/heads/a"}})
	m.view = progressView

	view := stripANSI(m.viewProgress())
	if !strings.Contains(view, "sentei ─ Removing worktrees") {
		t.Errorf("expected standard title, view:\n%s", view)
	}
	if !strings.Contains(view, "░") {
		t.Errorf("expected overall progress bar, view:\n%s", view)
	}
	if !strings.Contains(view, "q quit") {
		t.Errorf("expected key hints, view:\n%s", view)
	}
}

func TestBuildRemovalPhases_PruneStaging(t *testing.T) {
	m := NewModel(nil, nil, "/repo")
	m.remove.run = newRemovalRun([]git.Worktree{{Path: "/work/a", Branch: "refs/heads/a"}})

	phases := m.buildRemovalPhases()
	last := phases[len(phases)-1]
	if last.name != "Prune & cleanup" || last.total != 0 {
		t.Errorf("expected pending prune phase before removal completes, got %+v", last)
	}

	// Removal finished: prune phase becomes active work.
	m.remove.run.statuses["/work/a"] = statusRemoved
	phases = m.buildRemovalPhases()
	last = phases[len(phases)-1]
	if last.total != 2 || last.done != 0 {
		t.Errorf("expected active 0/2 prune phase after removal completes, got %+v", last)
	}

	pruneErr := error(nil)
	m.remove.run.pruneErr = &pruneErr
	m.remove.run.cleanupResult = &cleanup.Result{}
	phases = m.buildRemovalPhases()
	last = phases[len(phases)-1]
	if last.done != 2 || last.failed != 0 {
		t.Errorf("expected completed prune phase, got %+v", last)
	}
}

func TestBuildIntegrationPhases_PrePopulatesTargets(t *testing.T) {
	m := NewMenuModel(nil, nil, "/repo", nil, repo.ContextBareRepo)
	m.integ.targetWorktrees = []string{"/repo/feature-a", "/repo/feature-b"}

	phases := m.buildIntegrationPhases()
	if len(phases) != 2 {
		t.Fatalf("expected 2 pre-populated phases, got %d", len(phases))
	}
	for _, p := range phases {
		if p.total != 0 {
			t.Errorf("pre-populated target %q must be pending (total 0), got %d", p.name, p.total)
		}
	}

	view := stripANSI(m.viewIntegrationProgress())
	for _, want := range []string{"feature-a", "feature-b", "pending"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected %q visible before events arrive, view:\n%s", want, view)
		}
	}
}

func TestBuildIntegrationPhases_ErrorBakedIntoLabel(t *testing.T) {
	m := NewMenuModel(nil, nil, "/repo", nil, repo.ContextBareRepo)
	m.integ.events = []integration.ManagerEvent{
		{Worktree: "/repo/a", Step: "Install pipx", Status: integration.StatusFailed, Error: errors.New("exit 1")},
	}

	phases := m.buildIntegrationPhases()
	if len(phases) != 1 || len(phases[0].steps) != 1 {
		t.Fatalf("unexpected phases: %+v", phases)
	}
	if phases[0].steps[0].status != progress.StepFailed {
		t.Errorf("expected failed step, got %v", phases[0].steps[0].status)
	}
	if !strings.Contains(phases[0].steps[0].name, "exit 1") {
		t.Errorf("expected error in step label, got %q", phases[0].steps[0].name)
	}
}

func TestViewSummary_VerdictMarkerAndNoEmptyCleanupHeader(t *testing.T) {
	m := NewModel(nil, nil, "/repo")
	m.width = 80
	m.remove.run = newRemovalRun(nil)
	m.remove.run.result.SuccessCount = 2
	m.remove.run.cleanupResult = &cleanup.Result{NonWtBranchesRemaining: 1}
	m.view = summaryView

	view := stripANSI(m.viewSummary())
	if !strings.Contains(view, "✦ 2 worktrees removed successfully") {
		t.Errorf("expected ✦ verdict marker on the headline, view:\n%s", view)
	}
	if strings.Contains(view, "Cleanup:") {
		t.Errorf("Cleanup header must be omitted when nothing was cleaned, view:\n%s", view)
	}
	if !strings.Contains(view, "Tip: 1 local branch ") {
		t.Errorf("tip should render independently of the Cleanup header, view:\n%s", view)
	}
	if strings.Contains(view, " v ") {
		t.Errorf("legacy \"v\" marker must be gone, view:\n%s", view)
	}
}

func TestViewSummary_CleanupLinesPluralizeSingular(t *testing.T) {
	m := NewModel(nil, nil, "/repo")
	m.width = 80
	m.remove.run = newRemovalRun(nil)
	m.remove.run.result.SuccessCount = 1
	m.remove.run.cleanupResult = &cleanup.Result{
		StaleRefsRemoved:    1,
		GoneBranchesDeleted: 1,
		ConfigDedupResult:   cleanup.ConfigResult{Removed: 1},
		ConfigOrphanResult:  cleanup.ConfigResult{Removed: 1},
	}
	m.view = summaryView

	view := stripANSI(m.viewSummary())

	for _, want := range []string{
		"1 worktree removed successfully",
		"Pruned 1 remote ref",
		"Removed 1 config duplicate",
		"Deleted 1 branch with gone upstream",
		"Removed 1 orphaned config section",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("expected singular %q, view:\n%s", want, view)
		}
	}
	// Singular nouns are prefixes of their plurals, so guard against the naive
	// always-plural regression explicitly.
	for _, plural := range []string{"worktrees removed", "remote refs", "config duplicates", "branches with gone", "config sections"} {
		if strings.Contains(view, plural) {
			t.Errorf("count of 1 must not read as plural %q, view:\n%s", plural, view)
		}
	}
}

func TestViewCleanupResult_RunningTitleWhileNilResult(t *testing.T) {
	m := NewMenuModel(nil, nil, "/repo", nil, repo.ContextBareRepo)
	m.width = 80
	m.view = cleanupResultView

	view := stripANSI(m.viewCleanupResult())
	if !strings.Contains(view, "sentei ─ Running cleanup") {
		t.Errorf("expected Running cleanup title while result is nil, view:\n%s", view)
	}
	if strings.Contains(view, "Cleanup complete") {
		t.Errorf("must not claim completion while running, view:\n%s", view)
	}
}

func TestViewConfirm_NoBorderAndDetachedHeadLabel(t *testing.T) {
	wts := []git.Worktree{{Path: "/work/detached-head", Branch: ""}}
	m := NewModel(wts, nil, "/repo")
	m.width = 80
	m.remove.selected = map[string]bool{"/work/detached-head": true}
	m.view = confirmView

	view := stripANSI(m.viewConfirm())
	if strings.Contains(view, "╭") {
		t.Errorf("confirmation must not render a border box, view:\n%s", view)
	}
	if !strings.Contains(view, "detached-head") {
		t.Errorf("detached HEAD worktree must show its directory name, view:\n%s", view)
	}
	if !strings.Contains(view, "y delete · n go back") {
		t.Errorf("expected standard hints, view:\n%s", view)
	}
}

func TestConfirmationViewModel_NoBorderStandardHints(t *testing.T) {
	vm := ConfirmationViewModel{
		Title:      "Confirm cleanup",
		Items:      []ConfirmationItem{{Label: "Mode:", Value: "safe"}},
		CLICommand: "sentei cleanup --mode safe",
	}
	view := stripANSI(vm.View())
	if strings.Contains(view, "╭") {
		t.Errorf("confirmation view must not render a border, view:\n%s", view)
	}
	if !strings.Contains(view, "enter confirm · esc back · q quit") {
		t.Errorf("expected `·`-separated hints, view:\n%s", view)
	}
	if strings.Contains(view, "•") {
		t.Errorf("bullet separators must be gone, view:\n%s", view)
	}
}

func TestViewList_HasRepoSubtitleAndRules(t *testing.T) {
	m := NewModel([]git.Worktree{{Path: "/work/a", Branch: "refs/heads/a"}}, nil, "/repo/myrepo.git")
	m.width = 80
	m.view = listView

	view := stripANSI(m.viewList())
	if !strings.Contains(view, "myrepo.git (bare)") {
		t.Errorf("expected repo subtitle framing, view:\n%s", view)
	}
	if !strings.Contains(view, "┄") {
		t.Errorf("expected rule framing, view:\n%s", view)
	}
}
