package progress

import (
	"strings"
	"testing"
)

func TestNew_TotalSetUpfront(t *testing.T) {
	tr := New(10)
	if tr.Total() != 10 {
		t.Errorf("expected total=10, got %d", tr.Total())
	}
	if tr.Done() != 0 {
		t.Errorf("expected done=0, got %d", tr.Done())
	}
}

func TestNewFromGroups_CalculatesTotal(t *testing.T) {
	tr := NewFromGroups([]Group{
		{Name: "setup", Steps: []Step{{Name: "s1"}, {Name: "s2"}}},
		{Name: "deps", Steps: []Step{{Name: "d1"}}},
	})
	if tr.Total() != 3 {
		t.Errorf("expected total=3, got %d", tr.Total())
	}
}

func TestUpdate_RecordsStepStatus(t *testing.T) {
	tr := New(3)
	tr.Update("worktree-a", "install", Running, "")
	tr.Update("worktree-a", "install", Done, "")

	groups := tr.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(groups[0].Steps))
	}
	if groups[0].Steps[0].Status != Done {
		t.Errorf("expected Done, got %d", groups[0].Steps[0].Status)
	}
}

func TestUpdate_DeduplicatesSteps(t *testing.T) {
	tr := New(3)
	// Same step emits Running then Done — should be 1 step, not 2.
	tr.Update("wt", "setup", Running, "")
	tr.Update("wt", "setup", Done, "")

	groups := tr.Groups()
	if len(groups[0].Steps) != 1 {
		t.Errorf("expected 1 step (deduplicated), got %d", len(groups[0].Steps))
	}
}

func TestDone_CountsDoneAndFailed(t *testing.T) {
	tr := New(4)
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Failed, "oops")
	tr.Update("g", "s3", Running, "")

	if tr.Done() != 2 {
		t.Errorf("expected done=2 (1 Done + 1 Failed), got %d", tr.Done())
	}
}

func TestTotal_UpfrontTotalIsMinimum(t *testing.T) {
	tr := New(5)
	// Only 2 steps discovered so far — total should still be 5.
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Running, "")

	if tr.Total() != 5 {
		t.Errorf("expected total=5 (upfront), got %d", tr.Total())
	}
}

func TestTotal_GrowsBeyondUpfront(t *testing.T) {
	tr := New(2)
	// More steps discovered than declared upfront — total grows.
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Done, "")
	tr.Update("g", "s3", Done, "")

	if tr.Total() != 3 {
		t.Errorf("expected total=3 (actual > upfront), got %d", tr.Total())
	}
}

func TestIsComplete(t *testing.T) {
	tr := New(2)
	if tr.IsComplete() {
		t.Error("should not be complete with 0 steps")
	}

	tr.Update("g", "s1", Done, "")
	if tr.IsComplete() {
		t.Error("should not be complete with 1/2 done")
	}

	tr.Update("g", "s2", Done, "")
	if !tr.IsComplete() {
		t.Error("should be complete with 2/2 done")
	}
}

func TestIsComplete_FailedCountsAsComplete(t *testing.T) {
	tr := New(2)
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Failed, "err")
	if !tr.IsComplete() {
		t.Error("should be complete with 1 Done + 1 Failed = 2/2")
	}
}

func TestFailedCount(t *testing.T) {
	tr := New(3)
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Failed, "err1")
	tr.Update("g", "s3", Failed, "err2")
	if tr.FailedCount() != 2 {
		t.Errorf("expected 2 failed, got %d", tr.FailedCount())
	}
}

func TestGroups_PreservesInsertionOrder(t *testing.T) {
	tr := New(3)
	tr.Update("charlie", "s1", Done, "")
	tr.Update("alpha", "s1", Done, "")
	tr.Update("bravo", "s1", Done, "")

	groups := tr.Groups()
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Name != "charlie" || groups[1].Name != "alpha" || groups[2].Name != "bravo" {
		t.Errorf("expected insertion order charlie/alpha/bravo, got %s/%s/%s",
			groups[0].Name, groups[1].Name, groups[2].Name)
	}
}

