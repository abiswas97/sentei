## 1. Playground Package

- [x] 1.1 Create `internal/playground/setup.go` with `Setup() (repoPath string, cleanup func(), err error)` that removes any existing playground dir, creates `/tmp/sentei-playground/`, and initializes a bare repo at `repo.git`
- [x] 1.2 Add helper `gitRun(dir string, args ...string) error` for running git commands during setup (thin wrapper around `os/exec`)
- [x] 1.3 Create an initial commit on `main` branch with a seed file (needed before worktrees can be added)
- [x] 1.4 Add `feature/active` worktree — clean, recent commit
- [x] 1.5 Add `feature/wip` worktree — create a file, commit it, then modify it (uncommitted changes)
- [x] 1.6 Add `experiment/abandoned` worktree — add an untracked file (no staged/committed changes)
- [x] 1.7 Add `hotfix/locked` worktree — create worktree, then `git worktree lock`
- [x] 1.8 Add `chore/old-deps` worktree — commit with `GIT_AUTHOR_DATE` and `GIT_COMMITTER_DATE` set 90+ days in the past
- [x] 1.9 Add detached HEAD worktree — `git worktree add --detach`

## 2. Tests

- [x] 2.1 Write `setup_test.go` that calls `Setup()`, then runs `git worktree list --porcelain` on the result and verifies: correct number of worktrees, expected branch names, locked state, detached state
- [x] 2.2 Test idempotency: call `Setup()` twice, verify second call succeeds without error
- [x] 2.3 Test cleanup function removes the playground directory

## 3. CLI Integration

- [x] 3.1 Add `--playground` and `--playground-keep` flag parsing in `main.go`
- [x] 3.2 When `--playground` is set: call `Setup()`, defer cleanup (unless `--playground-keep`), set repoPath to the returned path, then continue normal TUI flow
- [x] 3.3 Print playground path on startup so user knows where the repo lives
