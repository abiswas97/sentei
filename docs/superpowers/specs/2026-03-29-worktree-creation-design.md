# Worktree Creation Flow

**Sub-project 2 of 3** for the sentei expansion from worktree cleanup tool to full worktree lifecycle manager.

**Scope:** TUI restructure (menu as entry point), worktree creation flow (branch input → options → progress → summary), creator pipeline, integration teardown on removal, phased parallel progress reporting, updated visual language.

**Depends on:** Sub-project 1 (config, ecosystem, integration packages).

---

## Architecture Overview

### Model Restructure

The existing flat `Model` struct is reorganized into grouped state:

```go
type Model struct {
    // Shared
    view       viewState
    runner     git.CommandRunner
    repoPath   string
    cfg        *config.Config
    width, height int

    // Menu
    menuCursor int
    menuItems  []menuItem

    // Remove flow
    remove removeState

    // Create flow
    create createState
}
```

`removeState` contains all existing removal fields (worktrees, selected, cursor, deletion progress, etc.). `createState` contains all new creation fields. Each view's logic stays in its own file.

### View States

```
menuView           → top-level menu (new entry point)
listView           → worktree removal list (existing)
confirmView        → removal confirmation (existing, enhanced with teardown info)
progressView       → removal progress (existing, enhanced with phased reporting)
summaryView        → removal summary (existing)
createBranchView   → branch name + base branch input (new)
createOptionsView  → setup toggles + integration toggles (new)
createProgressView → creation progress with phased reporting (new)
createSummaryView  → creation complete (new)
```

### Entry Point Change

`main.go` no longer eagerly loads worktrees. The TUI starts at `menuView` and loads data lazily when entering a specific flow. `--dry-run` continues to bypass the TUI entirely (loads worktrees directly as before).

### New Package

`internal/creator/` — worktree creation pipeline. Parallel to `internal/worktree/` (removal) and `internal/cleanup/`. Event-driven with the same emit callback pattern.

---

## Menu View

```
  sentei ─ Git Worktree Manager

  myproject (bare) · /Users/dev/code/myproject
  5 worktrees · 3 clean, 1 dirty, 1 locked

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  > Create new worktree
    Remove worktrees              5 available
    Cleanup                       safe mode

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  j/k navigate · enter select · q quit
```

**Context info:** Repo name (from directory name), repo type (bare), absolute path, worktree health summary (count, clean/dirty/locked breakdown).

**Context discovery on menu entry:**
- Repo name: `basename` of bare repo root
- Worktree count + health: `git worktree list --porcelain` with lightweight status check
- This runs async on `Init()` so the menu renders instantly with a spinner for context, then fills in

**Menu items:**
- "Create new worktree" — always available
- "Remove worktrees" — right-aligned count hint. Grayed if only bare root exists
- "Cleanup" — right-aligned mode hint

**Navigation:** `j/k` or arrows to navigate, `enter` to select, `q` to quit. `Esc` from any sub-flow returns here.

---

## Create Flow

### createBranchView

```
  sentei ─ Create Worktree

  myproject · /Users/dev/code/myproject

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  Branch name
  > feature/█

  Base branch
    main

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  enter continue · ctrl+enter quick create · tab switch field · esc back
```

**Branch name input:**
- Text input, focused by default
- Validation on submit: no spaces, no `..`, must not already exist as a worktree or local branch
- On validation failure: show inline error below the input field, keep focus

**Base branch:**
- Defaults to the repo's default branch (detected via `git symbolic-ref refs/remotes/origin/HEAD`, fallback to `main`)
- Tab to focus and type a different base
- Not validated until creation (any valid ref works)

**Quick create:** `ctrl+enter` skips the options view and goes straight to `createProgressView` with all defaults (all detected ecosystems enabled, integrations per `.sentei.yaml` config).

**Normal flow:** `enter` transitions to `createOptionsView`.

### createOptionsView

```
  sentei ─ Create Worktree

  feature/auth → from main

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  Setup

  [x] Install dependencies           pnpm detected
  [x] Merge default branch            main → feature/auth
  [x] Copy environment files           .env, .env.local

  Integrations

  [x] code-review-graph               installed
      Build code graph for AI-assisted review
      github.com/tirth8205/code-review-graph

  [ ] cocoindex-code                   installed
      Semantic code search index
      github.com/cocoindex-io/cocoindex-code

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  space toggle · enter create · esc back
```

