## 1. Add generation token and update message type

- [ ] 1.1 Unit test: `loadWorktreeContext` returns a `worktreeContextMsg` with the generation value it was given.
- [ ] 1.2 Add `worktreeGeneration uint64` field to Model in `model.go`. Add `generation uint64` field to `worktreeContextMsg` in `menu.go`. Update `loadWorktreeContext` signature to accept a generation parameter and include it in the returned message.

## 2. Global `worktreeContextMsg` handler

- [ ] 2.1 Unit tests for global handler: applies worktrees when generation matches and no error; discards when generation mismatches; discards when error is non-nil; works when current view is summaryView, progressView, or menuView (table-driven across views).
- [ ] 2.2 Add global `worktreeContextMsg` handler in `Update()` (model.go), before the `switch m.view` dispatch. Apply worktrees, call `reindex()` and `updateMenuHints()` only when `msg.generation == m.worktreeGeneration` and `msg.err == nil`. Discard otherwise.
- [ ] 2.3 Remove `case worktreeContextMsg` from `updateMenu` in `menu.go` (dead code after global handler). Verify existing menu tests still compile and pass.

## 3. Wire Init() to use generation token

- [ ] 3.1 Unit test: `Init()` returns a command and the model's `worktreeGeneration` is 1 after calling it.
- [ ] 3.2 Update `Init()` in `model.go` to increment `worktreeGeneration` and pass it to `loadWorktreeContext`.

## 4. Eager reload at mutation completion sites

- [ ] 4.1 Unit test for `progress.go`: when `cleanupCompleteMsg` is received, the returned command is a `tea.Batch` containing `loadWorktreeContext`, and `worktreeGeneration` was incremented.
- [ ] 4.2 `progress.go`: In the `cleanupCompleteMsg` handler, replace `m.stateStale = true` with `m.worktreeGeneration++` and return `tea.Batch(holdCmd, loadWorktreeContext(..., m.worktreeGeneration))` using the `updated, holdCmd := m.holdOrAdvance(...)` pattern.
- [ ] 4.3 Unit test for `create_progress.go`: when `createCompleteMsg` is received, the returned command is a `tea.Batch` containing `loadWorktreeContext`, and `worktreeGeneration` was incremented.
- [ ] 4.4 `create_progress.go`: Same pattern in the `createCompleteMsg` handler.
- [ ] 4.5 Unit test for `integration_progress.go`: when `integrationFinalizedMsg` is received with `returnView != migrateNextView`, the returned command is a `tea.Batch` containing `loadWorktreeContext`. When `returnView == migrateNextView`, verify NO `loadWorktreeContext` is fired.
- [ ] 4.6 `integration_progress.go`: Same pattern in the `integrationFinalizedMsg` handler, conditional on `returnView != migrateNextView`. When `returnView == migrateNextView`, call `holdOrAdvance` without batching a reload.

## 5. Remove stateStale from non-menu flows

- [ ] 5.1 Unit test for `repo_progress.go`: when `repoDoneMsg` is received, verify `stateStale` is not set and no `loadWorktreeContext` is fired.
- [ ] 5.2 `repo_progress.go`: Remove `m.stateStale = true` from `repoDoneMsg` handler.
- [ ] 5.3 Unit test for `cleanup_result.go`: when `standaloneCleanupDoneMsg` is received, verify `stateStale` is not set and no `loadWorktreeContext` is fired.
- [ ] 5.4 `cleanup_result.go`: Remove `m.stateStale = true` from `standaloneCleanupDoneMsg` handler.

## 6. Delete stateStale mechanism

- [ ] 6.1 Remove `stateStale` field from Model in `model.go`. Remove the `stateStale` gate from `updateMenu` in `menu.go`.
- [ ] 6.2 Fix any test compilation errors referencing `stateStale`. Update tests from 5.1/5.3 to assert absence of reload command rather than checking the deleted field.

## 7. E2E validation

- [ ] 7.1 E2E test (teatest): after worktree removal, return to menu and verify the menu count is updated and a j/k keypress is processed immediately as cursor movement (not swallowed by a reload gate).
- [ ] 7.2 E2E test (teatest): after worktree creation, return to menu and verify the menu count reflects the new worktree.
