# Config System, Ecosystem Detection & Integrations

**Sub-project 1 of 3** for the sentei expansion from worktree cleanup tool to full worktree lifecycle manager.

**Scope:** Foundation layer only — config loading, ecosystem registry, integration registry. No TUI changes. Sub-projects 2 (worktree creation) and 3 (repo operations) will wire these into the TUI.

---

## Architecture Overview

Three new packages under `internal/`:

```
internal/
├── config/          # Config schema, loading, merging
│   └── defaults/    # go:embed YAML files
├── ecosystem/       # Package manager detection and workspace resolution
└── integration/     # Built-in tool lifecycle (code-review-graph, ccc)
```

Existing packages (`git/`, `tui/`, `worktree/`, `cleanup/`) are untouched.

---

## Config System

### Three-Layer Loading

1. **Embedded defaults** — YAML files in `internal/config/defaults/` compiled via `go:embed`. Ship ecosystem definitions.
2. **Global config** — `~/.config/sentei/config.yaml` (respects `XDG_CONFIG_HOME`). User-wide overrides and additions.
3. **Per-repo config** — `.sentei.yaml` at the bare repo root. Repo-specific settings.

### Merge Semantics

- **Ecosystems:** Keyed by `name`. Per-repo can override a specific ecosystem's fields or disable it with `enabled: false`. New entries are appended.
- **Scalar values** (e.g., `protected_branches`, `integrations_enabled`): Replace entirely. If per-repo defines `protected_branches`, it replaces the global list.
- **Integrations enabled:** A list of integration names that are pre-checked in the TUI when creating a worktree.

### Config Discovery

`LoadConfig(repoPath string) (*Config, error)` is the single entry point.

For per-repo config: if `repoPath` is inside a worktree, resolve to the bare repo root via `git rev-parse --git-common-dir`, then look for `.sentei.yaml` there.

### Go Types

```go
type Config struct {
    Ecosystems        []EcosystemConfig `yaml:"ecosystems"`
    ProtectedBranches []string          `yaml:"protected_branches"`
    IntegrationsEnabled []string        `yaml:"integrations_enabled"`
}

type EcosystemConfig struct {
    Name             string   `yaml:"name"`
    Enabled          *bool    `yaml:"enabled,omitempty"`  // nil = enabled
    Detect           DetectConfig   `yaml:"detect"`
    Install          InstallConfig  `yaml:"install"`
    EnvFiles         []string `yaml:"env_files"`
    PostInstall      []string `yaml:"post_install"`
}

type DetectConfig struct {
    Files []string `yaml:"files"`  // presence of any file → match
}

type InstallConfig struct {
    Command           string `yaml:"command"`
    WorkspaceDetect   string `yaml:"workspace_detect,omitempty"`
    WorkspaceInstall  string `yaml:"workspace_install,omitempty"`
    Parallel          bool   `yaml:"parallel,omitempty"`
}
```

### Error Handling

- Missing global config → use embedded defaults only. Not an error.
- Missing per-repo config → use global + embedded. Not an error.
- Malformed YAML → return error with file path and parse error details. Do not silently fall back.
- Unknown fields → warn to stderr, do not error. Supports forward compatibility.

### Validation

After merging all layers:
- Each ecosystem must have a `name` and at least one `detect.files` entry.
- `integrations_enabled` entries must reference known integration names. Unknown names produce a warning (not error) for forward compatibility.

---

## Ecosystem Registry

### What It Is

A priority-ordered list of package managers / build tools that sentei can detect in a worktree and run dependency installation for. Each entry is declarative YAML — no Go code needed to add a new ecosystem.

### Detection Logic

Given a worktree path, iterate ecosystems in list order. All matches are returned (a repo can have `go.mod` + `package.json`). List order determines priority — embedded defaults set a sensible order, user config can reorder. Sub-project 2's TUI will present all detected ecosystems and let the user choose which to run.

For workspaces: if `workspace_detect` is set, check for the workspace config file. If found, discover sub-directories that need installation. `{dir}` in `workspace_install` is replaced with each sub-directory path.

### Day-1 Ecosystems

Listed in priority order. Detection stops at first match within a language group (e.g., pnpm beats yarn beats npm).

