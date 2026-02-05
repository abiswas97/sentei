## Why

wt-sweep needs a foundation layer that can discover all worktrees in a git repository and parse their metadata from `git worktree list --porcelain` output. This is the first building block — every other feature (TUI display, enrichment, deletion) depends on having a reliable, structured representation of worktrees parsed from git's porcelain output.

## What Changes

- Add a `Worktree` struct representing a single git worktree with fields for path, HEAD commit, branch, bare status, locked status, and prunable status
- Implement parsing of `git worktree list --porcelain` output into `[]Worktree`
- Add a git command execution layer that runs `git worktree list --porcelain` and returns raw output
- Handle edge cases: bare repository entries, detached HEAD, locked worktrees, prunable worktrees, worktrees with lock reasons
- Validate that the tool is running inside a git repository (bare or regular)

## Capabilities

### New Capabilities

- `worktree-discovery`: Discovering worktrees via `git worktree list --porcelain` and parsing the porcelain output into structured `Worktree` data. Covers the data model, parsing logic, git command execution, and repository validation.

### Modified Capabilities

(none — this is the first change)

## Impact

- **New code**: `internal/git/` package — `worktree.go` (struct), `parser.go` (parsing), `commands.go` (git execution)
- **Dependencies**: No external dependencies beyond Go stdlib (`os/exec`, `strings`, `bufio`)
- **APIs**: Exposes `Worktree` struct and `ListWorktrees(repoPath string) ([]Worktree, error)` as the public interface consumed by enrichment and TUI layers
