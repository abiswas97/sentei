# Selection, Danger, and Focus Weight

## Why

Audit cross-flow findings: selection salience relied on a lone `>` marker (weak at the 30-row scale on both palettes); the destructive `y delete` hint was the dimmest element on the confirm screen; the confirm screen named a detached worktree differently from the list (`detached-head` vs `06f5232`) and mixed prose with badges; `enter delete` was advertised at 0 selected; the filter prompt broke the chrome margin; `[x]` active and `[+]` staged shared one green; staging shifted the layout; input focus was thin.

## What Changes

- Cursor marker becomes `▸` with the selected row's text carrying the accent (menu, integrations); cursor rows in the list stay bold-emphasis.
- `worktreeLabel` becomes the canonical name everywhere (detached = short hash, matching the list); the confirm screen uses badge vocabulary (`[ok] clean`).
- `viewFooterDanger`: the first (destructive) hint renders in the warning style; the confirm screen uses it.
- `enter delete` hides at 0 selected; the filter prompt keeps the 2-space margin; `[x]` goes neutral (green reserved for `[+]`); the pending-count line is always reserved; focused input labels carry the accent.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tui-design-system`: gains the selection/danger weight requirement.

## Impact
helpers.go, list.go, confirm.go, chrome.go, keys.go, menu.go, integration_list.go, three input views.
