## 1. Dry-Run Scan API

- [x] 1.1 Write failing tests for `cleanup.DryRun` in `internal/cleanup/dryrun_test.go` — clean repo (all zeros), repo with stale refs + config duplicates, repo with non-worktree branches (aggressive targets), BranchInfo contains name + merge date + last commit
- [x] 1.2 Extract shared scan helpers from `cleanup.Run` into reusable functions
- [x] 1.3 Implement `cleanup.DryRun(runner, repoPath) DryRunResult` and `BranchInfo` type to pass tests
- [x] 1.4 Write E2E test for `DryRun` against a real bare repo with stale branches

## 2. Cleanup Preview View

- [x] 2.1 Write failing tests for `viewCleanupPreview` in `internal/tui/cleanup_preview_test.go` — loading state, safe results with actions, safe results all clean, aggressive offer with inline preview (2-3 names + "and N more"), no aggressive section when nothing to clean
- [x] 2.2 Create `internal/tui/cleanup_preview.go` with `cleanupPreviewView` state, `updateCleanupPreview`, and `viewCleanupPreview` to pass tests
- [x] 2.3 Write tests for key hints: with aggressive (`enter safe · a aggressive · ? details · esc back`), without aggressive (`enter safe · esc back`)
- [x] 2.4 Implement key hint logic in `viewCleanupPreview`
- [x] 2.5 Wire dry-run scan: menu → `cleanupPreviewView` → fire `DryRun` Cmd; the existing `progressStartedAt`/`holdOrAdvance` hold keeps the scanning state visible (`bufferTransition` never existed)

## 3. Detail Portal Integration for Cleanup

- [x] 3.1 Write tests for cleanup detail content builder — generates formatted branch list with merge date and last commit subject from `DryRunResult.NonWtBranches`
- [x] 3.2 Implement cleanup detail content builder
- [x] 3.3 Add a `cleanupPreviewView` case to the Model's `detailContent()` provider (the global `?` handler opens the portal when a view supplies content; no per-view key wiring)

## 4. Execute Cleanup from Preview

- [x] 4.1 Write tests for `enter` key executing safe cleanup from preview
- [x] 4.2 Write tests for `a` key showing aggressive confirmation ("Delete N branches? y/n"), then executing on `y`
- [x] 4.3 Implement execution paths: `enter` → safe cleanup → result view, `a` → confirm → aggressive cleanup → result view

## 5. Removal Safety Gate

- [x] 5.1 Write failing tests for pushed-to-remote detection in `internal/worktree/` — branch up to date, branch ahead of remote, no tracking branch
- [x] 5.2 Implement pushed-to-remote check using `git rev-list @{upstream}..HEAD`
- [x] 5.3 Add pushed status to worktree enrichment (extend `EnrichWorktrees`)
- [x] 5.4 Write failing tests for removal confirmation gate in `internal/tui/confirm_test.go` — all clean/pushed skips gate, dirty triggers gate, unpushed triggers gate, untracked triggers gate
- [x] 5.5 Implement confirmation gate logic: check selected worktrees on Enter, show gate or proceed
- [x] 5.6 Write tests for gate view rendering — mixed risk display, summary warning, key hints
- [x] 5.7 Implement gate view using chrome helpers

## 6. --yes CLI Flag

- [x] 6.1 Add `--yes` / `-y` flag to `cmd/cleanup.go` and `cmd/remove.go`
- [x] 6.2 Write tests: `--yes` skips confirmation, `--yes` does NOT skip dirty gate, `--yes` with aggressive proceeds
- [x] 6.3 Wire `--yes` flag to skip confirmation in CLI paths
- [x] 6.4 Update `sentei --help` and subcommand help text

## 7. CLI Command Echo Migration

- [x] 7.1 Add CLI command echo to removal summary view
- [x] 7.2 Add CLI command echo to cleanup result view
- [x] 7.3 Update tests for summary views to verify CLI command presence
- [x] 7.4 Remove CLI command echo from TUI confirmation path (keep in CLI confirmation path)

## 8. Menu Flow Update

- [x] 8.1 Wire menu "Cleanup & exit" to `cleanupPreviewView` instead of `cleanupConfirmView`
- [x] 8.2 Keep `cleanupConfirmView` for CLI path (`sentei cleanup --mode=X`)
- [x] 8.3 Update menu E2E tests for new flow

## 9. Design System Documentation

- [x] 9.1 Add cleanup flow patterns to `.impeccable.md` — preview-first pattern, inline preview, upgrade offer
- [x] 9.2 Add safety gate patterns to `.impeccable.md` — when to gate, at-risk indicators

## 10. E2E Tests

- [x] 10.1 Write E2E test: menu → cleanup preview → scan loading → safe results → enter → cleanup runs → result
- [x] 10.2 Write E2E test: cleanup preview with aggressive → press `a` → confirm → aggressive runs
- [x] 10.3 Write E2E test: cleanup preview → `?` → detail portal opens → esc → portal closes
- [x] 10.4 Write E2E test: removal with dirty worktree → gate shown → confirm → proceeds
- [x] 10.5 Write E2E test: removal with all clean → no gate → proceeds directly
- [x] 10.6 Write E2E test: `sentei cleanup --mode safe --yes` → no confirmation → runs immediately

## 11. Verification

- [x] 11.1 Run `go fmt ./...` and `go vet ./...`
- [x] 11.2 Run `go test ./...` — all tests pass
- [x] 11.3 Run `go build` — binary builds
- [x] 11.4 Manual test: full cleanup flow from menu with both safe and aggressive paths
- [x] 11.5 Manual test: remove dirty worktree and verify gate appears
- [x] 11.6 Manual test: `sentei cleanup --yes` from CLI
- [x] 11.7 Session meta doc was removed earlier (stale handoff docs); status lives here

## 12. Post-review honesty fixes (from playground verification)

- [x] 12.1 DryRun classifies merged-ness (`BranchInfo.Merged`, one `branch --merged` call); preview marks `(not merged)` candidates and discloses the count
- [x] 12.2 Aggressive confirm prompt promises only deletable branches (`Delete N branches? (M unmerged will be skipped)`)
- [x] 12.3 Cleanup result surfaces `BranchesSkipped` with names — a confirmed aggressive run can no longer read as a silent success
- [x] 12.4 Post-aggressive tip points at `--force` instead of re-recommending the mode that just ran
- [x] 12.5 Gate voice normalized (quiet `⚠` warnings, no shouting); tips pluralized; `--mode aggressive` form unified; CLI prints a result line for the all-clean non-worktree check
