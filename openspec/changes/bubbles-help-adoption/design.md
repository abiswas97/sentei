# Bubbles/help Adoption: Design

## Context

Three sources of key/label truth exist (keys.go WithHelp, footer literals, help.go tables) plus six inline hint strings. `bubbles/help` v2 renders short footers from `[]key.Binding` and is fully styleable, but its `FullHelp() [][]key.Binding` model is unnamed columns, while the tui-help spec requires named, grouped sections in the F1 portal. The chrome contract (tui-chrome spec) fixes the footer's visual format: dim, ` · ` separators, 2-space left pad.

## Goals / Non-Goals

**Goals**: one declaration per view of its key presentation; footers and portal content both derive from it; render sites carry zero key/label strings; list view advertises `?`/`F1`; footer output visually unchanged otherwise.

**Non-Goals**: changing any binding; the `viewFrame` composer (separate refactor); short/full footer toggle (the F1 portal already serves as full help; a second expansion mechanism would duplicate it).

## Decisions

### D1: keySection data in keys.go is the single source; both renderers consume it
`type keySection struct { name string; bindings []key.Binding }`. Each view declares `var <view>Keys = []keySection{...}` in keys.go. The footer flattens the sections' bindings (or a view-specific footer subset, see D3) into `help.ShortHelpView`; the portal renders sections with their names via the existing aligned-table formatter, now iterating bindings instead of hand-written pairs. Alternative considered: implement `help.KeyMap` per view and use `FullHelp` for the portal; rejected because FullHelp loses section names the spec requires, and an adapter type per view is more code than a slice literal.

### D2: Contextual descriptions are declared per view through shared constructors
The same physical key carries different meanings per view ("enter" = delete / continue / confirm). keys.go gains tiny constructors (`withDesc(base key.Binding, desc string) key.Binding`) so per-view sections reuse the canonical key strings from the global `keys` map and override only the description. Key STRINGS stay defined exactly once; descriptions are per-view data declared once each. Render sites reference `<view>Keys` and nothing else.

### D3: Footer = explicit subset, portal = full sections
Footers stay curated (3-5 hints) per the chrome aesthetic; portals show everything. Each view's declaration marks its footer subset by order: `footer []key.Binding` alongside `sections []keySection`, both in the same declaration so they cannot drift apart in separate files. The six stray inline hint sites become `viewFooter(<view>Keys.footer)` calls.

### D4: bubbles/help renders the footer, styled to the chrome contract
One package-level `help.Model` configured in `applyPalette` (styles derive from the active palette: ShortKey/ShortDesc dim, separator ` · `): theming stays consistent when the light palette applies. `viewFooter` wraps `ShortHelpView` and prepends the 2-space pad. Width is set per render so bubbles/help's built-in truncation handles narrow terminals (new behavior at <80 cols, strictly better than overflow).

### D5: Delete, don't deprecate
`KeyHint`, `viewKeyHints`, and help.go's hand-written entries are removed in this change (zero-references rule). The chrome spec requirement is MODIFIED accordingly; output format assertions survive on the new API.

## Risks / Trade-offs

- [bubbles/help default styling diverges from the chrome contract] → styles are explicitly set; chrome tests assert the exact rendered string (` · ` separator, dim codes), not the library's defaults.
- [Footer subset and sections drift] → single declaration site per view; a test walks every view's footer and asserts each footer binding appears in its sections.
- [Some views' footers are conditional (summary enter=menu vs quit)] → conditional footers stay conditional: two declared subsets selected by state, still zero literals at render sites.

## Migration Plan

Single PR `feat/bubbles-help` (feat: list view gains discoverability hints): keys.go sections, chrome viewFooter, help.go portal derivation, ~26 render-site swaps, tests, spec deltas, .impeccable.md View Chrome + Key Mapping sections updated. Playground verification: every view's footer eyeballed at 80x18, F1 portal per view, narrow-width truncation check. Rollback: revert merge commit.
