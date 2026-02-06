## Context

sentei needs a reproducible test environment — a bare repo with worktrees in various states. Currently the only way to test is against real repos. The playground should be a first-class dev tool that's fast, idempotent, and requires zero manual setup.

## Goals / Non-Goals

**Goals:**
- One-command playground: `sentei --playground`
- Covers all worktree states the TUI handles (clean, dirty, untracked, locked, old, detached, prunable)
- Fast setup (<2 seconds)
- Clean Go code in `internal/playground/` — no shell scripts
- Automatic cleanup on exit, optional keep for debugging

**Non-Goals:**
- Configurable number of worktrees or states (hardcoded fixtures are fine for v1)
- Running the playground in CI (it's a dev-only tool)
- Simulating large repos (50+ worktrees) — that's a separate benchmark concern

## Decisions

### D1: Fixed playground directory at `/tmp/sentei-playground/`

Use a fixed, well-known path rather than `mktemp -d`.

**Why**: Idempotent. Running twice cleans and recreates. Users can `cd` into it to inspect. No accumulation of orphaned temp dirs.

### D2: Playground as Go package, not shell script

All setup logic lives in `internal/playground/setup.go` using `os/exec` to run git commands.

**Why**: Testable, portable (works on macOS/Linux without bash version concerns), consistent with the rest of the codebase. Functions are composable — `Setup() (string, func(), error)` returns the repo path and a cleanup function.

### D3: Reuse `git.CommandRunner` for git operations

The playground uses a local `exec.Command` runner directly rather than the `git.CommandRunner` interface. The CommandRunner is designed for `-C <path>` operations on existing repos. The playground needs `git init --bare`, `git worktree add`, etc., which operate differently.

**Why**: The playground is a setup utility, not part of the core data flow. A thin `run(dir, args...)` helper keeps it simple without coupling to the git package's abstraction.

### D4: Fixture worktrees cover all TUI states

Create exactly these worktrees:
1. `feature/active` — clean, recent commit (exercises `[ok]` indicator)
2. `feature/wip` — uncommitted changes (exercises `[~]` and dirty warning)
3. `experiment/abandoned` — untracked files only (exercises `[!]`)
4. `hotfix/locked` — locked worktree (exercises `[L]` and lock warning)
5. `chore/old-deps` — commit dated 90+ days ago (exercises relative time "3 months ago")
6. `detached-head` — detached HEAD state (exercises detached display)

This covers all status indicators, the confirmation dialog warnings, and relative time display.

### D5: Lifecycle: setup, launch TUI, cleanup

```
main.go --playground:
  1. playground.Setup()  →  returns repoPath, cleanupFn
  2. defer cleanupFn()   (unless --playground-keep)
  3. Run normal TUI flow against repoPath
```

The TUI launch code is the same as normal mode — just the repoPath changes. No special TUI behavior for playground mode.

### D6: File organization

```
internal/playground/
├── setup.go       # Setup() function, creates bare repo + worktrees
└── setup_test.go  # Verifies playground creates expected worktree states
```

## Risks / Trade-offs

- **[Git version differences]** → Some git commands behave differently across versions. Mitigation: use only basic porcelain commands that have been stable since git 2.15+.
- **[Stale playground]** → If user kills the process before cleanup, the dir persists. Mitigation: `Setup()` removes any existing playground dir before creating a new one (idempotent).
