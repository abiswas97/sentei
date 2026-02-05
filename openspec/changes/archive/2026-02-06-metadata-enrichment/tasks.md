## 1. Extend Worktree Struct

- [x] 1.1 Add enrichment fields to `Worktree` struct in `internal/git/worktree.go`: `LastCommitDate` (time.Time), `LastCommitSubject` (string), `HasUncommittedChanges` (bool), `HasUntrackedFiles` (bool), `IsEnriched` (bool), `EnrichmentError` (string)
- [x] 1.2 Add `"time"` import to `internal/git/worktree.go`
- [x] 1.3 Verify existing parser tests still pass with extended struct (zero-value enrichment fields)

## 2. Status Parsing Logic

- [x] 2.1 Create `internal/worktree/` package directory
- [x] 2.2 Implement `ParseStatusPorcelain(output string) (hasUncommitted bool, hasUntracked bool)` in `internal/worktree/enricher.go` — parse `git status --porcelain` output to detect uncommitted changes (M/A/D/R/C/U prefixes) and untracked files (`??` prefix)
- [x] 2.3 Write table-driven tests for `ParseStatusPorcelain`: clean output, uncommitted only, untracked only, both, empty string

## 3. Commit Date Parsing Logic

- [x] 3.1 Implement `ParseCommitDate(output string) (time.Time, error)` in `internal/worktree/enricher.go` — parse `git log -1 --format=%ai` output into `time.Time`
- [x] 3.2 Write tests for `ParseCommitDate`: valid date string, empty output (orphan branch), malformed date

## 4. Single Worktree Enrichment

- [x] 4.1 Implement `enrichWorktree(runner CommandRunner, wt *Worktree)` that runs all three git commands (`log -1 --format=%ai`, `log -1 --format=%s`, `status --porcelain`) and populates enrichment fields
- [x] 4.2 Handle command failures: set `EnrichmentError` and leave fields at zero values
- [x] 4.3 Write tests using a mock `CommandRunner` for: successful enrichment, failed log command, failed status command

## 5. Parallel Enrichment Orchestrator

- [x] 5.1 Implement `EnrichWorktrees(runner CommandRunner, worktrees []Worktree, maxConcurrency int) []Worktree` with bounded concurrency via semaphore channel + sync.WaitGroup
- [x] 5.2 Skip enrichment for worktrees with `IsBare=true` or `IsPrunable=true` (leave IsEnriched=false)
- [x] 5.3 Set `IsEnriched=true` on successfully enriched worktrees
- [x] 5.4 Write tests: mixed slice with bare/prunable/normal entries, verify correct skip/enrich behavior
- [x] 5.5 Write test: simulate one worktree failure, verify other worktrees still enriched and error captured

## 6. Integration Verification

- [x] 6.1 Run `go vet ./...` and `go fmt ./...` to verify no issues
- [x] 6.2 Run full test suite `go test ./...` to verify no regressions
- [x] 6.3 Verify `go build` succeeds
