# Proposal: p3-polish

## Why

Wave 3d closes the audit's P3 list: small frictions individually, a polish debt collectively.

## What Changes

- **Sort arrow describes what the eye sees**: the Age column displays elapsed time, which runs opposite to the underlying commit-date sort; its arrow now flips so ▼ means descending ages on screen.
- **Portal scroll hints only when scrollable**: fitting content drops `j/k scroll` and `↓ more`.
- **Options footer gains `j/k navigate`.**
- **Tab lands at the end of prefilled inputs** (six field-switch sites).
- **Cursor and vocabulary strays**: option views still used `> ` (now `▸ `), and the GitHub auth status carried a literal `●` (now `✦`).

## Capabilities

### Modified

- `tui-design-system`: the above presentation rules.

## Impact

- list.go, portal.go, keys.go, three option views, three input views; golden regenerated for the intentional arrow flip.
