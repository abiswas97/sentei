## 1. Removal run state split

- [x] 1.1 Write failing regression test: two deletion runs in one session — second run starts at 0%, completes at 100%, transitions to summary (reproduces the 200%/hang bug)
- [x] 1.2 Extract per-run fields from `removeState` into a `removalRun` struct (`deletionResult`, `deletionStatuses`, `deletionTotal`, `progressCh`, `teardownResults`, `pruneErr`, `cleanupResult`); construct via `newRemovalRun(selected)` in the confirm enter handler
- [x] 1.3 Update `progress.go` and `summary.go` accessors to read through `m.remove.run`
- [x] 1.4 Write failing test: selection cleared after a completed run (`0 selected`, no checked rows on return to list)
- [x] 1.5 Clear the `selected` map after a completed run and on menu entry to the removal flow
- [x] 1.6 Test: pending prune/cleanup phase renders pending on a second run while removal is active
- [x] 1.7 Make the teardown phase visible while running (active indicator) instead of appearing idle on pending rows — found during implementation; the observed "hang" was an invisible, unbounded teardown

## 2. Create flow reset-on-entry

- [x] 2.1 Write failing test: after a completed create, re-entering the flow shows an empty branch input with placeholder; typing produces exactly the typed text
- [x] 2.2 Write failing test: after an abandoned create (esc mid-flow), re-entering shows pristine inputs
- [x] 2.3 Implement `resetCreateFlow()` on `Model` (clear branch input, restore base input to default branch, reset staged options); call it from the menu's create transition
- [x] 2.4 Audit other menu transitions (repo create, clone, migrate name inputs) for the same leak; add reset + test for any found

## 3. Integration apply summary

- [x] 3.1 Write failing test: apply with all steps succeeding transitions to `integrationSummaryView` showing per-worktree outcomes and counts
- [x] 3.2 Write failing test: apply with failed steps shows failed steps with error text under their worktrees
- [x] 3.3 Write failing test: save failure shows the save error prominently and marks the set as not persisted
- [x] 3.4 Add `integrationSummaryView` state and `internal/tui/integration_summary.go` rendering from collected `m.integ.events` (pure render function over event data)
- [x] 3.5 Route `integrationFinalizedMsg` to the summary view (except `migrateNextView` hand-off, which keeps its existing transition); dismissing the summary returns to the list via `loadIntegrationState()`
- [x] 3.6 Write failing test: after dismissing the summary, staged markers match persisted state on both success and save-failure paths (no surviving `[+]`/`[-]`, no pending counter)
- [x] 3.7 Test: migrate-flow apply still proceeds to migrate summary unchanged

## 4. Playground isolation

- [x] 4.1 Write failing test: two concurrent `Setup()` calls return distinct directories, both functional
- [x] 4.2 Replace fixed `PlaygroundDir` with `os.MkdirTemp(os.TempDir(), "sentei-playground-*")` inside `Setup()`; drop the startup `RemoveAll`; cleanup function removes the session's own directory
- [x] 4.3 Update `main.go` and playground tests for the returned-path API; remove the exported fixed-path var
- [x] 4.4 Verify `--playground` end-to-end: launch two sessions simultaneously, confirm isolation

## 5. Verification

- [x] 5.1 `go fmt ./...`, `go vet ./...`, full `go test ./...` green
- [x] 5.2 Manual playground recheck of all four fixed behaviors (second removal run, re-entered create flow, failed apply summary, parallel sessions)
- [x] 5.3 Run commitlint + full test suite before push (pre-push discipline)