| Name | Detect Files | Install Command | Workspace Detect | Workspace Install |
|------|-------------|-----------------|------------------|-------------------|
| pnpm | `pnpm-lock.yaml` | `pnpm install` | `pnpm-workspace.yaml` | `pnpm install --filter {dir}` |
| yarn | `yarn.lock` | `yarn install` | `package.json` (workspaces) | `yarn workspace {dir} install` |
| npm | `package-lock.json` | `npm install` | `package.json` (workspaces) | `npm install --workspace {dir}` |
| bun | `bun.lockb` | `bun install` | `package.json` (workspaces) | `bun install --filter {dir}` |
| cargo | `Cargo.toml` | `cargo build` | `Cargo.toml` (workspace.members) | — |
| go | `go.mod` | `go mod download` | `go.work` | — |
| uv | `uv.lock` | `uv sync` | — | — |
| poetry | `poetry.lock` | `poetry install` | — | — |
| pip | `requirements.txt` | `pip install -r requirements.txt` | — | — |
| ruby | `Gemfile.lock` | `bundle install` | — | — |
| php | `composer.lock` | `composer install` | — | — |
| dotnet | `*.sln`, `*.csproj` | `dotnet restore` | — | — |
| elixir | `mix.lock` | `mix deps.get` | — | — |
| swift | `Package.swift` | `swift package resolve` | — | — |
| dart | `pubspec.lock` | `dart pub get` | — | — |
| deno | `deno.lock` | `deno install` | — | — |

### Workspace Detection Details

For ecosystems with `workspace_detect`, the detection logic:

1. Check if the workspace config file exists at the worktree root.
2. Parse it to extract workspace member directories:
   - **pnpm:** Parse `pnpm-workspace.yaml` → `packages:` list (supports globs).
   - **yarn/npm/bun:** Parse `package.json` → `workspaces` field if present (supports globs). If `workspaces` field is absent, treat as non-workspace.
   - **cargo:** Parse `Cargo.toml` → `[workspace] members` list.
   - **go:** Parse `go.work` → `use` directives.
3. Resolve globs to actual directories.
4. Return list of directories needing installation.

Workspace parsing logic lives in `internal/ecosystem/workspace.go` with per-format parsers. This is Go code (not config) because parsing TOML/YAML/JSON workspace formats requires language-specific logic.

### Registry Interface

```go
type Registry struct {
    ecosystems []Ecosystem
}

func NewRegistry(cfg []EcosystemConfig) *Registry
func (r *Registry) Detect(worktreePath string) ([]Ecosystem, error)
func (r *Registry) All() []Ecosystem
```

`Detect` returns all matched ecosystems in priority order. Multiple matches are possible (e.g., `go.mod` + `package.json` in the same repo). Sub-project 2's TUI will let the user choose which to run.

No ecosystem detected → empty slice, not an error.

### Edge Cases

- Detection file exists but workspace config is malformed → log warning, treat as non-workspace. Do not fail the whole operation.
- Glob patterns in workspace config resolve to no directories → treat as non-workspace for that ecosystem.
- Multiple ecosystems of the same language (e.g., both `pnpm-lock.yaml` and `package-lock.json`) → both are detected. TUI (sub-project 2) will present them in priority order and let the user pick.

### Auditability

`sentei ecosystems` CLI command: lists all registered ecosystems with name, detection files, install command, source (embedded/global/per-repo), and enabled/disabled status. Formatted table output to stdout.

---

## Integrations

### What They Are

Built-in, hardcoded tool integrations that sentei can set up when creating a worktree and tear down when removing one. Unlike ecosystems (config-driven, auto-detected), integrations are:
- Defined in Go code in `internal/integration/`
- Explicitly toggled by the user in the TUI
- Not user-extensible via config (only enable/disable per-repo)

### Go Types

```go
type Integration struct {
    Name         string
    Description  string
    URL          string       // project homepage, shown in TUI
    Dependencies []Dependency
    Detect       DetectSpec
    Install      InstallSpec
    Setup        SetupSpec
    Teardown     TeardownSpec
    GitignoreEntries []string
}

type Dependency struct {
    Name    string
    Detect  string // command to check (e.g., "pipx --version")
    Install string // install command (e.g., "brew install pipx")
}

type DetectSpec struct {
    Command    string // e.g., "code-review-graph --version"
    BinaryName string // fallback: check if binary is on PATH
}

type InstallSpec struct {
    Command      string
    FirstRunNote string // shown in TUI before install (e.g., "Downloads ~87MB model")
}

type SetupSpec struct {
    Command    string // the setup command(s)
    WorkingDir string // "repo" = run from repo root, "worktree" = must cd into worktree
}

type TeardownSpec struct {
    Command string   // preferred teardown command (e.g., "ccc reset --all --force")
    Dirs    []string // fallback: delete these dirs if command unavailable
}
```

