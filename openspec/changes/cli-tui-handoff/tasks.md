## 1. Architecture & Foundation

- [x] 1.1 Define `CommandType` enum (Output, Decision) and `Command` struct with Name, Type, ParseFlags, RunCLI, BuildTUI fields
- [x] 1.2 Create command registry in `internal/cli/registry.go` with Register and Dispatch functions
- [x] 1.3 Add `--non-interactive` flag parsing to the registry dispatch (applies to all decision commands)
- [x] 1.4 Add `--force` flag parsing; enforce requirement for destructive `--non-interactive` operations
- [x] 1.5 Refactor `main.go` to use command registry instead of ad-hoc if/else dispatch
- [x] 1.6 Register existing output commands (ecosystems, integrations, version) in registry
- [x] 1.7 Unit tests for registry dispatch: known command, unknown command, root (no command), output vs decision routing

## 2. Confirmation View Component

- [x] 2.1 Create `internal/tui/confirmation.go` with reusable `ConfirmationViewModel` — accepts title, key-value pairs, CLI command string
- [x] 2.2 Implement View rendering: title, key-value rows, separator, CLI command echo, keybindings footer
- [x] 2.3 Implement Update: Enter → proceed (returns a `ConfirmMsg`), Esc → back, q/Ctrl+C → quit
- [x] 2.4 Add `CLICommand() string` method to each options struct that generates the equivalent CLI command from its fields
- [x] 2.5 Unit tests for confirmation view: rendering, key handling, CLI command generation

## 3. Phase 1 — Cleanup Command

- [x] 3.1 Define `CleanupOptions` struct (Mode, Force, DryRun) if not already present; add `ParseFlags` function
- [x] 3.2 Register cleanup as a decision command in the registry
- [x] 3.3 Add cleanup confirmation view: display mode, force, dry-run settings with CLI command echo
- [x] 3.4 Implement TUI entry: no flags → full flow, partial flags → enter at first missing, all flags → confirmation view
- [x] 3.5 Implement `--non-interactive` path: parse flags → validate required (mode) → execute → print results to stdout
- [x] 3.6 Unit tests for cleanup flag parsing, option validation, missing flag errors
- [x] 3.7 teatest E2E: launch cleanup TUI with no flags, navigate full flow, verify screens
- [x] 3.8 teatest E2E: launch cleanup TUI with `--mode safe`, verify enters at confirmation view
- [x] 3.9 Binary E2E: `sentei cleanup --mode safe --non-interactive` against real bare repo, verify stdout output
- [x] 3.10 Binary E2E: `sentei cleanup --non-interactive` (missing --mode), verify error message and exit code 1

## 4. Phase 2 — Create Worktree Command

- [x] 4.1 Extract `CreateOptions` struct from TUI state (Branch, Base, Ecosystems, MergeBase, CopyEnvFiles)
- [x] 4.2 Add `ParseFlags` for create command flags (--branch, --base, --ecosystems, --merge-base, --copy-env)
- [x] 4.3 Register create as a decision command in the registry
- [x] 4.4 Refactor create flow to determine entry screen based on which options are populated
- [x] 4.5 Add create confirmation view using the shared component
- [x] 4.6 Implement `--non-interactive` path: parse flags → validate (branch + base required) → execute → print results
- [x] 4.7 Unit tests for create flag parsing, option validation, screen-skip logic
- [x] 4.8 teatest E2E: `sentei create --branch foo` → verify enters at base selection step
- [x] 4.9 teatest E2E: `sentei create --branch foo --base main` → verify enters at confirmation view
- [x] 4.10 Binary E2E: `sentei create --branch foo --base main --non-interactive` against real bare repo

## 5. Phase 3 — Clone Command

- [x] 5.1 Define `CloneOptions` struct (URL, Name) with `ParseFlags`
- [x] 5.2 Register clone as a decision command in the registry
- [x] 5.3 Add clone confirmation view using the shared component
- [x] 5.4 Implement TUI entry: no flags → full flow, --url → enter at name step, all flags → confirmation
- [x] 5.5 Implement `--non-interactive` path: parse flags → validate (url required) → execute → print results
- [x] 5.6 Unit tests for clone flag parsing
- [x] 5.7 teatest E2E: `sentei clone --url <url>` → verify enters at name step
- [x] 5.8 Binary E2E: `sentei clone --url <url> --name myrepo --non-interactive` → verify clone executes

## 6. Phase 4 — Remove Command with Filter Flags

- [x] 6.1 Define `RemoveOptions` struct (Stale duration, Merged bool, All bool, Force bool)
- [x] 6.2 Add `ParseFlags` for remove with `--stale`, `--merged`, `--all` filter flags
- [x] 6.3 Implement duration parsing for `--stale` (supports `Nd`, `Nw`, `Nm` formats)
- [x] 6.4 Implement filter resolution: given enriched worktrees and filter flags, produce a selection set
- [x] 6.5 Register remove as a decision command in the registry
- [x] 6.6 Wire filter-based pre-selection into the TUI list view (pass initial selection set to model)
- [x] 6.7 Add filter indicator to list view status bar (e.g., "filter: stale > 30d, merged")
- [x] 6.8 Implement `--non-interactive` path: parse flags → resolve selection → validate (--force required, non-empty selection) → execute → print summary
- [x] 6.9 Unit tests for duration parsing, filter resolution, composable filter OR logic, protected worktree exclusion
- [x] 6.10 teatest E2E: `sentei remove --merged` → verify list view with merged worktrees pre-selected
- [x] 6.11 teatest E2E: `sentei remove --all` → verify all non-protected worktrees pre-selected
- [x] 6.12 Binary E2E: `sentei remove --merged --force --non-interactive` → verify deletion and summary output
- [x] 6.13 Binary E2E: `sentei remove --merged --non-interactive` (missing --force) → verify error

## 7. Phase 5 — Migrate Command

- [x] 7.1 Define `MigrateOptions` struct (DeleteBackup bool, Force bool) with `ParseFlags`
- [x] 7.2 Register migrate as a decision command in the registry
- [x] 7.3 Add migrate confirmation view using the shared component (show current repo, target bare repo path)
- [x] 7.4 Implement `--non-interactive` path: parse flags → validate → execute → print results
- [x] 7.5 Unit tests for migrate flag parsing
- [x] 7.6 teatest E2E: `sentei migrate` → verify full TUI flow
- [x] 7.7 Binary E2E: `sentei migrate --force --non-interactive` → verify migration executes

## 8. Dry-Run Integration

- [x] 8.1 Make `--dry-run` composable with decision commands: `sentei cleanup --mode aggressive --dry-run --non-interactive` prints what would happen
- [x] 8.2 Add dry-run support to remove's `--non-interactive` path: list worktrees that would be deleted without deleting
- [x] 8.3 Unit tests for dry-run output formatting per command
- [x] 8.4 Binary E2E: `sentei remove --merged --dry-run --non-interactive` → verify preview output, no deletions

## 9. teatest Infrastructure

- [x] 9.1 Add `github.com/charmbracelet/x/exp/teatest` dependency
- [x] 9.2 Create test helper: `setupBareRepo(t *testing.T) string` — creates a bare repo with N worktrees in `t.TempDir()`
- [x] 9.3 Create test helper: `setupBareRepoWithState(t *testing.T, opts RepoOpts) string` — configurable worktree count, dirty state, stale dates, merged branches
- [x] 9.4 Create test helper: `launchTUI(t *testing.T, args ...string) *teatest.TestModel` — builds and launches sentei with given args
- [x] 9.5 Verify teatest setup works with a smoke test: launch TUI, send 'q', assert clean exit
