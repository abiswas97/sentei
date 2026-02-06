## 1. Core Logic

- [x] 1.1 Add `IsProtectedBranch(branch string) bool` function in `internal/git/` with protected list: main, master, develop, dev
- [x] 1.2 Add tests for `IsProtectedBranch` covering all scenarios (exact match, substring no-match, case sensitivity, empty/detached)

## 2. TUI Selection Logic

- [x] 2.1 Update spacebar handler in `model.go` to skip protected worktrees
- [x] 2.2 Update select-all (`a`) handler to skip protected worktrees when toggling

## 3. TUI Rendering

- [x] 3.1 Update `list.go` row rendering to show `[P]` instead of checkbox for protected worktrees
- [x] 3.2 Update status bar legend to include `[P] protected` indicator

## 4. Dry Run

- [x] 4.1 Update `internal/dryrun/` output to show `[P]` status for protected worktrees

## 5. Playground

- [x] 5.1 Ensure playground setup includes a `main` worktree so protection is visible during testing
