## 1. Project Setup

- [x] 1.1 Initialize Go module (`go mod init`) and create `internal/git/` package directory
- [x] 1.2 Create `main.go` with minimal entry point (placeholder that calls `ListWorktrees` and prints results)

## 2. Worktree Data Model

- [x] 2.1 Define `Worktree` struct in `internal/git/worktree.go` with all fields (Path, HEAD, Branch, IsBare, IsLocked, LockReason, IsPrunable, PruneReason, IsDetached)

## 3. Porcelain Parser

- [x] 3.1 Implement `ParsePorcelain(input string) ([]Worktree, error)` in `internal/git/parser.go` — split input into blocks by blank lines, parse each block line-by-line
- [x] 3.2 Handle `worktree <path>`, `HEAD <sha>`, `branch <ref>`, `bare`, `detached`, `locked`, `locked <reason>`, `prunable`, `prunable <reason>` lines
- [x] 3.3 Write table-driven tests for parser covering: empty input, single bare entry, normal worktree, detached HEAD, locked with/without reason, prunable with/without reason, multiple worktrees, branch ref preservation

## 4. Git Command Execution

- [x] 4.1 Define `CommandRunner` interface and `GitRunner` struct in `internal/git/commands.go`
- [x] 4.2 Implement `GitRunner.Run()` using `os/exec` — return trimmed stdout on success, stderr-wrapped error on failure
- [x] 4.3 Implement repository validation function using `git -C <path> rev-parse --git-dir`
- [x] 4.4 Implement `ListWorktrees(runner CommandRunner, repoPath string) ([]Worktree, error)` — validate repo, run `git worktree list --porcelain`, parse output

## 5. Tests for Command Layer

- [x] 5.1 Write unit tests with a mock `CommandRunner` for: successful listing, repo validation failure, git command failure
- [x] 5.2 Write tests for edge cases: path does not exist, empty worktree list output

## 6. Verification

- [x] 6.1 Run `go vet ./...` and `go fmt ./...` with no issues
- [x] 6.2 Run `go build` to confirm the binary compiles
- [x] 6.3 All tests pass with `go test ./...`