### Interface

```go
func All() []Integration          // returns all registered integrations
func Get(name string) *Integration // lookup by name, nil if not found
```

Each integration is defined as a Go struct literal in its own file (`crg.go`, `ccc.go`).

### Day-1 Integrations

#### code-review-graph

| Field | Value |
|-------|-------|
| Name | `code-review-graph` |
| Description | Build code graph for AI-assisted code review |
| URL | `https://github.com/tirth8205/code-review-graph` |
| Dependencies | Python 3.10+ (detect: `python3 -c "import sys; assert sys.version_info >= (3,10)"`, assume pre-installed — error with install guidance if missing), pipx (detect: `pipx --version`, install: `brew install pipx` on macOS, `pip install --user pipx` elsewhere) |
| Detect | command: `code-review-graph --version` |
| Install | `pipx install code-review-graph` |
| Setup | command: `code-review-graph build --repo {path}`, working dir: `repo` (supports `--repo` flag) |
| Teardown | dirs: [`.code-review-graph/`] |
| Gitignore | [`.code-review-graph/`] |

Notes:
- No API keys needed for core functionality.
- ~10s build time for 500 files.
- The tool auto-creates `.code-review-graph/.gitignore` with `*` inside.

#### cocoindex-code

| Field | Value |
|-------|-------|
| Name | `cocoindex-code` |
| Description | Semantic code search index |
| URL | `https://github.com/cocoindex-io/cocoindex-code` |
| Dependencies | Python 3.11+ (detect: `python3 -c "import sys; assert sys.version_info >= (3,11)"`, assume pre-installed — error with install guidance if missing), uv (detect: `uv --version`, install: `brew install uv` on macOS, `curl -LsSf https://astral.sh/uv/install.sh \| sh` elsewhere) |
| Detect | binary name: `ccc` (no `--version` flag available) |
| Install | `uv tool install --upgrade cocoindex-code --prerelease explicit --with "cocoindex>=1.0.0a24"` |
| First-run note | "Downloads ~87MB embedding model on first use" |
| Setup | command: `ccc init && ccc index`, working dir: `worktree` (must cd into worktree, no `--repo` flag) |
| Teardown | command: `ccc reset --all --force` (from worktree dir), fallback dirs: [`.cocoindex_code/`] |
| Gitignore | [`.cocoindex_code/`] |

Notes:
- Runs a background daemon (~430MB RSS) that auto-starts and persists across projects.
- First run downloads ~87MB embedding model from HuggingFace.
- Incremental indexing — subsequent `ccc index` calls only process changed files.

### Lifecycle

**On worktree create (sub-project 2 will wire this into TUI):**

1. Show checkboxes for all integrations. Per-repo `.sentei.yaml` `integrations_enabled` controls which are pre-checked.
2. For each enabled integration, resolve dependency chain:
   a. Check integration `Detect` → if found, skip to step 3.
   b. Walk `Dependencies` in order → check each `Detect`.
   c. Missing dependency → prompt "Install {dep.Name}?" → run `dep.Install`.
   d. All deps satisfied → run integration `Install`.
   e. Re-check integration `Detect` → if still fails, error: "Installed {name} but it's not on PATH. You may need to restart your shell."
3. If `FirstRunNote` is set and this is the first install, display note in TUI.
4. Run `Setup.Command` in the appropriate working directory.
5. Append `GitignoreEntries` to worktree `.gitignore` if not already present.

**On worktree remove (extends existing removal flow):**

1. Before `git worktree remove`, for each integration:
   a. Check if `Teardown.Dirs` exist in the worktree.
   b. If yes and `Teardown.Command` is set, run it from the worktree dir.
   c. If command fails or is unavailable, fall back to deleting `Teardown.Dirs`.
2. This is automatic — no user toggle needed for cleanup.

### Error Handling

- `Detect` command fails → expected state (not installed), offer install.
- `Install` command fails → surface the full error output. Do not retry.
- Dependency `Install` succeeds but `Detect` still fails → error with PATH guidance.
- `Setup` command fails → log error, continue with remaining integrations. Report in summary.
- `Teardown` command fails → fall back to directory deletion. Log warning.

### Auditability

