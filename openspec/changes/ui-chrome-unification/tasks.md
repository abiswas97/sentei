> **Reset 2026-06-10.** A previous session checked these boxes but the implementation was never committed (no chrome/window/layout files exist in any ref). Tasks restructured for today's main: `internal/pipeline` + `buildPhaseDisplays` + `holdOrAdvance` now exist and are built upon; the original `bufferTransition`/`MinProgressDisplay` tasks are superseded by the landed timing work. Depends on `flow-state-correctness` landing first.

## 1. Foundation: Constants and Key Bindings

- [ ] 1.1 Create `internal/tui/constants.go` with `WindowCompletedTrail` (2), `WindowPendingLead` (1), and progress bar width (20)
- [ ] 1.2 Add contextual `keyDetails` (`?`) and global `keyGlobalHelp` (`F1`) bindings to `internal/tui/keys.go`
- [ ] 1.3 Write unit tests for constants accessibility and key binding definitions

## 2. Chrome Helpers

- [ ] 2.1 Write failing tests for `viewTitle`, `viewSeparator`, `viewKeyHints` in `internal/tui/chrome_test.go` — standard cases, edge cases (empty title, narrow width, single hint, no hints); hints always join with `·`
- [ ] 2.2 Create `internal/tui/chrome.go` with `KeyHint` type and `viewTitle()`, `viewSeparator()`, `viewKeyHints()` implementations to pass tests
- [ ] 2.3 Write failing tests for `truncateWithEllipsis(s string, width int)` — long path, exact fit, width smaller than ellipsis, multi-byte runes
- [ ] 2.4 Implement `truncateWithEllipsis` in `chrome.go`; it is the only sanctioned way to fit overflowing paths/branches/errors
- [ ] 2.5 Write failing tests for `viewStatLine` — standard stats, zero count omission, with failures
- [ ] 2.6 Implement `viewStatLine(stats WindowStats) string` to pass tests

## 3. Windowing

- [ ] 3.1 Write failing tests for `WindowSteps` in `internal/tui/window_test.go` — all-fit, windowed, failed-pinned, active-pinned, budget-zero, parallel active, exact boundary
- [ ] 3.2 Create `internal/tui/window.go` with `ProgressStep`, `WindowResult`, `WindowStats` types and `WindowSteps()` implementation to pass tests
- [ ] 3.3 Write responsive windowing tests: same 30 items at heights 60, 30, 20, 15 — verify windowed/not-windowed and showing counts

## 4. Progress Layout

- [ ] 4.1 Write failing tests for `ProgressLayout.View()` in `internal/tui/progress_layout_test.go` — with/without subtitle, completed phase collapse, active phase expansion, pending phase, failed phase keeps steps, overall bar percentage, step indentation (4-space), phase indicator left of the phase name (`● Phase Name  100%`)
- [ ] 4.2 Write failing test: a phase with Done==0 and Total==0 renders as pending, never `100% ●`
- [ ] 4.3 Write failing test: done exceeding total clamps the bar at 100% and does not panic (negative `strings.Repeat` regression)
- [ ] 4.4 Write failing test: bar renders filled cells in accent 62 and track cells in dim 241 (never uncolored)
- [ ] 4.5 Write failing test: step error text is truncated with `…` to the available width, never hard-clipped
- [ ] 4.6 Create `internal/tui/progress_layout.go` with `ProgressLayout` consuming `phaseDisplay`-shaped data (per design D4), integrating chrome helpers, windowing, stat line, and the styled progress bar — pass all the above

## 5. Migrate Progress Views

