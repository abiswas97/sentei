# Progress Arc Correctness and Live Region Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every Sentei progress flow monotonic and terminally honest, render it in a pinned responsive live region, and provide deterministic VHS GIFs of the corrected UX.

**Architecture:** `internal/progress` gains stable step identity and a plan-owned, concurrency-safe execution contract. Producers prepare exact plans before running and emit through that contract; every TUI progress view folds the resulting event stream through `progress.Snapshot`. `ProgressLayout` projects snapshots into bounded history, focus, and queue regions while the existing Charm primitives supply the bar, help, portal, and purposeful motion.

**Tech Stack:** Go 1.25, Charm Bubble Tea/Bubbles/Lip Gloss v2, `teatest`, shell hook tests, Charmbracelet VHS 0.11, ffmpeg/ffprobe.

---

## File Map

| File | Responsibility |
|------|----------------|
| `internal/progress/plan.go` | Stable planned IDs/labels and plan validation |
| `internal/progress/execution.go` | Concurrency-safe declaration, transitions, checkpoints, and terminalization |
| `internal/progress/snapshot.go` | Single fold retaining IDs, labels, messages, and errors |
| `internal/progress/validate.go` | Stream invariant validation |
| `internal/integration/apply.go` | Read-only preparation plus frozen integration execution |
| `internal/creator/plan.go` | Exact creator plan and stable IDs |
| `internal/creator/*.go` | Execute creator work only through the prepared plan |
| `internal/repo/plan.go` | Static create/clone/migrate plans |
| `internal/repo/{create,clone,migrate}.go` | Execute repository plans and skip blocked suffixes |
| `internal/worktree/deleter.go` | Removal checkpoint events through shared execution |
| `internal/tui/removal_progress.go` | Removal event history and Snapshot projection |
| `internal/tui/progress_viewport.go` | Pure bounded history/focus/queue projection |
| `internal/tui/progress_layout.go` | Pinned responsive rendering using the projection |
| `internal/tui/progress_details.go` | Full progress trace/error portal content |
| `internal/tui/motion_preference.go` | Static-motion environment policy |
| `scripts/pre-commit` | Effective author/committer identity guard |
| `scripts/pre-commit_test.sh` | Hook identity regression tests |
| `.demos/progress-arc/*.tape` | Deterministic VHS recordings |

## Task 1: Make Plans Own Progress Truth

**Files:**
- Modify: `internal/progress/plan.go`
- Create: `internal/progress/execution.go`
- Modify: `internal/progress/progress.go`
- Modify: `internal/progress/snapshot.go`
- Modify: `internal/progress/validate.go`
- Test: `internal/progress/execution_test.go`
- Test: `internal/progress/snapshot_test.go`
- Test: `internal/progress/plan_test.go`

- [ ] **Step 1.1: Write failing stable-identity and execution-contract tests**

Add tests that express the final API and invariants:

```go
func TestExecution_FinishSettlesDistinctStepsWithEqualLabels(t *testing.T) {
	plan := Plan{Phases: []PlannedPhase{{ID: "integrations", Label: "Integrations", Steps: []PlannedStep{
		{ID: "ccc.copy-index", Label: "Copy index from main"},
		{ID: "crg.copy-index", Label: "Copy index from main"},
	}}}}
	var events []Event
	x, err := Start(plan, func(ev Event) { events = append(events, ev) })
	if err != nil { t.Fatal(err) }
	if _, err := x.Done("integrations", "ccc.copy-index", ""); err != nil { t.Fatal(err) }
	if err := x.Finish("blocked by earlier failure"); err != nil { t.Fatal(err) }

	states := Snapshot(events)
	if len(states) != 1 || len(states[0].Steps) != 2 { t.Fatalf("states = %#v", states) }
	if !states[0].Settled() { t.Fatalf("phase did not settle: %#v", states[0]) }
	if states[0].Steps[1].Status != StepSkipped { t.Fatalf("second step = %#v", states[0].Steps[1]) }
}

func TestExecution_RejectsUndeclaredAndTerminalMutation(t *testing.T) {
	x, err := Start(Plan{Phases: []PlannedPhase{{ID: "p", Label: "Phase", Steps: []PlannedStep{{ID: "s", Label: "Step"}}}}}, func(Event) {})
	if err != nil { t.Fatal(err) }
	if _, err := x.Done("p", "missing", ""); err == nil { t.Fatal("undeclared step accepted") }
	if _, err := x.Done("p", "s", ""); err != nil { t.Fatal(err) }
	if _, err := x.Fail("p", "s", errors.New("late")); err == nil { t.Fatal("terminal mutation accepted") }
}

func TestExecution_CheckpointsAreMonotonicUnderConcurrency(t *testing.T) {
	// Start one two-checkpoint step, race duplicate checkpoint reports, finish it,
	// then assert Snapshot reports exactly 2/2 and ValidateStream succeeds.
}
```

