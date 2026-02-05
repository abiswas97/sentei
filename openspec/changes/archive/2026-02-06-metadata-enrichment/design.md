## Context

The discovery layer (`internal/git/`) parses `git worktree list --porcelain` into `[]Worktree` structs containing structural data (path, branch, HEAD, lock/prune status). The TUI needs additional metadata per worktree — last commit date, last commit subject, and dirty/untracked status — to render status indicators and enable staleness sorting. This enrichment layer sits between discovery and TUI in the data flow pipeline (PRD §7.3).

Current state:
- `Worktree` struct in `internal/git/worktree.go` has 9 fields (all from porcelain parsing)
- `CommandRunner` interface in `internal/git/commands.go` abstracts git execution
- No enrichment logic exists yet

## Goals / Non-Goals

**Goals:**
- Enrich all worktrees with commit metadata and status in parallel
- Complete enrichment in <5s for 50 worktrees
- Handle edge cases gracefully (bare entries, prunable/missing, command failures)
- Keep enrichment testable via the existing `CommandRunner` interface

**Non-Goals:**
- TUI rendering of enriched data (separate feature)
- Sorting/filtering logic (separate feature)
- Caching enrichment results across runs
- Enrichment of remote branch status (merge status with main)

## Decisions

### D1: Enrichment fields live on the Worktree struct (not a separate type)

**Decision**: Extend `git.Worktree` with enrichment fields rather than creating a wrapper type.

**Rationale**: The TUI will pass worktrees around as a single unit. A wrapper type adds indirection with no benefit — the struct is internal, so adding fields is non-breaking. Zero-value fields (`time.Time{}`, `false`, `""`) naturally represent "not enriched."

**Alternative considered**: `EnrichedWorktree` wrapper struct. Rejected because it forces type conversions everywhere and the struct is already internal.

### D2: New package `internal/worktree/` for enrichment logic

**Decision**: Place `EnrichWorktrees` in `internal/worktree/enricher.go`, separate from `internal/git/`.

**Rationale**: The `internal/git/` package handles git command execution and output parsing — low-level concerns. Enrichment is business logic that orchestrates multiple git commands per worktree. Separating keeps `git/` focused and `worktree/` as the domain layer. Matches the project structure in PRD §7.2.

**Alternative considered**: Adding enrichment to `internal/git/`. Rejected because it mixes abstraction levels.

### D3: Bounded concurrency with semaphore pattern

**Decision**: Use a `sync.WaitGroup` + buffered channel semaphore to limit concurrent git commands (default: 10 concurrent).

**Rationale**: Unbounded goroutines for 50+ worktrees would spawn 150+ concurrent git processes, risking OS resource exhaustion. A semaphore channel (`make(chan struct{}, maxConcurrency)`) provides simple, stdlib-only bounded parallelism.

**Alternative considered**: Worker pool with job channel. Rejected as over-engineered for this use case — the semaphore pattern is idiomatic Go and sufficient.

### D4: Per-worktree error collection, not fail-fast

**Decision**: Collect enrichment errors per-worktree in an `EnrichmentError` field on the struct. Return all worktrees regardless of individual failures.

**Rationale**: A single missing/broken worktree should not prevent the user from seeing and managing the rest. The TUI can display a warning indicator for worktrees that failed enrichment.

**Alternative considered**: Return `([]Worktree, []error)` tuple. Rejected because associating errors with specific worktrees is clearer when the error is on the struct itself.

### D5: Skip enrichment for bare and prunable entries

**Decision**: Check `IsBare` and `IsPrunable` before spawning enrichment goroutines. Set a flag indicating enrichment was skipped (not failed).

**Rationale**: Bare entries have no working tree to inspect. Prunable entries have missing directories — git commands would fail. Skipping avoids unnecessary work and error noise.

## Risks / Trade-offs

- **[Risk] Git commands slow on network filesystems** → Mitigation: Configurable concurrency limit; all commands use `-C <path>` for local execution only.
- **[Risk] Worktree directory deleted between discovery and enrichment** → Mitigation: Treat command failure same as prunable — mark as enrichment-unavailable, don't crash.
- **[Trade-off] Enrichment fields on Worktree struct increase coupling** → Acceptable: struct is internal, and separating would add complexity without benefit.
- **[Trade-off] No enrichment caching** → Acceptable for MVP: enrichment runs once per TUI session, <5s is fast enough.
