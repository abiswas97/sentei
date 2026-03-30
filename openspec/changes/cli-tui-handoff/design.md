## Context

sentei is a single-binary TUI tool for managing git worktrees. It currently has two disconnected interaction modes: a full Bubble Tea TUI launched by `sentei` (or `sentei <path>`), and standalone CLI commands (`cleanup`, `ecosystems`, `integrations`) with their own flag parsing in `cmd/`. The TUI is a 21-state state machine driven by `viewState` in `internal/tui/model.go`. Business logic lives in internal packages (`creator`, `cleanup`, `repo`, `integration`) that communicate via channel-based events.

The CLI commands that exist today (`cleanup`) already separate option assembly from execution. The TUI flows (create worktree, remove worktrees, repo operations) have their logic interleaved with TUI state — option construction happens inside Bubble Tea message handlers.

## Goals / Non-Goals

**Goals:**
- Unify CLI and TUI under a single mental model: flags are form values, `--non-interactive` forces headless mode
- Enable a natural discoverability path: TUI → learn flags → use CLI
- Each decision command has both an interactive (TUI) and non-interactive (CLI) path sharing the same execution logic
- Comprehensive E2E testing via teatest against real git repos

**Non-Goals:**
- Machine-readable output formats (JSON, YAML) for CLI mode — future work
- Configurable keybindings
- Remote/daemon mode
- Plugin system for custom commands

## Decisions

### 1. Command taxonomy determines default mode

Commands are classified as **output** (read-only, always CLI) or **decision** (require choices, default to TUI). The classification is inherent to the command, not toggled by a flag.

| Type | Commands | Default | `--non-interactive` |
|------|----------|---------|---------------------|
| Output | `ecosystems`, `integrations`, `--version` | CLI | Redundant (no-op) |
| Decision | `create`, `remove`, `cleanup`, `clone`, `migrate`, root menu | TUI | Forces CLI mode |

**Why over a `--tui` opt-in flag:** The TUI is sentei's primary interface. Making CLI opt-in (via `--non-interactive`) rather than TUI opt-in preserves the tool's identity and makes the discoverability path natural.

### 2. Flags skip TUI screens, not prefill form fields

When flags are provided for a decision command, they build an **options struct** directly and skip the TUI screens that would have collected those values. This avoids the "hydration problem" — reconstructing async state accumulated through screen transitions.

Three entry modes:
- **No flags**: Full multi-step TUI flow from the beginning
- **Partial flags**: Enter the TUI at the first screen with a missing required value. Earlier screens are skipped; their values come from flags
- **All required flags**: Skip to a confirmation view that renders the options struct as a summary

**Why over literal form prefill:** The TUI state machine has async dependencies between screens (worktree loading, ecosystem detection, integration state). Jumping into a mid-flow screen would require a hydration function per flow per entry point. Skipping screens and constructing the options struct directly sidesteps this entirely.

**Implementation approach:** Each decision command gets an `Options` struct (some already exist, e.g., `cleanup.Options`). A new `ParseFlags(args []string) (*Options, error)` function per command builds the struct from flags. The TUI flow checks which fields are populated to determine the entry screen. The confirmation view and `--non-interactive` path both consume the same `Options` struct.

### 3. Confirmation view pattern with CLI command echo

Each decision command gets a confirmation view — a summary screen showing the resolved options. Below the summary, the equivalent CLI command is displayed so users can copy it for future use or scripting.

The confirmation view is a single reusable component parameterized by:
- Title (e.g., "Create Worktree")
- Key-value pairs to display
- The CLI command string (generated from the options struct)

**Why a shared component:** All confirmation views have the same structure (title, key-value summary, command echo, enter/esc/q). A single `ConfirmationViewModel` avoids duplicating this across flows.

### 4. Filter flags for multi-select flows

The `remove` command's TUI is a checkbox list, not a form. Flags act as **filters** that set the initial selection rather than specifying individual items:

- `--stale <duration>` — pre-select worktrees with no commits in the given period
- `--merged` — pre-select worktrees whose branches are fully merged into the default branch
- `--all` — pre-select all non-protected worktrees

Filters are additive (OR logic). The TUI list always appears (without `--non-interactive`) so users can adjust the selection. With `--non-interactive`, the resolved selection executes directly.