**On entry:** Auto-detect ecosystems in the source worktree (usually main) via the ecosystem registry's `Detect()`. Check integration install status.

**Setup section:**
- "Install dependencies" — shown per detected ecosystem, all pre-checked. If multiple ecosystems detected (e.g., go + npm), each shown as a separate toggle line
- "Merge default branch" — pre-checked
- "Copy environment files" — shown if ecosystem defines `env_files`, pre-checked. Lists which files

**Integrations section:**
- All registered integrations shown
- Pre-checked state from `.sentei.yaml` `integrations_enabled`
- Right-aligned install status: "installed" (green) or "not found" (dim). If "not found" and toggled on, note: "will install automatically"
- Description and URL shown below each integration name (dimmed)

**Navigation:** `j/k` to move cursor, `space` to toggle, `enter` to start creation, `esc` to go back to branch input.

### createProgressView

Phased parallel progress with collapse behavior:

```
  sentei ─ Creating Worktree

  feature/auth → from main

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  Setup                                   3/3 ●
  Dependencies                            1/3
  ● pnpm install (root)
  ◐ pnpm install --filter packages/ui
  ◐ pnpm install --filter packages/core

  Integrations                         pending
  · code-review-graph
  · cocoindex-code

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
```

**Status indicators:**
- `●` done (green)
- `◐` active/in-progress (accent color, animated spinner)
- `·` pending (dim)
- `✗` failed (red)

**Phase headers:** Bold, with right-aligned counter (`2/3`) or status (`pending`). When a phase completes, its detail items collapse — the header shows `3/3 ●` on one line.

**Phases for creation:**
1. **Setup** — sequential: create worktree, merge base, copy env files
2. **Dependencies** — parallel: ecosystem install commands (bounded concurrency)
3. **Integrations** — sequential per integration (dependency resolution → install → setup)

**Failure behavior:** If a step in a phase fails, the phase header shows `2/3 ⚠`. The failed item shows `✗` with inline error. Remaining phases still attempt to run (best-effort).

**Non-interactive** — user watches progress. On all phases complete → auto-transition to `createSummaryView`.

### createSummaryView

```
  sentei ─ Worktree Created

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  ● feature/auth ready

    Path     /Users/dev/code/myproject/feature-auth
    Branch   feature/auth (from main)
    Deps     pnpm ●
    Index    code-review-graph ●

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

    cd /Users/dev/code/myproject/feature-auth

  enter menu · q quit
```

**Success:** Green `●` header, key-value summary, `cd` path for copy-paste.

**With failures:**

```
  ⚠ feature/auth created with issues

    Path     /Users/dev/code/myproject/feature-auth
    Branch   feature/auth (from main)
    Deps     pnpm ●
    Index    code-review-graph ✗  build timed out

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

    cd /Users/dev/code/myproject/feature-auth

  enter menu · q quit
```

**Navigation:** `enter` returns to menu, `q` quits sentei.

---

## Creator Pipeline (`internal/creator/`)

### Types

```go
type Options struct {
    BranchName     string
    BaseBranch     string
    RepoPath       string                    // bare repo root
    SourceWorktree string                    // worktree to copy env files from
    MergeBase      bool
    CopyEnvFiles   bool
    Ecosystems     []config.EcosystemConfig  // enabled ecosystems to install
    Integrations   []integration.Integration // enabled integrations to set up
}

type Result struct {
    WorktreePath string
    Steps        []StepResult
}

type StepResult struct {
    Name    string
    Status  StepStatus
    Message string
    Error   error
}

type StepStatus int
const (
    StepPending StepStatus = iota
    StepRunning
    StepDone
    StepFailed
    StepSkipped
)

type Phase struct {
    Name  string
    Steps []StepResult
}

type Event struct {
    Phase   string
    Step    string
    Status  StepStatus
    Message string
    Error   error
}

func Run(runner git.CommandRunner, opts Options, emit func(Event)) Result
```

### Pipeline Phases

**Phase 1: Setup (sequential)**
1. Create worktree: `git worktree add <bare-root>/<sanitized-branch> -b <branch>`
   - Path sanitization: `feature/auth` → `feature-auth` (replace `/` with `-`)