- [ ] **Step 1.2: Run the focused tests and confirm RED**

Run:

```bash
go test ./internal/progress -run 'TestExecution|TestSnapshot_PreservesErrorAndLabel' -count=1 -v
```

Expected: compile failure because `Start`, stable phase/step IDs, labels, and `StepState.Error` do not exist.

- [ ] **Step 1.3: Implement the stable plan and execution API**

Use these public shapes:

```go
type PhaseID string
type StepID string

type PlannedPhase struct {
	ID    PhaseID
	Label string
	Steps []PlannedStep
}

type PlannedStep struct {
	ID          StepID
	Label       string
	Checkpoints int
}

type Event struct {
	Phase      PhaseID
	PhaseLabel string
	Step       StepID
	StepLabel  string
	Status     StepStatus
	Checkpoint int
	Of         int
	Close      bool
	Message    string
	Error      error
}

type StepResult struct {
	ID      StepID
	Name    string
	Status  StepStatus
	Message string
	Error   error
}

type Execution struct {
	mu     sync.Mutex
	emit   func(Event)
	phases map[PhaseID]*executionPhase
	order  []PhaseID
}

func Start(plan Plan, emit func(Event)) (*Execution, error)
func (x *Execution) Running(phase PhaseID, step StepID, checkpoint int, message string) error
func (x *Execution) Done(phase PhaseID, step StepID, message string) (StepResult, error)
func (x *Execution) Fail(phase PhaseID, step StepID, err error) (StepResult, error)
func (x *Execution) Skip(phase PhaseID, step StepID, reason string) (StepResult, error)
func (x *Execution) Run(phase PhaseID, step StepID, fn StepFunc) (StepResult, error)
func (x *Execution) SkipPending(phase PhaseID, reason string) error
func (x *Execution) Finish(reason string) error
```

`Start` validates nonempty unique IDs, normalizes checkpoint counts to one,
emits the complete Pending declaration prefix, then emits a close marker for
every phase. Transition methods reject unknown IDs, terminal mutations,
checkpoint regression, and checkpoint overflow. `Run` marks Running, executes
the function without holding the mutex, then resolves Done or Failed. `Finish`
resolves every nonterminal step as skipped.

Update `Snapshot` to fold by IDs while preserving declaration labels and errors:

```go
type StepState struct {
	ID       StepID
	Name     string
	Status   StepStatus
	Message  string
	Error    error
	Reached  int
	Declared int
}

type PhaseState struct {
	ID     PhaseID
	Name   string
	Steps  []StepState
	Total  int
	Done   int
	Failed int
	Closed bool
}
```

Strengthen `ValidateStream` to require declarations as one complete prefix,
reject undeclared work and terminal mutation, and retain checkpoint checks.

- [ ] **Step 1.4: Run RED tests to GREEN, then the package suite**

```bash
go test ./internal/progress -count=1
go test -race ./internal/progress -count=1
```

Expected: both commands exit 0.

- [ ] **Step 1.5: Convert existing progress helpers mechanically and commit**

Update `RunStep` and `PhaseRecorder` to accept stable IDs plus labels, preserving
their current behavior until producer-specific tasks replace them. Keep a
temporary `Declare`/`ClosePhase` adapter for unconverted producers; remove it in
Task 5 after the last producer moves to `Execution`. Run:

```bash
go test ./internal/progress ./internal/creator ./internal/integration ./internal/repo ./internal/tui
```

Commit:

```bash
git add internal/progress
git commit -m "feat(progress): make plans own terminal state"
```

## Task 2: Prepare and Freeze Integration Applies

