## 1. Dry-run package

- [x] 1.1 Create `internal/dryrun/dryrun.go` with `Print(worktrees []git.Worktree, w io.Writer)` that outputs a plain-text table (Status, Branch, Age, Subject) sorted by age ascending — no ANSI colors
- [x] 1.2 Create `internal/dryrun/dryrun_test.go` with table-driven tests covering: normal worktrees, dirty/untracked/locked indicators, empty input, enrichment errors, sort order

## 2. CLI integration

- [x] 2.1 Add `--dry-run` flag to `main.go` flag parsing
- [x] 2.2 After enrichment and filtering, branch on dry-run: call `dryrun.Print` to stdout and exit instead of launching TUI
- [x] 2.3 Verify composability: `--playground --dry-run` and `--dry-run /path/to/repo` both work

## 3. Verification

- [x] 3.1 Run `go vet ./...` and `go test ./...`
- [x] 3.2 Manual test: `go run . --playground --dry-run` prints table and exits
- [x] 3.3 Manual test: pipe output through grep to confirm no ANSI codes
