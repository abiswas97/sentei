# Portal Absorbs the Integration Info Card

## Why

The spec has promised since the DetailPortal landed that the integration info card is "the one true overlay until the DetailPortal component lands and absorbs it." The portal landed two releases ago; the card survives as a second bordered-overlay implementation with its own sizing logic and magic numbers (`m.height+6` at two `lipgloss.Place` sites, clamped `innerWidth` math). Two overlay systems means two places for compositing bugs (the lipgloss v2 `Width` regression had to be fixed in both).

## What Changes

- The `?` key on the integration list and migrate-integrations views opens the standard DetailPortal with an "Integration Details" page listing ALL integrations (name, description, dependencies with install status, URL), scrollable like every other portal page.
- **REMOVED**: the carousel interaction (h/l paging, one integration per page). All integrations are visible on one scrollable page, which serves the 2-5 integration reality better than paging. The `Left`/`Right` bindings become dead and are deleted.
- Deleted wholesale: `showInfo`/`infoCursor` state, `updateIntegrationInfo`, `renderIntegrationInfo`, `styleInfoCard`, both `lipgloss.Place` call sites with their `m.height+6` magic, the wheel-to-carousel mapping.
- Sizing is hoisted into the global `Update`: 23 byte-identical per-view `WindowSizeMsg` handlers are deleted and the chrome budget becomes the `viewChromeRows` constant in `constants.go`.
- `.impeccable.md` Dialogs section becomes literally true: the portal is the only overlay.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-portal`: gains the integration-details page requirement (the card had no spec coverage; the portal spec now owns this surface).

## Impact

- **Code**: `integration_list.go` (large deletion + portal content builder), `migrate_integrations.go` (overlay block deletion), `model.go` (state fields, detailContent routing), `keys.go` (dead bindings), `menu.go`/`list.go` (chrome-rows constant), `constants.go`.
- **Tests**: carousel tests replaced by portal-content tests; existing integration list tests unaffected; wheel carousel tests replaced by list-cursor-only assertions.
- **UX change**: h/l paging is gone; scrolling replaces it. Recorded in the decision log.
