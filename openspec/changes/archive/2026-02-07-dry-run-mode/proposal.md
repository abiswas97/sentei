## Why

Users need a way to preview which worktrees would be deleted without entering the interactive TUI. This supports scripting, quick audits, and cautious workflows where you want to see the plan before committing to an interactive session. It also needs to compose with `--playground` for testing.

## What Changes

- Add `--dry-run` flag to CLI
- When active, skip the TUI entirely — print a non-interactive summary of all worktrees to stdout with their status indicators and metadata
- Exit immediately after printing (no confirmation, no deletion)
- Compatible with `--playground` (create temp repo, print summary, clean up)

## Capabilities

### New Capabilities
- `dry-run`: Non-interactive preview of worktrees with status, branch, age, and commit info printed to stdout

### Modified Capabilities

## Impact

- `main.go`: New flag parsing and conditional branch before TUI launch
- New or extended rendering logic for plain-text (non-TUI) worktree output
- No changes to existing TUI, deletion, or enrichment logic
