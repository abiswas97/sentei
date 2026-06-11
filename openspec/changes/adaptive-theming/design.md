# Adaptive Theming: Design

## Context

The palette is 10 `lipgloss.Color` package vars (PR #48); ~30 `styleX` package vars derive from them at package init; every view references the `styleX` vars directly. lipgloss v2 removed `AdaptiveColor`: the canonical v2 pattern is to learn darkness from `tea.BackgroundColorMsg` and select colors explicitly. Bubble Tea delivers the message before the first render when the terminal answers, and never delivers it otherwise.

## Goals / Non-Goals

**Goals**: light-terminal legibility with the approved mapping; dark default preserving byte-identical output on non-answering terminals; zero changes to view code.

**Non-Goals**: user-configurable themes; truecolor palettes; restyling any component; threading a theme object through views (see D2).

## Decisions

### D1: Palette becomes data: one struct, two declared values
A `palette` struct with the 10 token fields, and two package-level values, `darkPalette` and `lightPalette`, holding the approved ANSI codes. This makes the theme literally data (sacred tenet 2): adding a token means one struct field and two table entries; nothing else knows how many palettes exist. Alternative considered: parallel `colorXLight` vars next to each token; rejected as 10 more loose globals with no structural pairing.

### D2: Styles rebuild via one constructor; the styleX vars stay the views' API
A single `applyPalette(p palette)` assigns the color tokens and rebuilds every derived `styleX` var; package init calls `applyPalette(darkPalette)`. On `BackgroundColorMsg{IsDark: false}`, Update calls `applyPalette(lightPalette)`. The Elm loop is single-goroutine, the message arrives before first render, and tests that never send the message keep the dark default, so mutable package state is safe here.

Alternative considered (the cleaner-on-paper abstraction): a `Theme` struct carrying all styles, dependency-injected through `Model` into every render function. Rejected for now: it touches every view file (~30) and every render-helper signature for zero behavioral gain, violating small-blast-radius for purity's sake. The honest middle ground is centralizing CONSTRUCTION (one palette struct, one constructor, data tables) while keeping the existing reference API. If a third theme dimension ever appears (user themes, truecolor), that is the Rule-of-Three moment to revisit Theme-on-Model.

### D3: Detection request is additive in Init
`Init()` returns `tea.Batch(tea.RequestBackgroundColor, <existing cmd>)`. Handling lives at the top of `Update` next to the other global messages (`WindowSizeMsg`), before view routing, and is a no-op when `IsDark()` is true.

### D4: Light verification uses the throwaway-test technique
Background detection cannot be faked through tmux reliably. Unit tests assert palette switching and message handling; the light theme's visual check uses an in-package throwaway test that applies `lightPalette`, t.Logs `stripANSI`-free renders of menu/list/portal for eyeball inspection, then is deleted (established repo technique). The playground tmux pass on a dark terminal proves no-regression. A permanent test asserts `applyPalette(darkPalette)` restores startup state so the throwaway can't poison test order; the switching test itself must restore the dark palette (package state is shared across the test binary).

## Risks / Trade-offs

- [Package-level style mutation surprises a future parallel test] → switching tests restore dark in a defer; the only mutation sites are init and the Update handler.
- [A terminal answers slowly, after first paint] → Bubble Tea repaints on the message; one dark-then-light flicker on such terminals, same as every v2 Charm app.
- [Light values look wrong in practice] → they are two data lines to tune; the .impeccable.md table is the contract.

## Migration Plan

Single PR `feat/adaptive-theming` (the release the v2 migration rides into): styles.go restructure, Init/Update wiring, tests, spec delta, .impeccable.md table + theme line. Rollback: revert the merge commit.