**Files:**
- Create: `internal/integration/apply.go`
- Modify: `internal/integration/manager.go`
- Delete: `internal/integration/plan.go`
- Modify: `internal/tui/integration_list.go`
- Modify: `internal/tui/migrate_integrations.go`
- Modify: `internal/tui/integration_progress.go`
- Test: `internal/integration/apply_test.go`
- Modify: `internal/tui/integration_plan_test.go`

- [ ] **Step 2.1: Write failing preparation and failure-terminalization tests**

Create table tests for these exact cases:

```go
func TestPrepareApply_MissingToolPlansPrerequisiteOnce(t *testing.T) {
	// Two worktrees, one missing global tool, all dependencies present.
	// Assert plan has one Prerequisites install step and one setup step per worktree.
}

func TestPreparedApply_InstallFailureSkipsEverySetup(t *testing.T) {
	// Run the frozen plan with an install error.
	// Assert install failed, all setup steps are skipped with "blocked by Install ...",
	// ValidateStream succeeds, and every Snapshot phase is settled.
}

func TestPreparedApply_DeclarationTotalIsFixedAcrossPrefixes(t *testing.T) {
	// Two integrations x two worktrees with mixed detected/missing tools.
	// After the declaration prefix, assert CheckpointProgress total never changes.
}
```

- [ ] **Step 2.2: Run the tests and confirm RED**

```bash
go test ./internal/integration -run 'TestPrepareApply|TestPreparedApply' -count=1 -v
```

Expected: compile failure because `PrepareApply` and `PreparedApply.Run` do not exist.

- [ ] **Step 2.3: Implement one read-only preparation pass**

Use this contract:

```go
type PreparedApply struct {
	Plan       progress.Plan
	operations []applyOperation
}

func PrepareApply(shell git.ShellRunner, repoPath, mainWT string,
	toEnable, toDisable []Integration, wtPaths []string) (PreparedApply, error)

func (p PreparedApply) Run(shell git.ShellRunner, emit func(progress.Event)) []progress.Phase
```

Probe each enabled tool and dependency once using the first worktree as the
command directory. Put missing dependency installs and tool installs in one
`Prerequisites` phase. Put setup/teardown/removal operations in worktree phases.
Store commands, working directories, IDs, labels, and dependency edges in
`operations`; `Run` executes exactly those decisions and never calls detection.
When a prerequisite fails, skip its dependent install/setup operations. Continue
independent teardown removals after teardown-command failure.

In the TUI goroutine, prepare before setting determinate events. Represent the
probe interval using the existing indeterminate motion vocabulary and the copy
`Preparing plan...`; only call `progress.Start` after preparation succeeds.
Replace `buildIntegrationPhases` with:

```go
func (m Model) buildIntegrationPhases() []progress.PhaseState {
	states := progress.Snapshot(m.integ.events)
	for i := range states { states[i].Name = filepath.Base(states[i].Name) }
	return states
}
```

Use the preserved `StepState.Error` only while rendering; never recompute counts.

- [ ] **Step 2.4: Verify focused and TUI integration tests**

```bash
go test ./internal/integration ./internal/tui -run 'Integration|PreparedApply|PrepareApply' -count=1
go test -race ./internal/integration ./internal/tui -count=1
```

Expected: exit 0 with no phase reopening or pending-on-failure assertions.

- [ ] **Step 2.5: Commit**

```bash
git add internal/integration internal/tui/integration_list.go internal/tui/migrate_integrations.go internal/tui/integration_progress.go internal/tui/integration_plan_test.go
git commit -m "fix(integration): freeze exact apply plans"
```

## Task 3: Convert Creator to One Exact Plan

**Files:**
- Create: `internal/creator/plan.go`
- Modify: `internal/creator/creator.go`
- Modify: `internal/creator/setup.go`
- Modify: `internal/creator/deps.go`
- Modify: `internal/creator/integrations.go`
- Test: `internal/creator/plan_test.go`
- Modify: `internal/creator/setup_test.go`
- Modify: `internal/creator/integrations_test.go`

- [ ] **Step 3.1: Write failing creator contract tests**

Add fixtures asserting:

```go
func TestRun_CreateWorktreeFailureSettlesWholePlan(t *testing.T) {
	// Enable merge, env copy, dependencies, and integrations; fail worktree add.
	// Assert all later steps are skipped with a blocking reason and every phase settles.
}

func TestRun_DuplicateIntegrationLabelsRemainDistinct(t *testing.T) {
	// Two integrations both copy an index. Assert two stable IDs and two rendered rows.
}

func TestRun_NoSetupCommandEmitsSkipped(t *testing.T) {
	// Empty setup command must produce a StepSkipped event, not only a result entry.
}
```

