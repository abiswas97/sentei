# Portal Absorbs Info Card: Design

## Context

The card renders one integration at a time (header with page indicator, wrapped description, dependency status rows, URL, nav line) inside `styleInfoCard`, composited via `lipgloss.Place(m.width, m.height+6, ...)` in two views. The portal already provides sizing, scrolling, compositing, key interception, and chrome. `detailContent()` is the established `?` routing: views returning non-empty content get the portal automatically; the integration views currently return empty and handle `?` themselves.

## Goals / Non-Goals

**Goals**: one overlay system; integration details reachable via the standard `?` path on both integration views; all card information preserved; magic sizing numbers gone.

**Non-Goals**: changing what information is shown per integration; portal key extensions (the page is read-only and scrollable, no carousel).

## Decisions

### D1: One scrollable page for all integrations, not a ported carousel
The portal intercepts all keys by design; teaching it per-view keys (h/l) would breach its "read-only scrollable" contract for one consumer. With 2-5 integrations, a single page with one section per integration is faster to scan than paging and matches the portal's documented purpose (progressive disclosure of read-only content). The page-indicator (`2 / 5`) disappears; section headers carry the names.

### D2: detailContent is the only entry point
`detailContent()` gains cases for `integrationListView` and `migrateIntegrationsView` returning the rendered page; the in-view `?` handlers and the model.go fall-through comment for them are removed. `?` toggling, esc dismissal, F1 switching all come free from the portal.

### D3: Section rendering ports the card's dependency-status logic verbatim
`renderIntegrationsDetail()` (new, in integration_list.go) iterates all integrations producing: bold name, description (wrapped by the portal width), dependency rows with `●` installed / `·` will-be-installed, dim URL, blank line between sections. The `depStatus` map usage is unchanged.

### D4: View sizing is global
Investigation found the `msg.Height-6` handler duplicated byte-identically in 23 view files, not two. The root-cause fix: the global `Update` (which already intercepts `WindowSizeMsg` for the portal) assigns `m.width`/`m.height` once, with the budget as the `viewChromeRows` constant, and every per-view handler is deleted. The `m.height+6` literals die with the Place calls, completing the magic-6 cleanup.

### D5: Portal scrolls vertically only
The bubbles viewport ships h/l horizontal-scroll defaults that leak through the portal whenever content is wider than the viewport (the live verification caught this). Portal content is wrapped/truncated to the portal width at build time, and the viewport horizontal bindings are disabled: the portal advertises and provides vertical scrolling only.

## Risks / Trade-offs

- [Carousel muscle memory] → h/l did nothing anywhere else; scroll keys are already portal-standard. Decision log records the UX change.
- [Long dependency lists overflow] → portal scrolls; that is its job.

## Migration Plan

Single PR `refactor/portal-absorb` (refactor type: information unchanged, one interaction simplified). Playground: `?` on integrations list opens portal with all sections, esc restores, h/l do nothing. Rollback: revert merge commit.
