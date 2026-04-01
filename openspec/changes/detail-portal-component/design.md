## Context

Change 1 (`ui-chrome-unification`) established shared chrome helpers, contextual `?` and global `F1` key bindings, and the `.impeccable.md` design system. This change builds on that foundation to add overlay capability — a scrollable detail view composited on top of the current view.

The `bubbletea-overlay` library handles the compositing (foreground rendered on top of background at a specified position). The `bubbles/viewport` component handles scrollable content. Combining these gives us a `DetailPortal` that any view can open with arbitrary content.

## Goals / Non-Goals

**Goals:**
- Reusable `DetailPortal` component usable by any view
- Help overlay as first consumer, proving the component works
- Consistent chrome: title bar, scroll indicators, dismiss key
- Scrollable content for lists that exceed terminal height

**Non-Goals:**
- Cleanup-specific detail views (Change 3)
- Interactive content inside the portal (forms, selections) — portal is read-only
- Nested portals

## Decisions

### D1: DetailPortal as a sub-model on Model, not a standalone tea.Model

**Decision**: `DetailPortal` is a struct embedded in `Model` with its own `Update` and `View` methods, but not a standalone `tea.Model`. The main `Model.Update` delegates to it when active, and `Model.View` composites it over the current view when visible.

**Alternatives considered**:
- *Standalone tea.Model with bubbletea-overlay.New()*: Would require a separate `tea.Program` or complex model composition. Doesn't integrate cleanly with the existing single-Model Elm architecture.
- *Pure function like chrome helpers*: Portal has state (scroll position, visibility), so it can't be a pure function.

**Why sub-model**: Keeps the single `Model` as the root of truth. Portal state (visible, content, scroll position) lives on `Model`. When visible, `Update` routes keys to the portal first. `View` uses `bubbletea-overlay` to composite the portal's rendered content over the background view.

### D2: Content provided as pre-rendered string

**Decision**: The portal takes a `string` (pre-rendered content) and wraps it in a scrollable viewport. The caller is responsible for rendering the content with appropriate styles.

**Alternatives considered**:
- *Structured content model (sections, items)*: Would require a content DSL. Over-engineered for what's essentially "show this text in a scrollable box."

**Why string**: Simple, flexible. Each consumer renders its content however it wants (using lipgloss, etc.) and passes the result to the portal. The portal only cares about scrolling and chrome.

### D3: Help content generated per-view

**Decision**: Each view state has a `helpContent() string` method that returns the help text for that view. The help overlay calls this when `F1` is pressed, providing contextual help.

**Alternatives considered**:
- *Static help text*: Would be generic and not contextual.
- *Help registry mapping view → content*: Extra indirection without benefit.

**Why per-view method**: Direct, no registry. Each view knows its own keys and behavior. The help text is rendered using consistent formatting (key tables, descriptions).

## Risks / Trade-offs

**[Risk] `bubbletea-overlay` is a third-party dependency** → Mitigated: the library is small (overlay compositing only), well-documented, and the overlay rendering could be replaced with lipgloss `Place` if needed. Low lock-in risk.

**[Trade-off] Portal intercepts all keys when visible** → Accepted. When the portal is open, only scroll keys (j/k, up/down), page navigation (g/G), and dismiss (esc, ?) work. All other keys are swallowed. This prevents accidental actions while reading details.

**[Risk] Viewport scroll state persists across open/close** → Mitigated by resetting scroll position to top on each open.
