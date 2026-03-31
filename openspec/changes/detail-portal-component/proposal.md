## Why

sentei needs a way to show detailed, scrollable content on demand — cleanup details (which branches would be deleted and why), contextual help for the current view, and eventually worktree detail views. Currently there's no overlay or modal capability. The `bubbletea-overlay` library provides compositing, and `bubbles/viewport` provides scrolling. Combining these into a reusable `DetailPortal` component gives every view access to progressive disclosure: summary by default, `?` for details.

## What Changes

- Add `bubbletea-overlay` as a dependency for overlay compositing
- Create a `DetailPortal` shared component: overlay + viewport with consistent chrome (title, scroll indicator, dismiss key)
- Implement a global help overlay as the first consumer — accessible via `F1` from any view, showing contextual key bindings and descriptions for the current view
- Wire `?` (contextual details) and `F1` (global help) key handling into the main Model update loop

## Capabilities

### New Capabilities
- `tui-portal`: Reusable `DetailPortal` component — overlay compositing with scrollable viewport, consistent chrome (title, scroll hints, dismiss key), and standard show/hide lifecycle
- `tui-help`: Global help overlay accessible via `F1` from any view — displays key bindings and descriptions contextual to the active view

### Modified Capabilities
_None — this adds new capabilities without changing existing behavior._

## Impact

- `go.mod` — new dependency: `github.com/rmhubbert/bubbletea-overlay`
- `internal/tui/portal.go` — NEW: `DetailPortal` component
- `internal/tui/help.go` — NEW: help overlay content builder
- `internal/tui/model.go` — MODIFY: add portal state, wire `?` and `F1` key handling
- `internal/tui/keys.go` — already has `keyDetails` and `keyGlobalHelp` from Change 1
- `.impeccable.md` — expand Component Patterns with portal/overlay documentation