**Why filters over item selectors:** Users rarely know exact worktree paths. Filters match the mental model ("clean up old stuff") rather than requiring precise identification.

### 5. `--non-interactive` + `--force` for destructive operations

Two flags control non-interactive behavior:
- `--non-interactive` — suppress TUI, execute from flags only, error if required flags are missing
- `--force` — required alongside `--non-interactive` for destructive operations (deletion, migration)

**Why two flags instead of one:** Separating "don't show TUI" from "yes I'm sure" prevents accidental destructive execution in scripts. A typo adding `--non-interactive` to a cleanup command shouldn't delete branches without the explicit `--force` gate.

### 6. Incremental rollout by flow

Each flow is refactored independently. Order is determined by existing separation quality and complexity:

| Phase | Flow | Rationale |
|-------|------|-----------|
| 0 | Architecture | Options struct pattern, `--non-interactive` flag, confirmation view component, teatest setup |
| 1 | `cleanup` | Already has CLI/business logic separation. 3 flat flags. Proof of concept. |
| 2 | `create` | Requires extracting option assembly from TUI. Medium effort. High user value. |
| 3 | `clone` | Simple inputs (URL, name). Low effort. |
| 4 | `remove` | Filter flags, pre-selection logic. Medium effort. |
| 5 | `migrate` | Complex multi-phase with backup decisions. Highest effort. |

Each phase is independently shippable and includes its own tests.

### 7. Testing strategy: teatest as E2E workhorse

Three test layers:
- **Unit tests**: Flag parsing → options struct, routing logic, filter resolution. No TUI, no git.
- **teatest E2E tests**: Real Bubble Tea program, real git repos in `t.TempDir()`. Send keystrokes, assert on rendered output. Covers: flag→TUI entry, confirmation views, filter pre-selection, full flow walkthroughs.
- **Binary E2E tests**: `exec.Command` for `--non-interactive` paths (no TUI to interact with). Extends existing `cmd/cli_e2e_test.go`.

Each test gets its own `t.TempDir()` with a fresh bare repo + worktrees. Go cleans up temp dirs automatically.

**Why not Docker:** `t.TempDir()` is sufficient for git repo isolation, fast, and has no external dependencies. Docker adds CI complexity without meaningful benefit for this use case.

### 8. Command routing refactor

Replace the current ad-hoc if/else chain in `main.go` with a command registry:

```go
type Command struct {
    Name        string
    Type        CommandType  // Output or Decision
    RunCLI      func(args []string) error
    BuildTUI    func(opts *Options, model *Model) // configure TUI entry point
    ParseFlags  func(args []string) (*Options, error)
}
```

The main function becomes: parse command → classify → if output, run CLI; if decision, parse flags → if `--non-interactive`, run CLI with options; else build TUI with options.

**Why over keeping if/else:** The current dispatch grows linearly with commands and duplicates the "parse flags, decide mode" logic per command. A registry centralizes this.

## Risks / Trade-offs

**[Risk] Partial flag entry may confuse users** → Mitigation: The TUI clearly shows which fields are already set (non-editable or visually distinct). Help text explains the flag that set each value.

**[Risk] teatest maturity** → Mitigation: teatest is maintained by Charmbracelet (same team as Bubble Tea). It's experimental (`x/exp`) but actively used in their own projects. If it proves insufficient, we can fall back to golden file tests with programmatic model updates.

**[Risk] Confirmation view adds a step to the TUI flow** → Mitigation: For the normal TUI path (no flags), flows that already have a natural confirmation step (like worktree deletion) keep their existing behavior. Only flows that currently skip straight to execution (like create worktree with Ctrl+Enter) get a new confirmation screen — and that screen can be skipped with Enter immediately.

**[Risk] `--non-interactive` path diverges from TUI path over time** → Mitigation: Both paths consume the same `Options` struct and call the same execution functions. The only difference is how the struct is populated (flags vs TUI screens) and how results are displayed (stdout vs TUI views).

**[Trade-off] Incremental rollout means inconsistent UX during development** → Accepted: Some commands will support flags-as-form-values before others. This is preferable to a big-bang release that touches all flows simultaneously.