`sentei integrations` CLI command: lists all integrations with name, description, URL, install status (installed/not installed), and dependency status. Formatted table output to stdout.

---

## CLI Commands

Two new non-TUI subcommands for auditability:

### `sentei ecosystems`

Lists all registered ecosystems. Output format:

```
Ecosystems (16 registered)

  NAME      DETECT FILES          INSTALL           SOURCE     STATUS
  pnpm      pnpm-lock.yaml        pnpm install      embedded   enabled
  yarn      yarn.lock             yarn install      embedded   enabled
  npm       package-lock.json     npm install       embedded   enabled
  ...
  custom    my-lock.json          my-install        per-repo   enabled
```

### `sentei integrations`

Lists all integrations with install status. Output format:

```
Integrations (2 registered)

  NAME                  STATUS      DESCRIPTION
  code-review-graph     installed   Build code graph for AI-assisted code review
                                    https://github.com/tirth8205/code-review-graph
  cocoindex-code        not found   Semantic code search index
                                    https://github.com/cocoindex-io/cocoindex-code
```

---

## Testing Strategy

### Unit Tests

**Config loading/merging (`config/config_test.go`):**
- Embedded-only loading (no user config files)
- Embedded + global override (field-level merge)
- Embedded + global + per-repo (three-layer merge)
- Per-repo replaces scalar lists (protected_branches)
- Ecosystem override by name (change install command)
- Ecosystem disable (`enabled: false`)
- New ecosystem added via user config (appended)
- Malformed YAML → error with file path
- Unknown fields → no error (forward compatibility)
- Validation: missing ecosystem name → error
- Validation: unknown integration name in `integrations_enabled` → warning

**Ecosystem detection (`ecosystem/ecosystem_test.go`):**
- Single ecosystem detected (one lock file present)
- Multiple ecosystems detected (go.mod + package.json)
- Priority ordering (pnpm-lock.yaml wins over package-lock.json)
- No ecosystem detected → empty slice
- Disabled ecosystem skipped
- Detection with glob patterns (e.g., `*.sln`)

**Workspace detection (`ecosystem/workspace_test.go`):**
- pnpm workspace (parse pnpm-workspace.yaml)
- npm/yarn workspace (parse package.json workspaces field)
- Cargo workspace (parse Cargo.toml workspace.members)
- Go workspace (parse go.work use directives)
- Glob resolution in workspace configs
- Malformed workspace config → warning, treat as non-workspace
- Workspace config exists but resolves to no directories

**Integration registry (`integration/integration_test.go`):**
- `All()` returns all registered integrations
- `Get(name)` returns correct integration or nil
- Each integration has all required fields populated
- Dependency chain is well-formed (no circular deps)
- Detect/Install/Setup/Teardown specs are non-empty

### E2E Tests

**Config discovery (`config/config_e2e_test.go`):**
- Create real directory structure with XDG_CONFIG_HOME, global config, bare repo with .sentei.yaml
- Invoke `LoadConfig`, verify merged result matches expected
- Test config discovery from inside a worktree (resolves to bare repo root)

**Ecosystem detection against real files (`ecosystem/ecosystem_e2e_test.go`):**
- Create temp dirs mimicking real project structures:
  - Go project with `go.mod`
  - Node monorepo with `pnpm-workspace.yaml` + multiple `package.json`
  - Multi-language repo with `go.mod` + `package.json`
- Run full detection pipeline, verify correct ecosystems and workspace directories

**CLI audit commands:**
- Run `sentei ecosystems` as subprocess, verify output contains expected entries
- Run `sentei integrations` as subprocess, verify output format and content

### Test Fixtures

- Embedded YAML fixtures for config merge tests (valid, override, malformed)
- Temp directory builders for ecosystem detection (reusable helpers that create realistic project layouts per ecosystem)
- All tests use `t.TempDir()` for automatic cleanup

---

## Sub-project Boundaries

This spec covers only the data/foundation layer:
- `internal/config/` — config loading and merging
- `internal/ecosystem/` — ecosystem registry and detection
- `internal/integration/` — integration registry and definitions
- CLI subcommands: `sentei ecosystems`, `sentei integrations`

**Not in scope (sub-project 2):** TUI views for worktree creation, plugin toggle UI, dependency install UI, ecosystem install execution, integration setup/teardown execution.

**Not in scope (sub-project 3):** Repo creation, clone-as-bare, migrate-to-bare, GitHub publish.
