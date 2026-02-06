## Context

sentei is a greenfield Go project. This is the first change — there's no existing code. We're building the foundation layer that parses `git worktree list --porcelain` output into structured Go data. All downstream features (metadata enrichment, TUI, deletion) will consume the types and functions defined here.

Git's porcelain format is line-based and block-separated by blank lines. Each block represents one worktree with key-value pairs (some keys are bare flags like `bare`, `locked`, `prunable`).

## Goals / Non-Goals

**Goals:**

- Parse all fields from `git worktree list --porcelain` into a well-typed struct
- Handle every documented worktree state: bare, normal, detached HEAD, locked (with optional reason), prunable (with optional reason)
- Provide a clean public API: `ListWorktrees(repoPath string) ([]Worktree, error)`
- Validate that the target path is a git repository before running commands
- Make git command execution testable by abstracting behind an interface

**Non-Goals:**

- Metadata enrichment (commit dates, dirty status) — separate change
- TUI rendering or user interaction
- Worktree creation or deletion operations
- Configuration or CLI flag parsing

## Decisions

### D1: Separate parser from command execution

Parse from `string` input, not directly from `exec.Cmd`. The parser takes raw porcelain text and returns `[]Worktree`. A separate `CommandRunner` interface handles executing git and returning stdout.

**Why over coupling them**: Testability. We can test parsing with fixture strings without needing a real git repo. The runner can be mocked for integration tests.

### D2: `CommandRunner` interface for git execution

```go
type CommandRunner interface {
    Run(dir string, args ...string) (string, error)
}
```

A `GitRunner` struct implements this using `os/exec`. Tests can supply a mock.

**Why over direct `exec.Command`**: Allows unit testing without git. Also positions us for the enrichment layer which will need the same runner for multiple commands.

### D3: Worktree struct with value types

```go
type Worktree struct {
    Path       string
    HEAD       string
    Branch     string     // empty if detached
    IsBare     bool
    IsLocked   bool
    LockReason string     // empty if not locked or no reason
    IsPrunable bool
    PruneReason string    // empty if not prunable or no reason
    IsDetached bool
}
```

**Why no pointer fields**: Worktrees are small, immutable after parsing, and passed around by value or in slices. Pointers add nil-check complexity for no benefit.

**Why `Branch` as string instead of `*string`**: Empty string clearly means "no branch" (detached HEAD). The `IsDetached` bool makes intent explicit.

### D4: Repository validation via `git rev-parse`

Before listing worktrees, run `git -C <path> rev-parse --git-dir` to confirm the path is a git repository. This gives a clear error early rather than a confusing failure from `git worktree list`.

**Why over checking for `.git` directory**: `rev-parse` handles bare repos, linked worktrees, and non-standard layouts correctly. File existence checks are fragile.

### D5: File organization

- `internal/git/worktree.go` — `Worktree` struct definition
- `internal/git/parser.go` — `ParsePorcelain(input string) ([]Worktree, error)`
- `internal/git/commands.go` — `CommandRunner` interface, `GitRunner` struct, `ListWorktrees` function
- `internal/git/git_test.go` — table-driven tests for parsing + integration tests

Single package (`git`) keeps the API surface small. No sub-packages needed at this scale.

## Risks / Trade-offs

**[Porcelain format changes across git versions]** → Pin to documented behavior from git 2.20+. The porcelain format has been stable since worktree support matured. Add git version output to error context if parsing fails unexpectedly.

**[Lock/prune reason text may contain newlines]** → Git's porcelain format uses a single-line `locked <reason>` format. Multi-line reasons are not supported by git itself, so we don't need to handle them. Parse as everything after the first space on the line.

**[Empty worktree list in bare repo with no worktrees]** → This is valid — a bare repo has itself as the only "worktree" entry (with `bare` flag). Return the list as-is; the TUI layer decides whether to filter bare entries.
