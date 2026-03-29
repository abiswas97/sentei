# Integration Selection & Management

## Overview

Add repo-level integration management to sentei. Integrations (dev tools like code-review-graph and cocoindex-code) are enabled/disabled per bare repo, not per branch. Main's worktree is the source of truth — new worktrees inherit whatever is active on main.

## Data Model

### Persistent State

Enabled integrations stored in `.bare/sentei.json`:

```json
{
  "integrations": ["code-review-graph", "cocoindex-code"]
}
```

- Lives inside `.bare/` — invisible to worktrees
- Only stores the enabled list
- Created on first explicit integration management or during migration onboarding

### Source of Truth Hierarchy

1. `.bare/sentei.json` — what *should* be active
2. Disk scan of main worktree — what *is* active
3. Config says enabled but main missing artifacts → "enabled (not set up)", offer to run setup
4. No config file (fresh/pre-existing bare repo) → fall back to disk detection

### Detection

Check for artifact directories on main's worktree, using `Integration.GitignoreEntries` as detection paths:

- `{mainWorktreePath}/.code-review-graph/` → crg active
- `{mainWorktreePath}/.cocoindex_code/` → ccc active

Config-driven — no hardcoded directory checks.

## Menu Changes

New item "Manage integrations" in bare repo menu, 2nd position:

```
> Create new worktree
  Manage integrations
  Remove worktrees        3 available
  Cleanup                 safe mode
```

## Integration Management View

Simple list of all known integrations with toggle checkboxes:

```
  sentei ─ Integrations

  myrepo (bare)

  ─────────────────────────────────────────

> [x] code-review-graph
       Build code graph for AI-assisted code review

  [ ] cocoindex-code
       Semantic code search index

  ─────────────────────────────────────────

  [x] active  [ ] inactive  [+] adding  [-] removing

  w/s navigate · space toggle · enter apply
  ? info · esc back
```

### Staged Changes

Toggling stages changes visually without applying:

- `[+]` staged to add (green)
- `[-]` staged to remove (yellow/red)
- Footer shows "{N} changes pending" and adds `esc discard`

### Info Dialog (Carousel)

Pressing `?` opens a rounded-border dialog showing the focused integration's details:

```
  ╭─────────────────────────────────────────╮
  │                                         │
  │  code-review-graph              1 / 2   │
  │                                         │
  │  Build code graph for AI-assisted       │
  │  code review                            │
  │                                         │
  │  github.com/tirth8205/code-review-graph │
  │                                         │
  │  Requires  python3.10+, pipx            │
  │                                         │
  │              ◀ a/← prev · d/→ next ▶    │
  │                          esc to close   │
  ╰─────────────────────────────────────────╯
```

- Content driven by `Integration` struct: `Name`, `Description`, `URL`, `Dependencies[].Name`
- `a/←` and `d/→` cycle through all integrations
- Page indicator `N / total` in top right
- `esc` closes back to the list

### Navigation

- `w/s` or `↑/↓` — move between integrations
- `space` — toggle
- `enter` — apply pending changes
- `?` — open info dialog
- `esc` — discard changes / back to menu

## Apply Progress View

After pressing `enter` with pending changes:

```
  sentei ─ Applying Integration Changes

  ─────────────────────────────────────────

  main
    ✓ Install cocoindex-code
    ⏳ Setup cocoindex-code

  feature/auth
    ✓ Copy index from main
    ⏳ Setup cocoindex-code

  ─────────────────────────────────────────

  ████████████░░░░░░░░ 3/4
```

### Enabling an Integration

1. Set up on main first (detect → deps → install → setup → gitignore)
2. For each existing worktree:
   - **ccc**: Copy `.cocoindex_code/` from main, run `ccc index` (incremental)
   - **crg**: Run `code-review-graph build --repo {path}` (fresh, sub-second)
3. Update `.bare/sentei.json`

### Disabling an Integration

1. For each worktree (including main):
   - Run teardown command if defined (e.g., `ccc reset --all --force`)
   - Remove artifact directories
2. Update `.bare/sentei.json`

### Error Handling

- Errors shown inline with `✗` per step
- Continue with remaining operations on failure
- Return to management view with updated state (re-scan disk)
- Failed add operations leave the integration as "not set up" — user can retry

### UX Parity

