# Charm v2 Migration

## Why

Sentei's UI stack (bubbletea 1.3.10, bubbles 1.0.0, lipgloss 1.1.0) is the previous major generation; Charm shipped stable v2 across the stack (bubbletea/lipgloss v2.0.0, bubbles v2.1.0) under the `charm.land` module path. The session goal is full Charmbracelet alignment, and every queued UI improvement (adaptive theming, bubbles/help, spinner) builds on APIs that v2 reshaped: migrating first means each lands Charm-canonical with no rework. v2 also directly fixes a latent defect: the advertised `ctrl+enter` quick-create binding is unreachable on most terminals under v1.

## What Changes

- **BREAKING (build-level only)**: module path swap to `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2` (plus `lipgloss/v2/table` and teatest v2); Go toolchain 1.24.2 → 1.25 in go.mod and CI.
- `Model.View()` (and the two sub-model `View() string` implementations feeding it) migrate to the v2 declarative `tea.View` struct; `tea.WithAltScreen()` at both `tea.NewProgram` sites becomes `view.AltScreen = true`.
- All `tea.KeyMsg` handling becomes `tea.KeyPressMsg`; the space-key binding in `keys.go` changes from `" "` to `"space"` (v2 `String()` semantics).
- **Keyboard enhancements enabled**: `view.KeyboardEnhancements` requested, making `ctrl+enter` quick-create functional on kitty-protocol terminals. Quick-create gains its first E2E test (kept per usage trace: it is the skip-options creation path).
- **Mouse wheel enabled**: `view.MouseMode = tea.MouseModeCellMotion`; wheel scrolls the worktree list, viewport-backed portal, and integration carousel. No click handling in this change.
- Renderer swap side effects accepted: synchronized output (mode 2026, flicker-free), grapheme clustering (mode 2027), automatic color-profile downsampling.
- No visual redesign: rendered output must be byte-comparable to v1 modulo renderer framing; `.impeccable.md` decision log records the migration.
- Ships as `refactor`-type commits; rides into the next feature release (adaptive theming) rather than forcing its own.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-design-system`: platform requirement moves to Charm v2 (synchronized output, declarative View, keyboard enhancement request); quick-create binding requirement becomes "functional on kitty-protocol terminals" instead of aspirational.
- `tui-list-view`: mouse wheel scrolls the worktree list.
- `tui-portal`: mouse wheel scrolls the portal viewport.

## Impact

- **Code**: all 61 files importing bubbletea (every `internal/tui` view + tests, `main.go`), 25 importing bubbles, 5 importing lipgloss, 3 teatest E2E files. Mechanical, compile-error-driven.
- **Dependencies**: charm.land v2 modules replace github.com/charmbracelet v1 modules; `colorprofile` becomes direct behavior, termenv drops out.
- **CI**: Go version bump in workflow matrix; existing gates (build, lint, test x2 OS, commitlint, codecov/patch) unchanged.
- **Tests**: teatest v2 API differences; key-event construction in unit tests (`tea.KeyMsg{...}` literals) rewritten; new quick-create E2E.
- **Verification**: playground at 80x18 and normal size, dark terminal, before/after eyeball per view; wheel scroll manually verified.
- **Risk**: renderer rewrite (cursed renderer) could change edge-case output (wide glyphs, overlay compositing via `compositeOverlay`); the overlay/`lipgloss.Place` code paths get explicit verification.
