# Worktree File Propagation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add safe, automatic propagation of repository-private files into every newly created Sentei worktree.

**Architecture:** A new `internal/worktreefile` domain package defines and executes rules. Config parses and validates the top-level list, while the shared creator pipeline performs one Setup step used by both CLI and TUI.

**Tech Stack:** Go 1.21+, `gopkg.in/yaml.v3`, Sentei progress/creator contracts, real temporary Git repositories.

---

### Task 1: Configuration contract

**Files:**
- Create: `internal/worktreefile/rule.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `internal/config/config_e2e_test.go`

- [ ] Add YAML parsing assertions for `source`, `destination`, and default/explicit `overwrite`.
- [ ] Run `go test ./internal/config` and verify the new assertions fail because `worktree_files` is absent.
- [ ] Add merge tests proving a non-empty overlay replaces rules and an empty overlay preserves the lower layer.
- [ ] Add table-driven validation tests for blank, absolute, dot, and traversal paths.
- [ ] Implement `worktreefile.Rule`, add `Config.WorktreeFiles`, merge it with existing list conventions, and validate every rule.
- [ ] Run `go test ./internal/config` and verify the package passes.

### Task 2: Safe copy engine

**Files:**
- Create: `internal/worktreefile/copy.go`
- Create: `internal/worktreefile/copy_test.go`

- [ ] Add tests for nested parent creation, source permissions, default no-overwrite, and explicit overwrite.
- [ ] Run `go test ./internal/worktreefile` and verify failure because `Copy` is absent.
- [ ] Implement contained source resolution, destination component checks, temporary-file copy, permission application, and atomic rename.
- [ ] Run focused tests and verify green.
- [ ] Add failing tests for missing source, non-regular source, source symlink escape, destination ancestor symlink, destination symlink, and copy failure.
- [ ] Implement the smallest additional checks needed and rerun the focused package.

### Task 3: Creator orchestration

**Files:**
- Modify: `internal/creator/creator.go`
- Modify: `internal/creator/prepared.go`
- Modify: `internal/creator/creator_test.go`
- Modify: `internal/creator/contract_test.go`

- [ ] Add tests expecting a `Copy worktree files` Setup step after merge and before dependencies.
- [ ] Add tests that a failed worktree-file copy is visible while independent dependency and integration steps continue.
- [ ] Run the focused creator tests and verify the expected missing-step failures.
- [ ] Add `WorktreeFiles` to `creator.Options`, plan the semantic step, and call the copy engine with `RepoPath` and the created worktree path.
- [ ] Run `go test ./internal/creator` and verify green.

### Task 4: CLI and TUI parity

**Files:**
- Modify: `cmd/create.go`
- Modify: `cmd/create_run_test.go`
- Modify: `internal/tui/create_options.go`
- Modify: `internal/tui/create_options_test.go`
- Modify: `internal/tui/create_confirm.go`
- Modify: `internal/tui/create_confirm_test.go`

- [ ] Add a non-interactive creation test whose container `.sentei.yaml` copies a nested private template without ecosystem flags.
- [ ] Add TUI option/confirmation tests showing configured automatic file count and forwarding rules.
- [ ] Run the focused tests and verify they fail before wiring.
- [ ] Load config for every create flow, forward `WorktreeFiles`, and render automatic setup as read-only information rather than a toggle.
- [ ] Run `go test ./cmd ./internal/tui` and verify green.

### Task 5: Real bare-repository regression coverage

**Files:**
- Modify: `internal/creator/e2e_test.go`
- Modify: `cmd/cli_e2e_test.go`

- [ ] Extend the e2e fixture with a mode-`0600` container-private template and nested worktree destination.
- [ ] Verify the test fails before creator wiring.
- [ ] Verify real `git worktree add` creation copies content and permissions while existing `env_files` still copy.
- [ ] Run `go test -tags=e2e ./internal/creator ./cmd`.

### Task 6: User documentation and summaries

**Files:**
- Modify: `README.md`
- Modify: `docs/PRD.md`

- [ ] Document the `worktree_files` schema, root semantics, overwrite policy, safety failures, and a `.codex/config.toml` example.
- [ ] Confirm CLI/TUI wording never prints source contents and distinguishes automatic worktree files from optional environment files.

### Task 7: Verification and release

- [ ] Run `gofmt` on changed Go files.
- [ ] Run focused packages, `go test -race ./...`, `go test -tags=e2e ./...`, `go vet ./...`, `golangci-lint run ./...`, `go build ./...`, and `goreleaser check`.
- [ ] Build a versioned branch binary and exercise real non-interactive creation.
- [ ] Refresh CCC and CRG, inspect impact, and review the full diff.
- [ ] Commit, push, create a PR, wait for required CI, fix failures test-first, and merge.
- [ ] Merge the release-please PR, wait for the tag/assets/tap update, upgrade Homebrew, verify `sentei --version`, and run a final installed-binary smoke test.
