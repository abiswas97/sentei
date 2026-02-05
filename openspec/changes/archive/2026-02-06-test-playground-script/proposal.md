## Why

There's no way to test the TUI without a real bare repo with worktrees in various states. Developers need a one-command way to spin up a realistic test environment and interact with wt-sweep against it. This also serves as a demo/onboarding tool for new users.

## What Changes

- Add a `--playground` flag to `wt-sweep` that creates a temporary bare repo in `/tmp/wt-sweep-playground/` with worktrees in distinct states (clean, dirty, untracked, locked, old commit, detached HEAD)
- Launch the TUI against the playground repo automatically after setup
- Playground is idempotent: re-running tears down and recreates
- Cleanup happens automatically on exit (deferred removal of `/tmp` dir), with a `--playground-keep` flag to preserve it for inspection
- All playground logic lives in `internal/playground/` as composable Go functions — no shell scripts

## Capabilities

### New Capabilities
- `playground-setup`: Programmatic creation of a bare repo with worktrees in various states for testing and demo purposes

### Modified Capabilities
_None — existing specs are unchanged._

## Impact

- **New package**: `internal/playground/` (setup, teardown, fixture creation)
- **Modified files**: `main.go` (add `--playground` and `--playground-keep` flag handling)
- **No new dependencies**: uses only `os/exec` for git commands (same pattern as existing `git.CommandRunner`)
- **Filesystem**: creates/removes `/tmp/wt-sweep-playground/` directory