- [ ] **Step 3.2: Confirm RED**

```bash
go test ./internal/creator -run 'SettlesWholePlan|DuplicateIntegrationLabels|NoSetupCommand' -count=1 -v
```

Expected: assertions fail because early returns leave declared work unresolved and duplicate labels fold together.

- [ ] **Step 3.3: Build and execute the creator plan**

Create stable IDs with phase prefixes, for example:

```go
const (
	phaseSetup        progress.PhaseID = "setup"
	phaseDependencies progress.PhaseID = "dependencies"
	phaseIntegrations progress.PhaseID = "integrations"
)

func buildPlan(opts Options) progress.Plan
func integrationStepID(integrationName, operation string) progress.StepID
```

Declare setup, dependency, and integration phases once in `Run`, pass the one
`Execution` through helpers, and call `Finish` on every return. A failed
worktree creation skips merge, env copy, dependencies, and integrations. A
dependency/install failure skips only dependent integration work. Record a
skipped index copy in both the event stream and `Result.Phases`.

- [ ] **Step 3.4: Verify and commit**

```bash
go test ./internal/creator -count=1
go test -race ./internal/creator -count=1
git add internal/creator
git commit -m "fix(creator): settle the declared creation plan"
```

## Task 4: Predeclare Repository Create, Clone, and Migrate

**Files:**
- Create: `internal/repo/plan.go`
- Modify: `internal/repo/create.go`
- Modify: `internal/repo/clone.go`
- Modify: `internal/repo/migrate.go`
- Test: `internal/repo/plan_test.go`
- Modify: `internal/repo/create_test.go`
- Modify: `internal/repo/clone_test.go`
- Modify: `internal/repo/migrate_test.go`

- [ ] **Step 4.1: Write failing table-driven stream-contract tests**

For each existing failure injection point, collect events and call one helper:

```go
func assertFinishedPlan(t *testing.T, events []progress.Event) {
	t.Helper()
	if err := progress.ValidateStream(events); err != nil { t.Fatal(err) }
	for _, phase := range progress.Snapshot(events) {
		if !phase.Settled() { t.Fatalf("phase not settled: %#v", phase) }
	}
}
```

Cover invalid clone target, clone/structure/worktree failures, create setup and
GitHub failures, migrate validation/backup/destructive-stage failures, and the
optional upstream/origin skips.

- [ ] **Step 4.2: Confirm RED**

```bash
go test ./internal/repo -run 'FinishedPlan|DeclaresBeforeRunning' -count=1 -v
```

Expected: failures because repository phases are still discovered step by step.

- [ ] **Step 4.3: Add static repository plan builders**

Implement:

```go
func createPlan(opts CreateOptions) progress.Plan
func clonePlan(opts CloneOptions) progress.Plan
func migratePlan(hasOrigin bool) progress.Plan
```

Create declares Setup plus conditional GitHub phases. Clone declares Validate,
Clone, Structure, and Worktree. Migrate performs only the safe origin-presence
probe before declaration, then declares Validate, Backup, Migrate, and Copy.
Each top-level function starts one execution and defers `Finish("blocked by an earlier phase")`.
Each early-return branch explicitly skips its downstream phases with the most
specific reason available.

- [ ] **Step 4.4: Verify and commit**

```bash
go test ./internal/repo -count=1
go test -race ./internal/repo -count=1
git add internal/repo
git commit -m "fix(repo): seed repository operation plans"
```

## Task 5: Move Removal onto the Shared Event Fold

**Files:**
- Modify: `internal/worktree/deleter.go`
- Modify: `internal/worktree/deleter_test.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/progress.go`
- Create: `internal/tui/removal_progress.go`
- Modify: `internal/tui/removal_plan_test.go`
- Modify: `internal/tui/removal_e2e_test.go`

- [ ] **Step 5.1: Write failing checkpoint and Snapshot parity tests**

