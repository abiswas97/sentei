## 1. Package foundation

- [x] 1.1 Delete dead `internal/progress/tracker.go` and `tracker_test.go`; create `internal/progress` vocabulary: `StepStatus` (Pending, Running, Done, Failed, Skipped, ordering preserved from pipeline) and `Event{Phase, Step, Status, Message, Error}`, with an invariant test pinning enum ordering
- [x] 1.2 Move `RunStep`, `PhaseRecorder`, `Phase`, `StepResult`, `HasFailures`, `PhasesHaveFailures`, `FirstFailure` from `internal/pipeline` into `internal/progress` with their tests
- [x] 1.3 Move the fold from `internal/tui/phase_display.go` into `internal/progress` as exported `Snapshot(events) []PhaseState` and `WithPendingPhases`, exporting `PhaseState`/`StepState`; move and extend the fold tests (determinism, order preservation, skipped-counts-as-resolved, later-status-supersedes)

## 2. Mechanical migration

- [x] 2.1 Rename all `internal/pipeline` imports to `internal/progress` across `internal/repo`, `internal/creator`, `internal/tui`, `cmd` (no logic edits in this task); delete `internal/pipeline`
- [x] 2.2 Rewire `internal/tui` consumers of the fold (`progress.go`, `ProgressLayout` call sites) to `progress.Snapshot`/`progress.PhaseState`; delete `internal/tui/phase_display.go`
- [x] 2.3 Gauntlet checkpoint: gofmt, go vet, go test -race ./..., golangci-lint; golden tests must pass without `-update`

## 3. Integration dialect replacement

- [x] 3.1 Replace `integration.ManagerEvent` emission with `progress.Event` (worktree name maps to `Event.Phase` at emit time); update manager tests
- [x] 3.2 Rewire `internal/tui` integration consumers (`integration_progress.go`, `migrate_integrations.go`, `model.go` channels) to `progress.Event`; delete the `ManagerEvent` type
- [x] 3.3 Table-driven test: folding the mapped events reproduces today's `buildIntegrationPhases` display state for a representative apply scenario (success, failure, skip)

## 4. Removal dialect replacement

- [x] 4.1 Replace `worktree.DeletionEvent` emission with `progress.Event` (phase `Removing worktrees`, step = path, Running/Done/Failed); update deleter tests
- [x] 4.2 Rewire the removal flow in `internal/tui/progress.go` to consume `progress.Event`, deleting `worktreeDeleteStartedMsg`, `worktreeDeletedMsg`, `worktreeDeleteFailedMsg`; preserve Cmd-chained one-event-per-Msg consumption
- [x] 4.3 Run the full removal E2E suite unmodified; verify locked-worktree, missing-directory, and failure-path edge cases pass

## 5. Verification

- [x] 5.1 Repo-wide search proves no `pipeline.`, `ManagerEvent`, or `DeletionEvent` references remain; `internal/progress` is the only `StepStatus` definition
- [x] 5.2 Full gauntlet + golden tests byte-identical + commitlint on all commits
- [x] 5.3 Update `.impeccable.md` decision log with the consolidation entry
