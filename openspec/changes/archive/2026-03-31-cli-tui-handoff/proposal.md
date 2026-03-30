## Why

sentei currently has two disconnected interaction modes: the TUI (full interactive experience) and standalone CLI commands (`cleanup`, `ecosystems`, `integrations`) with their own flag parsing. Users cannot gradually transition from TUI to CLI — they must learn the CLI flags independently. Additionally, when a CLI command requires multiple flags and not all are provided, the only option is an error. A unified model where flags prefill TUI state and a single `--non-interactive` flag switches to headless mode creates a natural discoverability path (TUI → flags → full CLI) and eliminates the jarring mode boundary.

## What Changes

- Introduce a **command taxonomy**: commands are either "output" (read-only, always CLI) or "decision" (require choices, default to TUI)
- Add **`--non-interactive` flag** for decision commands to force headless execution (errors if required flags are missing; destructive operations also require `--force`)
- Flags become **form values** that skip TUI screens: partial flags enter the flow at the first missing required field; all required flags skip to a confirmation view
- Add **filter flags** for multi-select flows (e.g., `sentei remove --stale 30d`, `--merged`) that set the initial selection in the TUI list
- Add a **confirmation view pattern** to each decision command — a summary screen showing resolved options with the equivalent CLI command for discoverability
- Refactor each decision command to separate **option assembly** from TUI state, enabling both TUI and `--non-interactive` paths to share the same execution logic
- Add **teatest-based E2E tests** using real git repos in `t.TempDir()` to verify flag→TUI entry, confirmation views, and `--non-interactive` execution

## Capabilities

### New Capabilities
- `cli-tui-handoff`: Core routing logic — command taxonomy, `--non-interactive` flag, flag-to-TUI-entry mapping, and the "skip to step" behavior
- `confirmation-view`: Reusable confirmation view pattern that renders an options struct as a summary with the equivalent CLI command shown for discoverability
- `filter-flags`: Filter flag system for multi-select flows (`--stale`, `--merged`, `--all`) that resolve to initial selections in the TUI list

### Modified Capabilities
- `tui-confirmation`: Extend to support the new confirmation view pattern (currently only covers worktree deletion confirmation)
- `tui-list-view`: Extend to accept pre-selected items from filter flags
- `dry-run`: Subsume into `--non-interactive` model (dry-run becomes a composable flag rather than a standalone mode)
- `worktree-deletion`: Add `--stale`, `--merged`, `--all`, `--force` flags for non-interactive removal

## Impact

- **main.go**: Major refactor of CLI routing — replace ad-hoc if/else dispatch with command taxonomy router
- **cmd/**: Each decision command needs an options struct and flag definitions
- **internal/tui/**: Each flow needs a confirmation view; list view needs pre-selection support; model initialization needs "enter at step N" capability
- **internal/tui/model.go**: `NewMenuModel` signature changes to accept prefilled options
- **Testing**: Add `teatest` dependency; new E2E test suite covering the full flag×command matrix against real bare repos in temp dirs
- **Existing CLI commands**: `cleanup` already has separation; `ecosystems` and `integrations` are output commands (no changes needed beyond taxonomy classification)
