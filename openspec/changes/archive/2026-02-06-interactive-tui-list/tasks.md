## 1. Dependencies and Project Setup

- [x] 1.1 Add Bubble Tea, Bubbles, and Lip Gloss dependencies to go.mod (`go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles github.com/charmbracelet/lipgloss`)
- [x] 1.2 Create `internal/tui/` package directory structure: model.go, list.go, confirm.go, progress.go, summary.go, styles.go, keys.go

## 2. Worktree Deletion Logic

- [x] 2.1 Define `DeletionEvent` (type: started/completed/failed, path, error) and `DeletionResult` (success count, failure count, per-worktree outcomes) types in `internal/worktree/deleter.go`
- [x] 2.2 Implement `DeleteWorktrees(runner, repoPath, worktrees, maxConcurrency, progress)` with bounded concurrency, sending events via channel, running `git worktree remove --force <path>` for each worktree
- [x] 2.3 Write tests for `DeleteWorktrees`: successful deletion, failed deletion, mixed results, empty input, concurrency bounds

## 3. TUI Foundation

- [x] 3.1 Define view state enum (`listView`, `confirmView`, `progressView`, `summaryView`) and `Model` struct in `model.go` — holds worktrees slice, selected set, cursor position, current view state, deletion results
- [x] 3.2 Implement `Init()` returning nil command (data loaded before TUI starts)
- [x] 3.3 Define Lip Gloss styles in `styles.go`: header, selected row, cursor row, status indicators, status bar, dialog box
- [x] 3.4 Define key bindings in `keys.go`: j/k, arrows, space, a, enter, q, y/n, escape, ctrl+c

## 4. List View

- [x] 4.1 Implement `relativeTime(t time.Time) string` helper returning human-readable relative time ("3 days ago", "2 months ago", "unknown" for zero time)
- [x] 4.2 Implement `stripBranchPrefix(ref string) string` to strip `refs/heads/` from branch names
- [x] 4.3 Implement `statusIndicator(wt git.Worktree) string` returning ASCII indicators `[L]`/`[~]`/`[!]`/`[ok]` based on locked > dirty > untracked > clean priority
- [x] 4.4 Implement list row rendering in `list.go`: checkbox + status + branch + relative time + commit subject, with cursor highlighting
- [x] 4.5 Implement list `Update` logic: j/k/arrow navigation, spacebar toggle, 'a' select/deselect all, enter to confirm (if selections > 0), q/ctrl+c to quit
- [x] 4.6 Implement viewport scrolling when cursor moves beyond visible area
- [x] 4.7 Implement status bar rendering: "N selected | space: toggle | a: all | enter: delete | q: quit"
- [x] 4.8 Write tests for relativeTime, stripBranchPrefix, statusIndicator helpers

## 5. Confirmation View

- [x] 5.1 Implement confirmation dialog rendering in `confirm.go`: header, list of selected worktrees with clean/dirty/locked status, warning counts, y/n prompt
- [x] 5.2 Implement confirmation `Update` logic: y confirms → transition to progress, n/escape → back to list with selections preserved
- [x] 5.3 Write tests for confirmation rendering with dirty/locked worktree warnings

## 6. Progress View

- [x] 6.1 Define Bubble Tea message types: `worktreeDeleteStartedMsg`, `worktreeDeletedMsg`, `worktreeDeleteFailedMsg`, `allDeletionsCompleteMsg`
- [x] 6.2 Implement `startDeletion` tea.Cmd that calls `DeleteWorktrees` and converts channel events into Bubble Tea messages
- [x] 6.3 Implement progress view rendering in `progress.go`: progress bar (bubbles/progress), per-worktree status list (pending/removing/removed/failed)
- [x] 6.4 Implement progress `Update` logic: handle deletion messages, update progress bar, transition to summary on completion

## 7. Summary View

- [x] 7.1 Implement summary view rendering in `summary.go`: success count, failure count, failure details, "git worktree prune" suggestion
- [x] 7.2 Implement summary `Update` logic: q/enter/escape to quit

## 8. Main Entry Point

- [x] 8.1 Update `main.go`: discover worktrees, enrich them, filter out bare entry, initialize TUI model, run `tea.NewProgram(model).Run()`
- [x] 8.2 Verify full flow works end-to-end with a real bare repo setup
