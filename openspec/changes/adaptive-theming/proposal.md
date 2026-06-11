# Adaptive Theming

## Why

`.impeccable.md` has promised "Adaptive — detect terminal background and adjust colors for both dark and light terminals" since the design spec was written, but every palette value is a fixed dark-terminal ANSI code: on light terminals, titles (white 15) and body rows (light gray 252) are nearly invisible. The Charm v2 migration (PR #49) landed `tea.RequestBackgroundColor`, the canonical detection mechanism, making this the cheapest moment to deliver the promise. The light palette mapping was approved by the user on 2026-06-11.

## What Changes

- The palette in `internal/tui/styles.go` becomes a dark/light pair per token; the active set is selected by terminal background detection at startup.
- `Model.Init()` additionally issues `tea.RequestBackgroundColor`; `Model.Update()` handles `tea.BackgroundColorMsg` and applies the light palette when `IsDark()` is false.
- Default is the dark palette: terminals that never answer the background query render exactly as today; all existing tests remain valid unchanged.
- Approved light mapping: accent 62→56, success 42→29, warning 214→166, error 196→160, dim 241→245, emphasis 15→235, body 252→238, selected 212→168, protected 63→26, muted 245→243.
- `.impeccable.md` palette table gains the Light column; the theme line stops being aspirational.
- No layout, copy, or interaction changes of any kind.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-design-system`: the design-system gains an adaptive-palette requirement (background detection, light palette application, dark default).

## Impact

- **Code**: `internal/tui/styles.go` (palette + style construction), `internal/tui/model.go` (Init cmd, BackgroundColorMsg handling). Nothing else: the token layer from PR #48 exists precisely so this change stays two files plus spec/docs.
- **Tests**: unit tests for palette switching and message handling; existing render assertions stay on the dark default. Light rendering verified via the throwaway in-package t.Log technique plus playground on a dark terminal for no-regression.
- **Release**: feat-type commit; this is the release the v2 migration rides into.