2. Merge base branch: `git -C <worktree-path> merge <baseBranch> --no-edit` (if `MergeBase` is true)
3. Copy env files: for each file in ecosystem's `env_files`, copy from `SourceWorktree` to new worktree (if file exists in source). Skip silently if source file doesn't exist.

**Phase 2: Dependencies (parallel)**
1. For each enabled ecosystem:
   - If ecosystem has `workspace_detect`, call `DetectWorkspaces()` to find sub-directories
   - If workspaces found and `IsParallel()`, run workspace install commands concurrently (bounded by semaphore, max 5)
   - Otherwise run the ecosystem's `install.command` once from worktree root
2. Emit events per install command (started, completed, failed)

**Phase 3: Integrations (sequential per integration)**

For each enabled integration:
1. Check if installed via `Detect` spec
2. If not installed:
   a. Check dependencies in order via their `Detect` commands
   b. Missing dependency → emit event, run `dep.Install`
   c. Run integration `Install.Command`
   d. Re-check `Detect` → if still fails, emit error event with PATH guidance
3. Run `Setup.Command`:
   - If `WorkingDir` is `"repo"`: replace `{path}` in command with worktree absolute path, run from repo root
   - If `WorkingDir` is `"worktree"`: run command from within the worktree directory
4. Append `GitignoreEntries` to worktree `.gitignore` if not already present

### Source Worktree Discovery

The `SourceWorktree` (where env files are copied from) is determined by:
1. Parse `git worktree list --porcelain`
2. Find the worktree whose branch matches the repo's default branch
3. Fallback: first non-bare worktree

### Error Handling

- **Create worktree fails** (branch exists, path exists): abort entire pipeline. Return error immediately.
- **Merge fails** (conflict): emit warning, continue. Worktree is usable, user resolves conflict manually.
- **Env file copy fails** (permission): emit warning, continue. Non-critical.
- **Dep install fails**: emit error for that ecosystem, continue to next ecosystem / integrations.
- **Integration install fails**: emit error, skip that integration's setup, continue to next integration.
- **Integration setup fails**: emit error, continue to next integration.

---

## Enhanced Removal Flow

### Confirm View (with integration teardown info)

Before displaying the confirmation dialog, scan each selected worktree for integration `TeardownSpec.Dirs`:

```
  sentei ─ Confirm Deletion

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  Removing 3 worktrees:

    bugfix/login-redirect              clean
    experiment/new-ui                  ⚠ uncommitted changes
    chore/deps-update                  clean

  Cleaning up:

    .code-review-graph/     in 2 worktrees
    .cocoindex_code/        in 1 worktree

  ⚠ 1 worktree has uncommitted changes that will be lost

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  y confirm · n go back
```

The "Cleaning up" section only appears if any integration artifacts are found. Teardown is automatic — no toggles.

### Removal Progress (phased)

The removal progress view adopts the same phased reporting as creation:

```
  sentei ─ Removing Worktrees

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  Teardown                                3/3 ●

  Removing worktrees                      1/3
  ● bugfix/login-redirect
  ◐ experiment/new-ui
  ◐ chore/deps-update

  Prune & cleanup                      pending

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
```

**Phases for removal:**
1. **Teardown** — parallel (max 5): for each worktree, run integration teardown commands / delete artifact dirs
2. **Removing worktrees** — parallel (max 5): `git worktree remove --force` (existing behavior)
3. **Prune & cleanup** — sequential: `git worktree prune`, then `cleanup.Run()` (existing behavior)

The teardown phase runs per-worktree. For each worktree:
1. Check which integration `TeardownSpec.Dirs` exist
2. If `TeardownSpec.Command` is set, run it from the worktree dir
3. If command fails, fall back to deleting the dirs directly

---

## Visual Language Updates

### Consistent across all views

**Indicators:**
- `●` done (green, ANSI 42)
- `◐` active (accent, ANSI 62)
- `·` pending (dim, ANSI 241)
- `✗` failed (red, ANSI 196)
- `⚠` warning (yellow, ANSI 214)

**Layout:**
- Dotted separators (`┄`) instead of box borders
- Title line: `sentei ─ <View Name>`
- Status bar at bottom with key hints
- Right-aligned context hints on menu items and toggles