```go
func TestDeleteWorktrees_EmitsStartAndTerminalCheckpoints(t *testing.T) {
	// Two removals with concurrency two. Assert declarations precede work,
	// Running reaches checkpoint 1/2, terminals reach 2/2, and phases settle.
}

func TestRemovalProgress_UsesSnapshotAsSourceOfTruth(t *testing.T) {
	// Feed teardown/removal/prune/cleanup events to the model and assert the
	// layout phases equal progress.Snapshot(eventHistory).
}
```

- [ ] **Step 5.2: Confirm RED**

```bash
go test ./internal/worktree ./internal/tui -run 'StartAndTerminalCheckpoints|SnapshotAsSourceOfTruth' -count=1 -v
```

Expected: current deleter has no declaration/checkpoint events and TUI rebuilds phase state manually.

- [ ] **Step 5.3: Emit the complete removal plan and retain event history**

At confirmation time build phases for teardown, removal (two checkpoints per
worktree), prune, and cleanup. Store events on `removalRun`. Pass the shared
`Execution` to `DeleteWorktrees`; report checkpoint one before `remover`, then
Done/Failed after it returns. Teardown failure remains independent of removal;
prune failure remains independent of cleanup where current safety permits.

Replace `buildRemovalPhases` with `progress.Snapshot(run.events)`. Preserve
display labels via declaration labels, not a second worktree-status fold.
Now that integration, creator, repository, and removal producers use
`Execution`, delete `PlannedPhase.Open`, the legacy `Declare`/`ClosePhase`
adapter, and tests that permit discovery after declaration.

- [ ] **Step 5.4: Verify concurrency and commit**

```bash
go test ./internal/worktree ./internal/tui -run 'Removal|DeleteWorktrees' -count=1
go test -race ./internal/worktree ./internal/tui -run 'Removal|DeleteWorktrees' -count=1
git add internal/worktree internal/tui/model.go internal/tui/progress.go internal/tui/removal_progress.go internal/tui/removal_plan_test.go internal/tui/removal_e2e_test.go
git commit -m "refactor(progress): fold removal from shared events"
```

## Task 6: Build the Pinned Responsive Live Region

**Files:**
- Create: `internal/tui/progress_viewport.go`
- Create: `internal/tui/progress_viewport_test.go`
- Modify: `internal/tui/progress_layout.go`
- Modify: `internal/tui/window.go`
- Modify: `internal/tui/window_test.go`
- Modify: `internal/tui/model.go`
- Create: `internal/tui/progress_details.go`
- Modify: `internal/tui/help.go`
- Modify: `internal/tui/keys.go`
- Modify: `internal/tui/progress_layout_test.go`

- [ ] **Step 6.1: Write failing viewport and dimension properties**

```go
func TestProgressLayout_NeverExceedsTerminal(t *testing.T) {
	for height := 8; height <= 40; height++ {
		for _, width := range []int{40, 50, 80, 120} {
			out := denseProgressFixture(width, height).View()
			if got := lipgloss.Height(out); got > height { t.Fatalf("%dx%d rendered %d rows", width, height, got) }
			for _, line := range strings.Split(out, "\n") {
				if got := lipgloss.Width(line); got > width { t.Fatalf("%dx%d rendered %d cols", width, height, got) }
			}
		}
	}
}

func TestBuildProgressViewport_ChoosesOneFocus(t *testing.T) {
	// Settled history + two pending phases + one running phase.
	// Assert one focus, static queued summary, and bounded detail rows.
}

func TestWindowSteps_HardBoundsRunningAndFailures(t *testing.T) {
	// More active/failure steps than available lines must still return <= budget.
}
```

- [ ] **Step 6.2: Confirm RED**

```bash
go test ./internal/tui -run 'NeverExceedsTerminal|BuildProgressViewport|HardBounds' -count=1 -v
```

Expected: 80x24 height assertion fails and pending phases are classified as active.

- [ ] **Step 6.3: Implement the pure viewport projection**

```go
type ProgressViewport struct {
	History        []progress.PhaseState
	HistoryOmitted int
	Focus          *progress.PhaseState
	Queued         int
	DetailRows     int
	Tier           progressViewportTier
}

func BuildProgressViewport(phases []progress.PhaseState, rows int) ProgressViewport
```

A phase is active only when one step is `StepRunning`. Focus order is latest
failed phase, running phase, then the earliest unresolved phase. Normal (18+),
compact (12-17), and minimal (<12) tiers reserve fixed rows. History keeps the
newest settled/failed rows; omitted rows become one summary. Queue becomes one
static `N phases waiting` row.