func TestGroups_StepsPreserveOrder(t *testing.T) {
	tr := New(3)
	tr.Update("g", "install", Running, "")
	tr.Update("g", "setup", Running, "")
	tr.Update("g", "verify", Running, "")

	steps := tr.Groups()[0].Steps
	if steps[0].Name != "install" || steps[1].Name != "setup" || steps[2].Name != "verify" {
		t.Error("steps should preserve insertion order")
	}
}

func TestUpdate_ErrorMessage(t *testing.T) {
	tr := New(1)
	tr.Update("g", "s1", Failed, "connection refused")
	if tr.Groups()[0].Steps[0].Error != "connection refused" {
		t.Error("expected error message to be stored")
	}
}

func TestUpdate_ErrorClearedOnRetry(t *testing.T) {
	tr := New(1)
	tr.Update("g", "s1", Failed, "timeout")
	tr.Update("g", "s1", Done, "")
	if tr.Groups()[0].Steps[0].Error != "" {
		t.Error("expected error to be cleared after success")
	}
}

func TestBar_EmptyTracker(t *testing.T) {
	tr := New(0)
	bar := tr.Bar(10)
	if len(bar) == 0 {
		t.Error("expected a bar even for empty tracker")
	}
}

func TestBar_HalfFilled(t *testing.T) {
	tr := New(4)
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Done, "")

	bar := tr.Bar(10)
	filled := strings.Count(bar, "\u2588")
	empty := strings.Count(bar, "\u2591")
	if filled != 5 || empty != 5 {
		t.Errorf("expected 5 filled + 5 empty for 2/4, got %d filled + %d empty", filled, empty)
	}
}

func TestPercent_Zero(t *testing.T) {
	tr := New(0)
	if tr.Percent() != 0 {
		t.Errorf("expected 0%%, got %d%%", tr.Percent())
	}
}

func TestPercent_Half(t *testing.T) {
	tr := New(4)
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Done, "")
	if tr.Percent() != 50 {
		t.Errorf("expected 50%%, got %d%%", tr.Percent())
	}
}

func TestPercent_Full(t *testing.T) {
	tr := New(3)
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Done, "")
	tr.Update("g", "s3", Done, "")
	if tr.Percent() != 100 {
		t.Errorf("expected 100%%, got %d%%", tr.Percent())
	}
}

func TestBar_FullyComplete(t *testing.T) {
	tr := New(2)
	tr.Update("g", "s1", Done, "")
	tr.Update("g", "s2", Done, "")

	bar := tr.Bar(10)
	filled := strings.Count(bar, "\u2588")
	if filled != 10 {
		t.Errorf("expected 10 filled for 2/2, got %d", filled)
	}
}

func TestNewFromGroups_PreDeclaredStepsUpdateCorrectly(t *testing.T) {
	tr := NewFromGroups([]Group{
		{Name: "wt-main", Steps: []Step{
			{Name: "install"},
			{Name: "setup"},
		}},
		{Name: "wt-feature", Steps: []Step{
			{Name: "install"},
			{Name: "setup"},
		}},
	})

	if tr.Total() != 4 {
		t.Fatalf("expected total=4, got %d", tr.Total())
	}
	if tr.Done() != 0 {
		t.Fatalf("expected done=0 initially, got %d", tr.Done())
	}

	tr.Update("wt-main", "install", Done, "")
	tr.Update("wt-main", "setup", Done, "")
	tr.Update("wt-feature", "install", Running, "")

	if tr.Done() != 2 {
		t.Errorf("expected done=2, got %d", tr.Done())
	}
	if tr.Total() != 4 {
		t.Errorf("expected total=4 (stable), got %d", tr.Total())
	}
}

func TestMultipleGroups_ProgressAccumulatesCorrectly(t *testing.T) {
	tr := New(9) // 3 worktrees × 3 steps each
	for _, wt := range []string{"main", "feature", "hotfix"} {
		for _, step := range []string{"install", "setup", "verify"} {
			tr.Update(wt, step, Running, "")
			tr.Update(wt, step, Done, "")
		}
	}

	if tr.Done() != 9 {
		t.Errorf("expected done=9, got %d", tr.Done())
	}
	if tr.Total() != 9 {
		t.Errorf("expected total=9, got %d", tr.Total())
	}
	if !tr.IsComplete() {
		t.Error("should be complete")
	}
}
