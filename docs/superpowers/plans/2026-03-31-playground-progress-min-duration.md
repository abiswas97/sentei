# Playground Progress Min-Duration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix `--playground` so reads are instant and all progress views persist for at least 1.5 s before auto-advancing to summary.

**Architecture:** Remove `DelayRunner` (it was throttling reads, not deletion). Add `minProgressDuration` + `progressStartedAt` + `progressToken` to `Model`. A single `holdOrAdvance(targetView)` method either transitions immediately (production / elapsed ≥ min) or fires a `tea.Tick` that delivers `progressHoldExpiredMsg{token}`. Token guards against stale timers. Progress view entry points record `time.Now()` + bump token.

**Tech Stack:** Go 1.21+, Bubble Tea v1.3.10 (`tea.Tick`), table-driven unit tests.

---

## File Map

| File | Change |
|------|--------|
| `internal/git/commands.go` | Delete `DelayRunner` struct + `Run` method; remove `time` import |
| `internal/git/commands_test.go` | Delete `TestDelayRunner_*` tests |
| `main.go` | Remove `playgroundDelay` const, `tuiRunner` var; pass `runner` + `WithMinProgressDuration(1500ms)` |
| `internal/tui/model.go` | Add 4 fields, `ModelOption` type, `WithMinProgressDuration`, `progressHoldExpiredMsg`, pre-dispatch handler in `Update`, `holdOrAdvance` method; variadic opts on `NewMenuModel` |
| `internal/tui/confirm.go` | Set `progressStartedAt`/`progressToken++` on `m.view = progressView` |
| `internal/tui/create_confirm.go` | Same for `createProgressView` |
| `internal/tui/create_branch.go` | Same for `createProgressView` |
| `internal/tui/create_options.go` | Same for `createProgressView` |
| `internal/tui/clone_confirm.go` | Same for `repoProgressView` |
| `internal/tui/clone_input.go` | Same for `repoProgressView` |
| `internal/tui/repo_options.go` | Same for `repoProgressView` |
| `internal/tui/migrate_confirm.go` | Same for `migrateProgressView` |
| `internal/tui/integration_list.go` | Same for `integrationProgressView` |
| `internal/tui/migrate_integrations.go` | Same for `integrationProgressView` |
| `internal/tui/progress.go` | `cleanupCompleteMsg` calls `holdOrAdvance(summaryView)` |
| `internal/tui/create_progress.go` | `createCompleteMsg` calls `holdOrAdvance(createSummaryView)` |
| `internal/tui/repo_progress.go` | `repoDoneMsg` calls `holdOrAdvance(migrateSummaryView or repoSummaryView)` |
| `internal/tui/integration_progress.go` | `integrationFinalizedMsg` calls `holdOrAdvance(m.integ.returnView)` |
| `internal/tui/progress_test.go` | Update `TestUpdateProgress_CleanupComplete_TransitionsToSummary` |

---

## Task 1: Delete `DelayRunner`

**Files:**
- Modify: `internal/git/commands.go`
- Modify: `internal/git/commands_test.go`

- [ ] **Step 1.1: Delete `DelayRunner` from `commands.go`**

Remove lines 51–59 (the `DelayRunner` struct and its `Run` method):

```go
// DELETE these lines entirely:
type DelayRunner struct {
	Inner CommandRunner
	Delay time.Duration
}

func (r *DelayRunner) Run(dir string, args ...string) (string, error) {
	time.Sleep(r.Delay)
	return r.Inner.Run(dir, args...)
}
```

