## 1. Foundation: Constants and Key Bindings

- [x] 1.1 Create `internal/tui/constants.go` with `MinProgressDisplay` (300ms, `var` for test override), `WindowCompletedTrail` (2), `WindowPendingLead` (1), and progress bar width (20)
- [x] 1.2 Add contextual `keyDetails` (`?`) and global `keyGlobalHelp` (`F1`) bindings to `internal/tui/keys.go`
- [x] 1.3 Write unit tests for constants accessibility and key binding definitions

## 2. Chrome Helpers

- [x] 2.1 Write failing tests for `viewTitle`, `viewSeparator`, `viewKeyHints` in `internal/tui/chrome_test.go` — cover standard cases, edge cases (empty title, narrow width, single hint, no hints)
- [x] 2.2 Create `internal/tui/chrome.go` with `KeyHint` type and `viewTitle()`, `viewSeparator()`, `viewKeyHints()` implementations to pass tests
- [x] 2.3 Write failing tests for `viewStatLine` — standard stats, zero count omission, with failures
- [x] 2.4 Implement `viewStatLine(stats WindowStats) string` to pass tests

## 3. Windowing

- [x] 3.1 Write failing tests for `WindowSteps` in `internal/tui/window_test.go` — all-fit, windowed, failed-pinned, active-pinned, budget-zero, parallel active, exact boundary
- [x] 3.2 Create `internal/tui/window.go` with `ProgressStep`, `WindowResult`, `WindowStats` types and `WindowSteps()` implementation to pass tests
- [x] 3.3 Write responsive windowing tests: same 30 items at heights 60, 30, 20, 15 — verify windowed/not-windowed and showing counts

## 4. Progress Layout

- [x] 4.1 Write failing tests for `ProgressLayout.View()` in `internal/tui/progress_layout_test.go` — with/without subtitle, completed phase collapse, active phase expansion, pending phase, failed phase keeps steps, overall bar percentage, step indentation (4-space)
- [x] 4.2 Create `internal/tui/progress_layout.go` with `ProgressPhase`, `ProgressLayout` types and `View()` implementation to pass tests — integrates chrome helpers, windowing, stat line, and progress bar
- [x] 4.3 Write tests for `bufferTransition` — fast op buffered, slow op not delayed, zero MinProgressDisplay passthrough
- [x] 4.4 Implement `bufferTransition(started time.Time, cmd tea.Cmd) tea.Cmd` in `progress_layout.go`

## 5. Migrate Progress Views

- [x] 5.1 Write adapter tests: `viewProgress()` builds correct `ProgressLayout` from removal state (phases, step statuses, windowing at different heights)
- [x] 5.2 Refactor `viewProgress()` in `progress.go` to build a `ProgressLayout` and call `.View()` — delete `styleHeader` usage
- [x] 5.3 Update existing `progress_test.go` tests to match new rendering output (● instead of old indicators, chrome format)
- [x] 5.4 Write adapter tests: `viewCreateProgress()` builds correct `ProgressLayout` from creator events
- [x] 5.5 Refactor `viewCreateProgress()` in `create_progress.go` to use `ProgressLayout`
- [x] 5.6 Write adapter tests: `viewRepoProgress()` builds correct `ProgressLayout` from repo events
- [x] 5.7 Refactor `viewRepoProgress()` in `repo_progress.go` to use `ProgressLayout`
- [x] 5.8 Write adapter tests: `viewIntegrationProgress()` builds correct `ProgressLayout` from integration events
- [x] 5.9 Refactor `viewIntegrationProgress()` in `integration_progress.go` to use `ProgressLayout`

## 6. Migrate Summary Views

- [x] 6.1 Write tests for `viewSummary()` with new chrome — title, separators, key hints, `●` success marker
- [x] 6.2 Refactor `viewSummary()` in `summary.go` to use chrome helpers — replace `"v"` with `●`, add missing separator after title
- [x] 6.3 Write tests for `viewCreateSummary()` with new chrome
- [x] 6.4 Refactor `viewCreateSummary()` in `create_summary.go` to use chrome helpers
- [x] 6.5 Write tests for `viewCreateRepoSummary()` and `viewCloneRepoSummary()` with new chrome
- [x] 6.6 Refactor `viewRepoSummary()` variants in `repo_summary.go` to use chrome helpers
- [x] 6.7 Write tests for `viewMigrateSummary()` and `viewMigrateNext()` with new chrome
- [x] 6.8 Refactor `viewMigrateSummary()` and `viewMigrateNext()` in `migrate_summary.go` to use chrome helpers
- [x] 6.9 Write tests for `viewCleanupResult()` with new chrome
- [x] 6.10 Refactor `viewCleanupResult()` in `cleanup_result.go` to use chrome helpers

## 7. Migrate Confirmation Views

- [x] 7.1 Write tests for `ConfirmationViewModel.View()` with standard chrome (no border)
- [x] 7.2 Refactor `ConfirmationViewModel.View()` in `confirmation.go` to use `viewTitle`, `viewSeparator`, `viewKeyHints` — remove `styleDialogBox.Render()` wrapping
- [x] 7.3 Update `cleanup_confirm_test.go`, `create_confirm_test.go`, `clone_confirm_test.go`, `migrate_confirm_test.go` to match new rendering