- [ ] 5.1 Write adapter tests: `viewProgress()` builds correct `ProgressLayout` from removal run state (phases, step statuses, windowing at different heights)
- [ ] 5.2 Refactor `viewProgress()` in `progress.go` to build a `ProgressLayout` and call `.View()` — delete `styleHeader` usage, restore the `sentei ─ Removing Worktrees` title and key hints
- [ ] 5.3 Update existing `progress_test.go` tests to match new rendering output
- [ ] 5.4 Write adapter tests + refactor `viewCreateProgress()` in `create_progress.go` to `buildPhaseDisplays` → `ProgressLayout`
- [ ] 5.5 Write adapter tests + refactor `viewRepoProgress()` in `repo_progress.go` the same way
- [ ] 5.6 Write adapter tests + refactor `viewIntegrationProgress()` in `integration_progress.go` the same way
- [ ] 5.7 Pre-populate integration progress with all target worktrees as pending phases before events arrive (no premature jump from empty to mid-run); test: all targets visible from the first frame
- [ ] 5.8 Verify pending sub-steps render `·`, never `●` (create flow currently reuses `●` for pending); covered by 4.1 fixtures

## 6. Migrate Summary Views

- [ ] 6.1 Write tests for `viewSummary()` with new chrome — title, separators, key hints, `●` success marker, no empty `Cleanup:` section header when there is nothing to report
- [ ] 6.2 Refactor `viewSummary()` in `summary.go` to use chrome helpers and pass 6.1
- [ ] 6.3 Write tests + refactor `viewCreateSummary()` in `create_summary.go` (paths truncated via `truncateWithEllipsis`)
- [ ] 6.4 Write tests + refactor `viewRepoSummary()` variants in `repo_summary.go`
- [ ] 6.5 Write tests + refactor `viewMigrateSummary()` and `viewMigrateNext()` in `migrate_summary.go`
- [ ] 6.6 Write tests + refactor `viewCleanupResult()` in `cleanup_result.go`; while running, the title SHALL be `Running Cleanup`, switching to `Cleanup Complete` only once the result exists
- [ ] 6.7 Re-skin the integration apply summary (from `flow-state-correctness`) with the shared chrome

## 7. Migrate Confirmation Views

- [ ] 7.1 Write tests for `ConfirmationViewModel.View()` with standard chrome (no border, `·` hint separators)
- [ ] 7.2 Refactor `ConfirmationViewModel.View()` in `confirmation.go` — remove `styleDialogBox.Render()` wrapping and the `•` separators
- [ ] 7.3 Refactor `viewConfirm()` in `confirm.go` to standard chrome; when a worktree has no branch (detached HEAD), show the directory name or short SHA, never blank — with test
- [ ] 7.4 Update `cleanup_confirm_test.go`, `create_confirm_test.go`, `clone_confirm_test.go`, `migrate_confirm_test.go` to match new rendering

## 8. List View Framing

- [ ] 8.1 Give the remove list the standard framing (repo subtitle + separators) consistent with menu/integrations; update `list_test.go`

## 9. Cleanup Dead Code

- [ ] 9.1 Delete `styleHeader` and `styleDialogBox` from `styles.go`
- [ ] 9.2 Delete old `separator()` function (replaced by `viewSeparator`)
- [ ] 9.3 Verify no remaining references — `go vet ./...` and `go build`

## 10. Design System Documentation

- [ ] 10.1 Add "Component Patterns" section to `.impeccable.md`: view chrome, progress views, indicator vocabulary table (`●` done / `◐` active / `·` pending / `✗` failed — `●` never means pending), stat line, bar styling, truncation rule, key mapping
- [ ] 10.2 Review all views against the documented patterns

## 11. E2E Tests

- [ ] 11.1 Update `cleanup_e2e_test.go` to verify new chrome in cleanup flow
- [ ] 11.2 E2E: removal flow — select → progress with title + bar → summary with chrome
- [ ] 11.3 E2E: ctrl+c/q quits from all progress views
- [ ] 11.4 E2E: WindowSizeMsg updates adaptive windowing in removal progress

## 12. Verification

- [ ] 12.1 `go fmt ./...`, `go vet ./...`, `go test ./...` all green; binary builds
- [ ] 12.2 Manual visual check with `--playground` against `.impeccable.md` patterns (menu, list, removal, create, integrations, cleanup; also at 80x18)
- [ ] 12.3 Run commitlint + full test suite before push
