## ADDED Requirements

### Requirement: Plan declaration compiles into the stream
`internal/progress` SHALL provide a typed plan (phases containing steps, each step declaring a checkpoint count of at least 1) and a `Declare` operation that compiles the plan into the event stream as one Pending event per planned step (carrying the step's declared checkpoint count) followed by a close marker for every phase not marked Open, so that declaration, progress, and history are the same data.

#### Scenario: Declared plan establishes totals before work starts
- **WHEN** a flow declares a plan with phase `feat-1` containing 2 steps and then begins work
- **THEN** the folded display state shows `feat-1` with total 2 and done 0 before any Running event is emitted

#### Scenario: Open phase accepts later steps until closed
- **WHEN** a plan declares a scan phase as Open, work appends a discovered step, and the flow then emits the phase close
- **THEN** the fold accepts the appended step (total grows) before close, and the phase can render settled only after the close

### Requirement: Checkpoint progress within steps
A Running event MAY carry `Checkpoint k of n` meaning the step has reached sub-stage k of its n declared checkpoints; the fold SHALL track reached checkpoints per step, clamped to the declared count, monotonically non-decreasing, and SHALL treat a step's resolution (Done, Failed, Skipped) as reaching its final checkpoint.

#### Scenario: Step start moves checkpoint progress
- **WHEN** a step declared with 2 checkpoints emits Running with checkpoint 1 of 2
- **THEN** the fold reports 1 of the step's 2 checkpoints reached while the step's status remains Running

#### Scenario: Checkpoints never regress
- **WHEN** events arrive reporting checkpoint 2 of 3 followed by a stale checkpoint 1 of 3
- **THEN** the folded reached count remains 2

### Requirement: Honesty invariants
The fold SHALL enforce: a phase's total never decreases; done never exceeds total; an undeclared phase with no resolved steps never reports completion; and in test builds, an event introducing a previously unseen step after its phase's close marker is flagged as an invariant violation (work events for already declared steps legitimately follow the close).

#### Scenario: Totals are monotonic
- **WHEN** any valid or invalid event interleaving is folded
- **THEN** the sequence of per-phase totals observed across successive folds of growing prefixes is non-decreasing

#### Scenario: Event after close is flagged in tests
- **WHEN** a test folds a stream containing an event that introduces a new step after that phase's close marker
- **THEN** the invariant check reports the violation

### Requirement: Settled is a single predicate
Phase display state SHALL expose one settled predicate, true exactly when the phase is closed and all declared steps are resolved, and all done-treatments of a phase (collapse, done indicator, success styling) SHALL derive from this predicate.

#### Scenario: Completed-but-open phase is not settled
- **WHEN** a phase's done count equals its current total but the phase has not closed
- **THEN** the settled predicate is false and the phase renders as in-progress, so a later-discovered step never reopens a settled phase

### Requirement: Flows declare what they know
Integration apply SHALL declare staged-changes-per-worktree upfront; teardown SHALL declare its artifact-removal step count upfront; worktree removal SHALL declare per-worktree steps whose checkpoint counts give start credit; repo create, clone, and migrate SHALL declare the step lists they know at flow start, leaving genuinely undiscoverable phases Open.

#### Scenario: Apply phases never reopen
- **WHEN** two integrations are applied across two worktrees
- **THEN** each worktree phase declares total 2 at start, progresses 0/2 to 2/2 monotonically, and renders settled exactly once

#### Scenario: Teardown counts are real
- **WHEN** teardown removes artifacts across N worktrees
- **THEN** the Teardown phase's displayed total equals the declared artifact-removal count from its first rendered frame

#### Scenario: Parallel removal moves the bar at start
- **WHEN** three parallel worktree removals start and none has completed
- **THEN** checkpoint progress is nonzero (start checkpoints reached) while all step statuses remain Running