Also remove `"time"` from the import block (it's no longer used).

The file after editing should look like:

```go
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CommandRunner interface {
	Run(dir string, args ...string) (string, error)
}

type GitRunner struct{}

func (r *GitRunner) Run(dir string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", fullArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ShellRunner executes arbitrary shell commands (not git-specific).
type ShellRunner interface {
	RunShell(dir string, command string) (string, error)
}

type DefaultShellRunner struct{}

func (r *DefaultShellRunner) RunShell(dir string, command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", command, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func ValidateRepository(runner CommandRunner, repoPath string) error {
	_, err := runner.Run(repoPath, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}
	return nil
}

func ListWorktrees(runner CommandRunner, repoPath string) ([]Worktree, error) {
	if err := ValidateRepository(runner, repoPath); err != nil {
		return nil, err
	}

	output, err := runner.Run(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	return ParsePorcelain(output)
}
```

- [ ] **Step 1.2: Delete `DelayRunner` tests from `commands_test.go`**

Remove the two test functions (`TestDelayRunner_DelegatesAndPreservesResult` and `TestDelayRunner_PreservesError`, lines 26–64). Also remove the `"time"` import if it's no longer used in that file (check whether other tests use `time`).

- [ ] **Step 1.3: Verify build and tests pass**

```bash
go build ./...
go test ./internal/git/...
```

Expected: no compile errors, all remaining tests PASS.

- [ ] **Step 1.4: Commit**

```bash
git add internal/git/commands.go internal/git/commands_test.go
git commit -m "chore: delete DelayRunner — throttled reads without benefiting progress screens"
```

---

## Task 2: Add Progress Timing Fields and `ModelOption` to `Model`

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 2.1: Write the failing test**

Add to `internal/tui/progress_test.go`:

```go
func TestWithMinProgressDuration_SetsField(t *testing.T) {
	m := NewMenuModel(nil, nil, "/repo", nil, repo.ContextNoRepo,
		WithMinProgressDuration(500*time.Millisecond))

	if m.minProgressDuration != 500*time.Millisecond {
		t.Errorf("minProgressDuration = %v, want 500ms", m.minProgressDuration)
	}
}
```

- [ ] **Step 2.2: Run test to verify it fails**

```bash
go test ./internal/tui/ -run TestWithMinProgressDuration_SetsField -v
```

Expected: FAIL — `WithMinProgressDuration` undefined, and `NewMenuModel` doesn't accept opts.

- [ ] **Step 2.3: Add the fields, type, and option to `model.go`**

Add these imports to `model.go` (if not already present): `"time"`.

Add after the `Model` struct's existing fields (after `integ integrationState`):

```go
// Progress hold state — used to enforce minimum visible duration for progress views.
minProgressDuration time.Duration  // 0 = no hold; set via WithMinProgressDuration
progressStartedAt   time.Time      // set when entering any progress view
progressToken       int            // bumped on each entry; guards stale timers
progressTargetView  viewState      // where to transition when hold expires
```

Add after the `Model` struct definition:

```go
// ModelOption configures a Model at construction time.
type ModelOption func(*Model)

// WithMinProgressDuration sets the minimum time any progress view stays
// visible before the model auto-advances to the summary/result view.
// Default is 0 (advance immediately). Set to ~1.5s in playground mode.
func WithMinProgressDuration(d time.Duration) ModelOption {
	return func(m *Model) {
		m.minProgressDuration = d
	}
}
```

Update `NewMenuModel` signature to accept variadic options and apply them:

```go
func NewMenuModel(runner git.CommandRunner, shell git.ShellRunner, repoPath string, cfg *config.Config, context repo.RepoContext, opts ...ModelOption) Model {
	// ... existing body unchanged ...

	m := Model{ /* existing fields */ }

	for _, opt := range opts {
		opt(&m)
	}

	return m
}
```

- [ ] **Step 2.4: Run test to verify it passes**

```bash
go test ./internal/tui/ -run TestWithMinProgressDuration_SetsField -v
```

Expected: PASS.

- [ ] **Step 2.5: Verify full suite still passes**

```bash
go test ./...
```

Expected: all PASS.

- [ ] **Step 2.6: Commit**

```bash
git add internal/tui/model.go internal/tui/progress_test.go
git commit -m "feat: add ModelOption + WithMinProgressDuration to tui.Model"
```

---

## Task 3: Implement `holdOrAdvance` + `progressHoldExpiredMsg` Handler

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 3.1: Write the failing tests**

Add to `internal/tui/progress_test.go`:

```go
import "time" // add to existing imports

func TestHoldOrAdvance_ZeroDuration_TransitionsImmediately(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	// minProgressDuration defaults to 0

	result, cmd := m.holdOrAdvance(summaryView)
	model := result.(Model)

	if model.view != summaryView {
		t.Errorf("expected summaryView, got %d", model.view)
	}
	if cmd != nil {
		t.Error("expected nil cmd when transitioning immediately")
	}
}

func TestHoldOrAdvance_MinDurationNotElapsed_HoldsAndReturnsCmd(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.minProgressDuration = 10 * time.Second
	m.progressStartedAt = time.Now()
	m.progressToken = 1

	result, cmd := m.holdOrAdvance(summaryView)
	model := result.(Model)

	if model.view == summaryView {
		t.Error("should not transition immediately when min duration not elapsed")
	}
	if cmd == nil {
		t.Error("expected a tea.Tick cmd when holding")
	}
	if model.progressTargetView != summaryView {
		t.Errorf("progressTargetView = %d, want summaryView", model.progressTargetView)
	}
}

func TestHoldOrAdvance_MinDurationElapsed_TransitionsImmediately(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.minProgressDuration = 1 * time.Millisecond
	m.progressStartedAt = time.Now().Add(-100 * time.Millisecond) // started 100ms ago
	m.progressToken = 1

	result, cmd := m.holdOrAdvance(summaryView)
	model := result.(Model)

	if model.view != summaryView {
		t.Errorf("expected summaryView when min elapsed, got %d", model.view)
	}
	if cmd != nil {
		t.Error("expected nil cmd when min duration already elapsed")
	}
}

func TestProgressHoldExpiredMsg_CorrectToken_Transitions(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView
	m.progressToken = 3
	m.progressTargetView = summaryView

	result, cmd := m.Update(progressHoldExpiredMsg{token: 3})
	model := result.(Model)

	if model.view != summaryView {
		t.Errorf("expected summaryView, got %d", model.view)
	}
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
}

func TestProgressHoldExpiredMsg_StaleToken_Ignored(t *testing.T) {
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView
	m.progressToken = 5
	m.progressTargetView = summaryView

	result, _ := m.Update(progressHoldExpiredMsg{token: 3}) // stale token
	model := result.(Model)

	if model.view == summaryView {
		t.Error("stale token should not transition view")
	}
}
```

- [ ] **Step 3.2: Run tests to verify they fail**

```bash
go test ./internal/tui/ -run "TestHoldOrAdvance|TestProgressHoldExpired" -v
```

Expected: FAIL — `holdOrAdvance` and `progressHoldExpiredMsg` undefined.

- [ ] **Step 3.3: Add `progressHoldExpiredMsg` and `holdOrAdvance` to `model.go`**

Add the message type near the other message type definitions (e.g., after imports):

```go
// progressHoldExpiredMsg fires when the minimum progress view duration has elapsed.
// The token must match model.progressToken to guard against stale messages.
type progressHoldExpiredMsg struct{ token int }
```

Add the `holdOrAdvance` method (add after `NewMenuModel`):

```go
// holdOrAdvance either transitions to targetView immediately (if minProgressDuration
// is zero or has already elapsed) or schedules a tea.Tick for the remaining time.
// The caller must store all result state into the model before calling.
func (m Model) holdOrAdvance(targetView viewState) (Model, tea.Cmd) {
	m.progressTargetView = targetView
	if m.minProgressDuration == 0 || time.Since(m.progressStartedAt) >= m.minProgressDuration {
		m.view = targetView
		return m, nil
	}
	remaining := m.minProgressDuration - time.Since(m.progressStartedAt)
	token := m.progressToken
	return m, tea.Tick(remaining, func(time.Time) tea.Msg {
		return progressHoldExpiredMsg{token: token}
	})
}
```

Add the pre-dispatch handler at the TOP of `Update`, before `switch m.view`:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if holdMsg, ok := msg.(progressHoldExpiredMsg); ok {
		if holdMsg.token == m.progressToken {
			m.view = m.progressTargetView
		}
		return m, nil
	}

	switch m.view {
	// ... existing cases unchanged ...
	}
	return m, nil
}
```

Add `"time"` to imports in `model.go` if not already present.

- [ ] **Step 3.4: Run tests to verify they pass**

```bash
go test ./internal/tui/ -run "TestHoldOrAdvance|TestProgressHoldExpired" -v
```

Expected: all 5 PASS.

- [ ] **Step 3.5: Verify full suite**

```bash
go test ./...
```

Expected: all PASS.

- [ ] **Step 3.6: Commit**

```bash
git add internal/tui/model.go internal/tui/progress_test.go
git commit -m "feat: add holdOrAdvance helper and progressHoldExpiredMsg for min-duration hold"
```

---

## Task 4: Wire Progress Timer Start at All View Entry Points

**Files:**
- Modify: `internal/tui/confirm.go`
- Modify: `internal/tui/create_confirm.go`
- Modify: `internal/tui/create_branch.go`
- Modify: `internal/tui/create_options.go`
- Modify: `internal/tui/clone_confirm.go`
- Modify: `internal/tui/clone_input.go`
- Modify: `internal/tui/repo_options.go`
- Modify: `internal/tui/migrate_confirm.go`
- Modify: `internal/tui/integration_list.go`
- Modify: `internal/tui/migrate_integrations.go`

The pattern at each entry point: immediately before (or after) setting `m.view = <progressView>`, add:

```go
m.progressStartedAt = time.Now()
m.progressToken++
```

Find each entry point:

**`confirm.go`** — `m.view = progressView` (two locations: key press and `teardownCompleteMsg`). Add the two lines before each `m.view = progressView` assignment.

**`create_confirm.go`** — `m.view = createProgressView`. Add before the assignment.

**`create_branch.go`** — `m.view = createProgressView`. Add before the assignment.

**`create_options.go`** — `m.view = createProgressView`. Add before the assignment.

**`clone_confirm.go`** — `m.view = repoProgressView`. Add before the assignment.

**`clone_input.go`** — `m.view = repoProgressView`. Add before the assignment.

**`repo_options.go`** — `m.view = repoProgressView`. Add before the assignment.

**`migrate_confirm.go`** — `m.view = migrateProgressView`. Add before the assignment.

**`integration_list.go`** — `m.view = integrationProgressView`. Add before the assignment.

**`migrate_integrations.go`** — `m.view = integrationProgressView`. Add before the assignment.

All files that use `time.Now()` need `"time"` in their imports. Add it to each file's import block if not present.

- [ ] **Step 4.1: Add timer start to all 10 files (example shown for `confirm.go`)**

In `confirm.go`, find the `m.view = progressView` assignment (inside `case key.Matches(msg, keys.Yes):`):

```go
// Before:
m.view = progressView