- [ ] **Step 6.4: Render fixed-height regions and add details**

Pass raw `WindowSizeMsg.Height` to progress layouts rather than the globally
chrome-budgeted body height. Render title, separator, history, a padded focus
region, queue, separator, bar, and footer within the tier's exact budget.
Build skip text before truncation so suffixes cannot overflow. Clamp the bar to
remaining width; omit elapsed first, then use percentage-only below the useful
bar floor.

Add progress views to `detailContent` and add `keys.Info` to their footer only
when event history contains failures or omitted detail. Render full messages and
errors in the existing portal viewport.

- [ ] **Step 6.5: Verify resize and rendering tests, then commit**

```bash
go test ./internal/tui -run 'Progress|Window|Resize|Detail' -count=1
go test -race ./internal/tui -count=1
git add internal/tui
git commit -m "feat(tui): pin the progress live region"
```

## Task 7: Scope Motion and Repair the Identity Guard

**Files:**
- Create: `internal/tui/motion_preference.go`
- Create: `internal/tui/motion_preference_test.go`
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/progress_layout.go`
- Modify: `internal/tui/motion_test.go`
- Modify: `scripts/pre-commit`
- Create: `scripts/pre-commit_test.sh`
- Modify: `.impeccable.md`

- [ ] **Step 7.1: Write failing motion and hook tests**

```go
func TestMotionPreference_OffDisablesProgressTicks(t *testing.T) {
	t.Setenv("SENTEI_MOTION", "off")
	m := newProgressModel()
	if m.motionActive() { t.Fatal("motion remains active") }
	if cmd := m.motionCmd(); cmd != nil { t.Fatal("motion tick scheduled") }
}

func TestProgressMotion_OnlyRunningPhaseShimmers(t *testing.T) {
	// Render several declared pending phases and one running phase.
	// Assert the injected star frame appears exactly once.
}
```

The shell test creates an isolated temporary repository and runs the hook with:

```bash
GIT_AUTHOR_NAME=sentei-test GIT_AUTHOR_EMAIL=test@sentei.invalid \
GIT_COMMITTER_NAME='Valid User' GIT_COMMITTER_EMAIL=valid@example.com \
scripts/pre-commit
```

Expected: nonzero. Reverse author/committer and expect nonzero. Use two valid
identities and expect zero when no Go files are staged.

- [ ] **Step 7.2: Confirm RED**

```bash
go test ./internal/tui -run 'MotionPreference|OnlyRunningPhase' -count=1 -v
bash scripts/pre-commit_test.sh
```

Expected: Go compile failure for the preference API and hook test failure because configured email masks environment identity.

- [ ] **Step 7.3: Implement static motion and effective identity checks**

```go
type MotionPreference int
const (
	MotionFull MotionPreference = iota
	MotionOff
)

