## 1. Dry-Run Scan API

- [ ] 1.1 Write failing tests for `cleanup.DryRun` in `internal/cleanup/dryrun_test.go` — clean repo (all zeros), repo with stale refs + config duplicates, repo with non-worktree branches (aggressive targets), BranchInfo contains name + merge date + last commit
- [ ] 1.2 Extract shared scan helpers from `cleanup.Run` into reusable functions
- [ ] 1.3 Implement `cleanup.DryRun(runner, repoPath) DryRunResult` and `BranchInfo` type to pass tests
- [ ] 1.4 Write E2E test for `DryRun` against a real bare repo with stale branches

## 2. Cleanup Preview View

- [ ] 2.1 Write failing tests for `viewCleanupPreview` in `internal/tui/cleanup_preview_test.go` — loading state, safe results with actions, safe results all clean, aggressive offer with inline preview (2-3 names + "and N more"), no aggressive section when nothing to clean
- [ ] 2.2 Create `internal/tui/cleanup_preview.go` with `cleanupPreviewView` state, `updateCleanupPreview`, and `viewCleanupPreview` to pass tests
- [ ] 2.3 Write tests for key hints: with aggressive (`enter safe · a aggressive · ? details · esc back`), without aggressive (`enter safe · esc back`)
- [ ] 2.4 Implement key hint logic in `viewCleanupPreview`
- [ ] 2.5 Wire dry-run scan: menu → `cleanupPreviewView` → fire `DryRun` Cmd with `bufferTransition` → populate preview on result

## 3. Detail Portal Integration for Cleanup

- [ ] 3.1 Write tests for cleanup detail content builder — generates formatted branch list with merge date and last commit subject from `DryRunResult.NonWtBranches`
- [ ] 3.2 Implement cleanup detail content builder
- [ ] 3.3 Wire `?` key in `updateCleanupPreview` to open portal with cleanup detail content (no-op when no aggressive items)

## 4. Execute Cleanup from Preview

- [ ] 4.1 Write tests for `enter` key executing safe cleanup from preview
- [ ] 4.2 Write tests for `a` key showing aggressive confirmation ("Delete N branches? y/n"), then executing on `y`
- [ ] 4.3 Implement execution paths: `enter` → safe cleanup → result view, `a` → confirm → aggressive cleanup → result view

## 5. Removal Safety Gate

- [ ] 5.1 Write failing tests for pushed-to-remote detection in `internal/worktree/` — branch up to date, branch ahead of remote, no tracking branch
- [ ] 5.2 Implement pushed-to-remote check using `git rev-list @{upstream}..HEAD`
- [ ] 5.3 Add pushed status to worktree enrichment (extend `EnrichWorktrees`)
- [ ] 5.4 Write failing tests for removal confirmation gate in `internal/tui/confirm_test.go` — all clean/pushed skips gate, dirty triggers gate, unpushed triggers gate, untracked triggers gate
- [ ] 5.5 Implement confirmation gate logic: check selected worktrees on Enter, show gate or proceed
- [ ] 5.6 Write tests for gate view rendering — mixed risk display, summary warning, key hints
- [ ] 5.7 Implement gate view using chrome helpers

## 6. --yes CLI Flag

- [ ] 6.1 Add `--yes` / `-y` flag to `cmd/cleanup.go` and `cmd/remove.go`
- [ ] 6.2 Write tests: `--yes` skips confirmation, `--yes` does NOT skip dirty gate, `--yes` with aggressive proceeds
- [ ] 6.3 Wire `--yes` flag to skip confirmation in CLI paths
- [ ] 6.4 Update `sentei --help` and subcommand help text

## 7. CLI Command Echo Migration

- [ ] 7.1 Add CLI command echo to removal summary view
- [ ] 7.2 Add CLI command echo to cleanup result view
- [ ] 7.3 Update tests for summary views to verify CLI command presence
- [ ] 7.4 Remove CLI command echo from TUI confirmation path (keep in CLI confirmation path)

## 8. Menu Flow Update

- [ ] 8.1 Wire menu "Cleanup & exit" to `cleanupPreviewView` instead of `cleanupConfirmView`
- [ ] 8.2 Keep `cleanupConfirmView` for CLI path (`sentei cleanup --mode=X`)
- [ ] 8.3 Update menu E2E tests for new flow

## 9. Design System Documentation

- [ ] 9.1 Add cleanup flow patterns to `.impeccable.md` — preview-first pattern, inline preview, upgrade offer
- [ ] 9.2 Add safety gate patterns to `.impeccable.md` — when to gate, at-risk indicators

## 10. E2E Tests

- [ ] 10.1 Write E2E test: menu → cleanup preview → scan loading → safe results → enter → cleanup runs → result
- [ ] 10.2 Write E2E test: cleanup preview with aggressive → press `a` → confirm → aggressive runs
- [ ] 10.3 Write E2E test: cleanup preview → `?` → detail portal opens → esc → portal closes
- [ ] 10.4 Write E2E test: removal with dirty worktree → gate shown → confirm → proceeds
- [ ] 10.5 Write E2E test: removal with all clean → no gate → proceeds directly
- [ ] 10.6 Write E2E test: `sentei cleanup --mode safe --yes` → no confirmation → runs immediately

## 11. Verification

- [ ] 11.1 Run `go fmt ./...` and `go vet ./...`
- [ ] 11.2 Run `go test ./...` — all tests pass
- [ ] 11.3 Run `go build` — binary builds
- [ ] 11.4 Manual test: full cleanup flow from menu with both safe and aggressive paths
- [ ] 11.5 Manual test: remove dirty worktree and verify gate appears
- [ ] 11.6 Manual test: `sentei cleanup --yes` from CLI
- [ ] 11.7 Update session meta doc with completion status