// After:
m.progressStartedAt = time.Now()
m.progressToken++
m.view = progressView
```

Apply the same pattern to the second `m.view = progressView` in the `teardownCompleteMsg` case.

Repeat for all other files using their respective progress view constants.

- [ ] **Step 4.2: Verify build and tests pass**

```bash
go build ./...
go test ./...
```

Expected: all PASS. No test exercises this directly yet — coverage comes in Task 5.

- [ ] **Step 4.3: Commit**

```bash
git add internal/tui/confirm.go internal/tui/create_confirm.go internal/tui/create_branch.go \
        internal/tui/create_options.go internal/tui/clone_confirm.go internal/tui/clone_input.go \
        internal/tui/repo_options.go internal/tui/migrate_confirm.go \
        internal/tui/integration_list.go internal/tui/migrate_integrations.go
git commit -m "feat: record progress view start time at all entry points"
```

---

## Task 5: Replace Direct View Transitions with `holdOrAdvance`

**Files:**
- Modify: `internal/tui/progress.go`
- Modify: `internal/tui/create_progress.go`
- Modify: `internal/tui/repo_progress.go`
- Modify: `internal/tui/integration_progress.go`
- Test: `internal/tui/progress_test.go`

- [ ] **Step 5.1: Update `TestUpdateProgress_CleanupComplete_TransitionsToSummary`**

The existing test uses `NewModel` (which has `minProgressDuration = 0`), so with `holdOrAdvance` the model still transitions immediately. The test should pass unchanged. Verify by running it first:

```bash
go test ./internal/tui/ -run TestUpdateProgress_CleanupComplete_TransitionsToSummary -v
```

Expected: PASS (test exercises `minProgressDuration = 0` path — immediate transition).

- [ ] **Step 5.2: Update `cleanupCompleteMsg` handler in `progress.go`**

Find the `cleanupCompleteMsg` case (currently sets `m.view = summaryView` directly):

```go
// Before:
case cleanupCompleteMsg:
    m.remove.cleanupResult = &msg.Result
    m.stateStale = true
    m.view = summaryView