- Progress view uses same spinner/checkmark/error styling as create-progress and repo-progress views
- Progress bar at bottom with `done/total` count
- Grouped by worktree for readability

## ccc Optimization

cocoindex-code stores **relative paths** in its index. This enables cross-worktree copying.

### Setup Flow (new worktree or enabling)

1. Ensure main's index is fresh:
   - `git fetch origin main && git merge origin/main` in main's worktree
   - Run `ccc index` in main (incremental, fast if recent)
2. Copy `.cocoindex_code/` from main to target worktree
3. Run `ccc index` in target (incremental — only processes delta, ~1s vs ~15s full)

### Embedding Model

The embedding model is stored globally at `~/.cocoindex_code/` and downloaded once on first use across all projects. Not a per-worktree cost.

## crg Optimization

code-review-graph stores **absolute paths** — copying between worktrees is not possible.

- Always run `code-review-graph build --repo {path}` fresh (sub-second for typical repos)
- For subsequent updates during development: `code-review-graph update --base main`

## Migration Onboarding

After the backup decision screen, a new integration selection screen is inserted before the what-next screen.

### Flow

```
Migration Progress → Backup Decision → Integration Selection → What Next
```

### Integration Selection Screen

```
  sentei ─ Set Up Integrations

  We detected your repo may benefit from
  these dev tools. Select any to enable.

  ─────────────────────────────────────────

> [x] code-review-graph          detected
       Build code graph for AI-assisted code review

  [ ] cocoindex-code
       Semantic code search index

  ─────────────────────────────────────────

  [x] active  [ ] inactive

  w/s navigate · space toggle · enter continue
  ? info · esc skip
```

- Scans migrated main worktree for existing artifacts
- Pre-checks detected integrations, shows "detected" hint
- `esc` skips — user can manage later from the menu
- `enter` runs setup (with progress view), writes `.bare/sentei.json`, then shows what-next
- `?` opens the same info carousel dialog

## Post-Migration What-Next Screen

Updated text — no longer mentions specific integrations:

```
  sentei ─ Migration Complete

  ✓ myrepo ready

    cd /path/to/myrepo/main

  Your repo is ready for worktrees.
  Continue in sentei to create worktrees
  and set up your workspace, or exit
  to your shell.

  ─────────────────────────────────────────

  enter open in sentei · q exit
```

## Create Worktree Changes

### Options View

Integration toggles removed. Replaced with informational line:

```
  sentei ─ Create Worktree

  feature/auth → from main

  ─────────────────────────────────────────

  Setup

> [x] Install dependencies (Node.js)  detected
  [x] Merge default branch           main → feature/auth
  [ ] Copy environment files          .env, .env.local

  ─────────────────────────────────────────

  Integrations from main: crg, ccc

  w/s navigate · space toggle · enter create · esc back
```

- "Integrations from main: crg, ccc" is non-toggleable, informational only
- Read from `.bare/sentei.json` (or disk detection fallback)
- Omitted entirely if no integrations are active

### Setup Flow

During worktree creation, after existing phases (Setup, Dependencies):

1. **ccc**: Copy `.cocoindex_code/` from main → run `ccc index` (incremental)
2. **crg**: Run `code-review-graph build --repo {path}` (fresh)

Progress shown in the existing create-progress view under an "Integrations" phase.

## New View States

- `integrationListView` — the management list with toggles
- `integrationProgressView` — progress view for applying changes
- `migrateIntegrationsView` — integration selection during migration onboarding

## Integration Struct Additions

The existing `Integration` struct needs a new field for detection paths (currently implicit via `GitignoreEntries`). If `GitignoreEntries` continues to serve this purpose, no struct change is needed. If detection logic diverges from gitignore entries in the future, add:

```go
type Integration struct {
    // ... existing fields
    DetectDirs []string // directories to check for presence (defaults to GitignoreEntries)
}
```

For now, reuse `GitignoreEntries` — YAGNI.

## Navigation Scheme

All new views use:
- `w/s` or `↑/↓` for vertical navigation
- `a/d` or `←/→` for horizontal navigation (carousel)
- `space` for toggle
- `enter` for confirm/apply
- `esc` for back/cancel/skip
- `?` for info dialog
- `q` for quit

Existing views retain their current keybindings. The `j/k` bindings in existing views are unchanged.