func motionPreference(getenv func(string) string) MotionPreference {
	if strings.EqualFold(getenv("SENTEI_MOTION"), "off") || getenv("TERM") == "dumb" { return MotionOff }
	return MotionFull
}
```

Store the preference on `Model`. In static mode do not schedule `motionTickMsg`
or progress spring frames, render the exact target bar, keep stopwatch ticks,
and let settle observation treat displayed progress as target progress. In full
mode, inject Motion only into the one running focus phase; pending/history rows
remain static. Keep the existing completion crystallization and success-green
bar, without extra transitions.

In the hook, obtain both identities with:

```bash
author_ident="$(git var GIT_AUTHOR_IDENT)"
committer_ident="$(git var GIT_COMMITTER_IDENT)"
case "$author_ident" in *sentei-test*|*"@sentei.invalid"*) reject="author" ;; esac
case "$committer_ident" in *sentei-test*|*"@sentei.invalid"*) reject="committer" ;; esac
```

Report only which role is invalid and how to inspect it; allow every other
contributor or automation identity. Record the env-first reduced-motion decision
in `.impeccable.md`.

- [ ] **Step 7.4: Verify and commit**

```bash
go test ./internal/tui -run 'Motion|Settle|Progress' -count=1
bash scripts/pre-commit_test.sh
git add internal/tui scripts/pre-commit scripts/pre-commit_test.sh .impeccable.md
git commit -m "fix(tui): scope motion and guard effective git identity"
```

## Task 8: Add End-to-End Contracts and Deterministic VHS Demos

**Files:**
- Modify: `internal/tui/integration_plan_test.go`
- Modify: `internal/tui/removal_e2e_test.go`
- Create: `internal/tui/progress_teatest_test.go`
- Create: `.demos/progress-arc/setup-fixture.sh`
- Create: `.demos/progress-arc/before-after.tape`
- Create: `.demos/progress-arc/removal-success.tape`
- Create: `.demos/progress-arc/integration-failure.tape`
- Create: `.demos/progress-arc/README.md`

- [ ] **Step 8.1: Write last-frame and monotonic integration tests**

Use `teatest` to drive success and failure models at 80x24. Capture output until
summary, then assert the last progress frame contains the bar/footer, resolved
blocked steps, and no pending active glyphs. For every stream prefix assert the
checkpoint total is unchanged after the declaration prefix and reached never
decreases.

- [ ] **Step 8.2: Confirm the new tests catch the old branch behavior**

Build the current commit's tests, then temporarily run the same assertions
against the parent implementation or use the retained before binary in the VHS
fixture. Record the expected old failure: missing footer/overflow, denominator
growth, or unresolved setup. Restore the working tree before continuing.

- [ ] **Step 8.3: Make isolated deterministic VHS fixtures**

`setup-fixture.sh` must create everything under `/tmp/sentei-vhs-progress-arc`:

- isolated HOME and Git config with `sentei-demo <demo@sentei.invalid>`;
- a bare repository and fixed worktrees/branches;
- a `git` PATH shim that sleeps only for `worktree remove` and delegates;
- `code-review-graph` and `ccc` shims with deterministic detection/setup;
- a failing `ccc index` that prints `error: deterministic demo failure` and exits 17.

Each tape must use `Hide`, reset the fixture, run `stty cols 80 rows 24`, set the
isolated HOME/PATH, then `Show` before launching Sentei. Prefer `Wait+Screen` for
semantic states and short sleeps only to expose the completion beat.

- [ ] **Step 8.4: Render and inspect GIFs plus checkpoint frames**

```bash
vhs .demos/progress-arc/before-after.tape
vhs .demos/progress-arc/removal-success.tape
vhs .demos/progress-arc/integration-failure.tape
ffprobe -v error -show_entries stream=width,height,r_frame_rate -of default=nw=1 .demos/progress-arc/removal-success.gif
ffmpeg -y -i .demos/progress-arc/removal-success.gif -vf "select=eq(n\,0)+eq(n\,25)+eq(n\,50)" -vsync 0 /tmp/sentei-vhs-progress-arc/frame-%02d.png
```

Inspect the GIFs and extracted frames for pinned footer/bar rows, monotonic fill,
one active animation, readable failure/skip copy, and coherent final frames.

- [ ] **Step 8.5: Run the full gauntlet, refresh semantic index, and commit**

```bash
go test -race ./...
go vet ./...
golangci-lint run
go build ./...
git diff --check
ccc index
git add internal/tui .demos/progress-arc
git commit -m "test: prove truthful progress end to end"
```

Expected: every command exits 0; GIF and frame inspection finds no clipping or
regression.

## Task 9: Final Parallel Review and PR Readiness

**Files:**
- Modify only files required by validated review findings.

- [ ] **Step 9.1: Dispatch independent correctness, architecture, and UX reviews**

Give each reviewer the design spec, this plan, base SHA `8a2e202`, and current
HEAD. Require file/line evidence and reproduction commands. Correctness focuses
on stream invariants and failure paths; architecture checks domain dependency
direction and duplicate sources of truth; UX checks 80x24, resize tiers, motion,
portal detail, and the GIFs.

- [ ] **Step 9.2: Fix every Critical or Important finding test-first**

For each accepted finding, write a regression test, observe RED, implement the
smallest fix, observe GREEN, and return it to the reporting reviewer for
re-review. Do not merge unrelated pre-existing CLI/docs issues into this branch.

- [ ] **Step 9.3: Run fresh final verification**

```bash
go test -race ./...
go vet ./...
golangci-lint run
go build ./...
git diff --check origin/main...HEAD
git status --short --branch
```

Verify the worktree is clean and only then summarize PR readiness, remaining
non-blocking debt, and links to the GIFs.