// After:
case cleanupCompleteMsg:
    m.remove.cleanupResult = &msg.Result
    m.stateStale = true
    return m.holdOrAdvance(summaryView)
```

- [ ] **Step 5.3: Run deletion progress tests**

```bash
go test ./internal/tui/ -run "TestUpdateProgress" -v
```

Expected: all PASS.

- [ ] **Step 5.4: Update `createCompleteMsg` handler in `create_progress.go`**

```go
// Before:
case createCompleteMsg:
    m.create.result = &msg.Result
    m.stateStale = true
    m.view = createSummaryView
    return m, nil

// After:
case createCompleteMsg:
    m.create.result = &msg.Result
    m.stateStale = true
    return m.holdOrAdvance(createSummaryView)
```

- [ ] **Step 5.5: Update `repoDoneMsg` handler in `repo_progress.go`**

```go
// Before:
case repoDoneMsg:
    m.repo.result = msg.result
    m.stateStale = true
    if m.repo.opType == "migrate" {
        m.view = migrateSummaryView
    } else {
        m.view = repoSummaryView
    }
    return m, nil

// After:
case repoDoneMsg:
    m.repo.result = msg.result
    m.stateStale = true
    targetView := repoSummaryView
    if m.repo.opType == "migrate" {
        targetView = migrateSummaryView
    }
    return m.holdOrAdvance(targetView)