**Styles (extending `internal/tui/styles.go`):**
- `stylePhaseDone` — green header
- `stylePhaseActive` — bold, accent
- `stylePhasePending` — dim
- `styleIndicatorDone` — green `●`
- `styleIndicatorActive` — accent `◐`
- `styleIndicatorPending` — dim `·`
- `styleIndicatorFailed` — red `✗`
- `styleSeparator` — dim dotted line

---

## Testing Strategy

### Unit Tests

**Creator pipeline (`internal/creator/`):**
- Each step tested with mock `CommandRunner`
- Table-driven: successful creation, branch already exists (abort), merge conflict (warning + continue), env file copy (source missing → skip), dep install failure (continues)
- Event emission: assert correct phase/step/status events in order
- Parallel workspace install: verify concurrent execution with bounded parallelism
- Source worktree discovery: mock porcelain output, verify correct worktree selected
- Path sanitization: `feature/auth` → `feature-auth`, `bugfix/login` → `bugfix-login`

**Integration teardown (`internal/creator/teardown.go`):**
- Scan worktrees for artifacts — temp dirs with/without `.code-review-graph/`, `.cocoindex_code/`
- Teardown command execution from correct working dir
- Fallback to dir deletion when teardown command fails
- No artifacts found → no-op, no events emitted

**TUI views:**
- `menu_test.go` — navigation, item rendering, context info, grayed items
- `create_branch_test.go` — validation (spaces, `..`, duplicate branch), tab between fields, quick create shortcut
- `create_options_test.go` — toggle state, ecosystem display, integration status display, defaults from config
- `create_progress_test.go` — phase transitions, collapse on completion, parallel item tracking, failure display
- `create_summary_test.go` — success rendering, failure rendering, cd path display

**Model restructure:**
- All existing removal tests continue to pass after grouping fields into `removeState`
- View transitions: menu → create → back, menu → remove → back

### E2E Tests

- Creator pipeline: create temp bare repo with a main worktree containing `go.mod`, run full pipeline, verify new worktree exists with correct branch, merged base, env files copied
- Teardown: create temp worktree with integration artifact dirs, run teardown, verify dirs removed
- Full TUI smoke test: build binary, verify `sentei` starts without error in a bare repo

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/creator/creator.go` | Create | `Options`, `Result`, `Event` types; `Run()` orchestrator |
| `internal/creator/setup.go` | Create | Create worktree, merge, copy env files |
| `internal/creator/deps.go` | Create | Ecosystem dependency installation (parallel workspace support) |
| `internal/creator/integrations.go` | Create | Integration dependency resolution, install, setup |
| `internal/creator/teardown.go` | Create | Scan artifacts, run teardown commands, fallback deletion |
| `internal/creator/creator_test.go` | Create | Pipeline unit tests |
| `internal/creator/teardown_test.go` | Create | Teardown unit tests |
| `internal/tui/model.go` | Modify | Restructure into grouped state, add menuView + create view states |
| `internal/tui/menu.go` | Create | Menu view rendering and update logic |
| `internal/tui/create_branch.go` | Create | Branch input view |
| `internal/tui/create_options.go` | Create | Options/toggles view |
| `internal/tui/create_progress.go` | Create | Phased progress view for creation |
| `internal/tui/create_summary.go` | Create | Creation summary view |
| `internal/tui/progress.go` | Modify | Refactor to use phased reporting (shared with create) |
| `internal/tui/confirm.go` | Modify | Add integration teardown info section |
| `internal/tui/styles.go` | Modify | Add phase/indicator styles, separator style |
| `internal/tui/keys.go` | Modify | Add menu and create flow key bindings |
| `internal/tui/list.go` | Modify | Access worktrees via `m.remove.worktrees` |
| `internal/tui/summary.go` | Modify | Access results via `m.remove.*` |
| `main.go` | Modify | Lazy loading — start at menu, don't eagerly list worktrees |

---

## Sub-project Boundaries

**In scope:**
- Model restructure with grouped state
- Menu view as new entry point
- Full create flow (4 views)
- Creator pipeline with phased progress
- Integration teardown on removal
- Visual language update (indicators, separators, phase headers)

**Not in scope (sub-project 3):**
- Repo creation, GitHub publish
- Clone-as-bare, migrate-to-bare
- Non-git-repo detection flow
- Non-bare-repo detection and migration offer
