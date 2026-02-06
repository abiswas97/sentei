## Context

The worktree list is currently rendered directly from `m.worktrees` — a flat slice populated once at startup. Navigation (`m.cursor`), selection (`m.selected`), and rendering all index into this slice by position. Adding sorting and filtering requires an indirection layer so the display order can change without invalidating selections or mutating the source data.

The TUI uses Bubble Tea's Elm architecture: state lives in `Model`, input flows through `Update`, and `View` is a pure render function. Any new state (sort config, filter text, visible indices) must follow this pattern.

## Goals / Non-Goals

**Goals:**
- Sort worktrees by age (default), branch name, or path with reversible direction
- Filter worktrees by branch name via inline text input
- Maintain selection integrity across sort/filter changes
- Keep the implementation minimal — no new packages, no architectural rewrites

**Non-Goals:**
- Fuzzy matching (simple case-insensitive substring is sufficient for branch names)
- Multi-field sort (single field at a time)
- Filter by fields other than branch name (age threshold filtering is PRD F9, non-interactive mode)
- Persistent sort/filter preferences across sessions (PRD F10, configuration file)

## Decisions

### D1: Indirection via `visibleIndices []int`

Add a `visibleIndices []int` field to `Model` that maps display positions to indices in the original `m.worktrees` slice. All cursor movement, rendering, and selection operations use this mapping.

**Why**: Avoids re-sorting the source slice (which would break index-based selection) and avoids copying worktrees on every filter change. A single `reindex()` method recomputes `visibleIndices` whenever sort or filter state changes.

**Alternative considered**: Sorting `m.worktrees` in place and using a stable identity map. Rejected because it complicates the selection model and requires more bookkeeping.

### D2: Path-based selection (`map[string]bool`)

Change `m.selected` from `map[int]bool` to `map[string]bool`, keyed by worktree path. Worktree paths are unique and stable regardless of sort/filter order.

**Why**: Index-based selection breaks when the display order changes. Path-based selection is the simplest stable key — no new ID generation needed.

**Impact**: `selectedWorktrees()`, `viewList()`, `updateList()` (toggle/select-all), and `viewConfirm()` all reference `m.selected` and must switch to path keys.

### D3: `bubbles/textinput` for filter input

Use `textinput.Model` from the existing bubbles dependency for the filter bar. When filter mode is active, the textinput receives key events; normal list key bindings are suppressed.

**Why**: `textinput` handles cursor movement, backspace, paste, and other editing natively. Building a custom text input would duplicate this for no benefit. The dependency already exists in `go.sum`.

**How it integrates**: A boolean `filterActive` controls whether `updateList` routes `tea.KeyMsg` to the textinput or to normal list bindings. When the textinput value changes, `reindex()` is called to update `visibleIndices`.

### D4: Sort state as enum + direction bool

```
type SortField int
const (
    SortByAge SortField = iota
    SortByBranch
)
```

Plus `sortAscending bool`. The `s` key cycles `SortField`; `S` (shift+s) toggles `sortAscending`.

**Why**: Simple, explicit. Two sort fields covers the practical use cases — age for finding stale worktrees, branch for locating by name. Path sorting was removed as it duplicates branch without adding value.

**Default**: `SortByAge` ascending (oldest first), since the primary use case is finding stale worktrees.

### D5: `reindex()` as the single recomputation point

A single `reindex()` method on `Model`:
1. Filters `m.worktrees` by current filter text (case-insensitive substring on branch name)
2. Sorts the filtered indices by current sort field and direction
3. Stores result in `m.visibleIndices`
4. Clamps `m.cursor` and `m.offset` to valid range

Called from: `Init` (initial sort), sort key handlers, filter text changes, and window resize (no-op but safe).

### D6: Key binding strategy for filter mode

- `/` enters filter mode (focuses textinput, shows filter bar)
- While filtering: all keys go to textinput except `esc` (exit filter, clear text) and `enter` (accept filter, return to list navigation)
- When filter is applied (text non-empty, mode inactive): status bar shows filter text and match count
- `esc` from list view with an active filter clears the filter first, second `esc` quits

**Why**: Matches Vim `/` search and the bubbles/list convention. Context-dependent `esc` avoids needing a separate "clear filter" key.

## Risks / Trade-offs

**[Risk] Selection confusion when filter hides selected items** → Mitigation: Status bar always shows total selected count (including hidden). Confirmation dialog shows all selected worktrees regardless of current filter.

**[Risk] `esc` key overloaded (quit vs clear filter vs exit filter mode)** → Mitigation: Clear precedence — filter mode active: exit filter mode. Filter applied: clear filter. No filter: quit. Each `esc` press does exactly one thing.

**[Risk] `reindex()` performance on large worktree lists** → Mitigation: Negligible. Even 100 worktrees is a trivial sort. No caching needed.

**[Trade-off] Substring match vs fuzzy match** → Substring is simpler, predictable, and sufficient for branch names which users know. Fuzzy can be added later if demanded.