```

- [ ] **Step 5.6: Update `integrationFinalizedMsg` handler in `integration_progress.go`**

```go
// Before:
case integrationFinalizedMsg:
    if m.integ.returnView != migrateNextView {
        m.integ.current = msg.current
        for _, integ := range m.integ.integrations {
            m.integ.staged[integ.Name] = m.integ.current[integ.Name]
        }
    }
    m.stateStale = true
    m.view = m.integ.returnView
    return m, nil

// After:
case integrationFinalizedMsg:
    if m.integ.returnView != migrateNextView {
        m.integ.current = msg.current
        for _, integ := range m.integ.integrations {
            m.integ.staged[integ.Name] = m.integ.current[integ.Name]
        }
    }
    m.stateStale = true
    return m.holdOrAdvance(m.integ.returnView)
```

- [ ] **Step 5.7: Verify full test suite**

```bash
go test ./...
```

Expected: all PASS.

- [ ] **Step 5.8: Commit**

```bash
git add internal/tui/progress.go internal/tui/create_progress.go \
        internal/tui/repo_progress.go internal/tui/integration_progress.go
git commit -m "feat: use holdOrAdvance in all progress completion handlers"
```

---

## Task 6: Activate Playground Min-Duration and Clean Up `main.go`

**Files:**
- Modify: `main.go`

- [ ] **Step 6.1: Remove `DelayRunner` usage and update `main.go`**

Remove the `playgroundDelay` constant (line 31):
```go
// DELETE:
const playgroundDelay = 800 * time.Millisecond
```

Remove the `tuiRunner` block (lines 343–346):
```go
// DELETE:
var tuiRunner git.CommandRunner = runner
if *playgroundFlag {
    tuiRunner = &git.DelayRunner{Inner: runner, Delay: playgroundDelay}
}
```

Update the `NewMenuModel` call that used `tuiRunner` to use `runner` with the playground option:
```go
// Before:
model := tui.NewMenuModel(tuiRunner, shell, repoPath, cfg, context)