## 8. Cleanup Dead Code

- [x] 8.1 Delete `styleHeader` and `styleDialogBox` from `styles.go`
- [x] 8.2 Delete old `separator()` function (replaced by `viewSeparator`)
- [x] 8.3 Verify no remaining references to deleted styles — `go vet ./...` and `go build`

## 9. Design System Documentation

- [x] 9.1 Add "Component Patterns" section to `.impeccable.md` covering: view chrome, progress views, status indicators table, stat line, timing, key mapping
- [x] 9.2 Review all views to verify consistency with documented patterns

## 10. E2E Tests

- [x] 10.1 Update `cleanup_e2e_test.go` to verify new chrome in cleanup flow
- [x] 10.2 Write E2E test for removal flow: select → progress with title + bar → summary with chrome
- [x] 10.3 Write E2E test verifying ctrl+c/q quits from all progress views
- [x] 10.4 Write E2E test for WindowSizeMsg updating adaptive windowing in removal progress

## 11. Verification

- [x] 11.1 Run `go fmt ./...` and `go vet ./...`
- [x] 11.2 Run `go test ./...` — all tests pass
- [x] 11.3 Run `go build` — binary builds
- [x] 11.4 Manual visual check with `--playground` flag against `.impeccable.md` patterns
- [x] 11.5 Update session meta doc (`docs/handoff/session-meta.md`) with completion status

## 12. Visual Testing Fixes (from playground review)

### 12.1 Cleanup loading title mismatch
- [x] 12.1.1 Fix `viewCleanupResult()` in `cleanup_result.go` — title says "Cleanup Complete" while body says "Running cleanup…". When `cleanupResult == nil`, use `viewTitle("Running Cleanup")` instead; only show "Cleanup Complete" after result is available
- [x] 12.1.2 Write test: `viewCleanupResult()` with nil result renders title "Running Cleanup", not "Cleanup Complete"

### 12.2 Confirm deletion — missing branch name for detached HEAD
- [x] 12.2.1 Fix `viewConfirm()` in `confirm.go` — when `stripBranchPrefix(wt.Branch)` returns empty string (detached HEAD), fall back to showing the worktree directory name or commit SHA
- [x] 12.2.2 Write test: confirmation view with detached HEAD worktree shows a meaningful identifier, not blank

### 12.3 Integration progress — premature 100% on unstarted worktrees
- [x] 12.3.1 Fix `buildIntegrationLayout()` in `integration_progress.go` — phases with Done==0 and Total==0 should not render as complete. The `ProgressLayout` renderer treats 0/0 as "100% ●" but it should show as pending. Fix in `renderPhase` in `progress_layout.go`: a phase with Total==0 is pending, not complete
- [x] 12.3.2 Write test: `ProgressLayout.View()` with a phase that has Done==0, Total==0, and non-empty Steps should render as pending, not "100% ●"

### 12.4 Menu refresh hint after mutations
- [x] 12.4.1 When `stateStale` triggers a reload in `updateMenu()`, set the menu hint to "refreshing…" instead of keeping the stale count visible
- [x] 12.4.2 Write test: after `stateStale=true` and a `WindowSizeMsg`, the menu hint shows "refreshing…" until `worktreeContextMsg` arrives

### 12.5 Playground delay too slow for menu refresh
- [x] 12.5.1 Change playground `DelayRunner` to only apply delay during progress operations (create, remove, integrate), not during worktree listing/enrichment. OR reduce delay from 800ms to a lower value for list operations
- [x] 12.5.2 Verify playground menu loads in <1s after returning from a mutation

## 13. Re-verification

- [x] 13.1 Run `go fmt ./...` and `go vet ./...`
- [x] 13.2 Run `go test ./...` — all tests pass
- [x] 13.3 Run `go build` — binary builds
- [ ] 13.4 Manual visual recheck with `--playground` — verify all 5 fixes

## 14. Round 2 Visual Fixes (from second playground review)

- [x] 14.1 Fix PANIC: `renderProgressBar` negative Repeat count — clamp `filled` and `pct` to max values
- [x] 14.2 Fix phase header indicator position — move `●` from after `%` to before phase name (`● Phase Name  100%`)
- [x] 14.3 Fix integration progress premature 100% — add `targetWorktrees` to model, pre-populate all as pending before events
- [x] 14.4 Fix cleanup prune-refs error on playground — check `git remote get-url origin` before pruning

### Tests added
- [x] `TestProgressLayout_DoneExceedsTotal_NoPanic` — panic regression
- [x] `TestProgressLayout_PhaseIndicatorOnLeft` — indicator column alignment
- [x] `TestViewIntegrationProgress_PrePopulatesTargetWorktrees` — all targets visible from start
- [x] `TestPruneRemoteRefs_NoOriginRemote` — graceful skip when no remote

## 15. Round 2 Verification

- [x] 15.1 `go fmt`, `go vet` — pass
- [x] 15.2 `go test ./...` — all pass
- [x] 15.3 `go build` — builds
- [ ] 15.4 Manual visual recheck with `--playground`