// After:
var menuOpts []tui.ModelOption
if *playgroundFlag {
    menuOpts = append(menuOpts, tui.WithMinProgressDuration(1500*time.Millisecond))
}
model := tui.NewMenuModel(runner, shell, repoPath, cfg, context, menuOpts...)
```

Remove the `"time"` import from `main.go` if it's no longer referenced (check whether `time.Duration` or other time usage remains).

- [ ] **Step 6.2: Build and verify**

```bash
go build ./...
go vet ./...
go test ./...
```

Expected: all PASS, no warnings.

- [ ] **Step 6.3: Commit**

```bash
git add main.go
git commit -m "feat: activate playground min-progress-duration; remove DelayRunner from main"
```

---

## Task 7: Verify End-to-End With Playground

**Files:**
- Test: `internal/tui/cleanup_e2e_test.go` (reference for patterns)

- [ ] **Step 7.1: Write a unit test for the full hold cycle**

Add to `internal/tui/progress_test.go`:

```go
func TestHoldCycle_CompletionHoldsUntilExpiry(t *testing.T) {
	// Simulate a model that just entered progressView 0ms ago with a 50ms hold.
	m := NewModel([]git.Worktree{}, nil, "/repo")
	m.view = progressView
	m.minProgressDuration = 50 * time.Millisecond
	m.progressStartedAt = time.Now()
	m.progressToken = 1

	// Operation completes immediately.
	result, cmd := m.updateProgress(cleanupCompleteMsg{})
	model := result.(Model)

	// Should NOT have transitioned yet — min duration not elapsed.
	if model.view == summaryView {
		t.Fatal("should not transition to summaryView before min duration elapses")
	}
	if cmd == nil {
		t.Fatal("expected a hold cmd (tea.Tick)")
	}

	// Execute the cmd — this blocks for up to 50ms then returns progressHoldExpiredMsg.
	msg := cmd()
	holdMsg, ok := msg.(progressHoldExpiredMsg)
	if !ok {
		t.Fatalf("expected progressHoldExpiredMsg, got %T", msg)
	}
	if holdMsg.token != 1 {
		t.Errorf("token = %d, want 1", holdMsg.token)
	}

	// Now deliver the expired message.
	result2, _ := model.Update(holdMsg)
	model2 := result2.(Model)

	if model2.view != summaryView {
		t.Errorf("expected summaryView after hold expired, got %d", model2.view)
	}
}
```

- [ ] **Step 7.2: Run the test**

```bash
go test ./internal/tui/ -run TestHoldCycle_CompletionHoldsUntilExpiry -v -timeout 10s
```

Expected: PASS (takes ~50ms for the tea.Tick to fire).

- [ ] **Step 7.3: Run the full test suite one final time**

```bash
go fmt ./...
go vet ./...
go test ./...
```

Expected: all PASS, zero warnings.

- [ ] **Step 7.4: Commit**

```bash
git add internal/tui/progress_test.go
git commit -m "test: add hold-cycle integration test for playground min-duration"
```

---

## Self-Review

### Spec Coverage

| Spec requirement | Task |
|-----------------|------|
| Remove `DelayRunner` (reads instant) | Task 1 + Task 6 |
| `minProgressDuration` field on Model | Task 2 |
| `WithMinProgressDuration` option | Task 2 |
| `holdOrAdvance` helper | Task 3 |
| `progressHoldExpiredMsg` + stale token guard | Task 3 |
| `progressStartedAt` / `progressToken` set on view entry | Task 4 |
| All 4 completion handlers use `holdOrAdvance` | Task 5 |
| Playground sets 1.5s min duration | Task 6 |
| Tests: zero min-duration → immediate (existing tests unchanged) | Tasks 3, 5 |
| Tests: min-duration > 0 → hold → expire → transition | Task 7 |

No gaps.

### Type / Name Consistency

- `progressHoldExpiredMsg` used consistently in Task 3, 7.
- `holdOrAdvance(targetView viewState) (Model, tea.Cmd)` defined in Task 3, called in Task 5.
- `WithMinProgressDuration` defined in Task 2, used in Task 6.
- `progressToken`, `progressStartedAt`, `progressTargetView`, `minProgressDuration` defined in Task 2, used in Tasks 3, 4.

### No Placeholders

Checked — all code blocks are complete.
