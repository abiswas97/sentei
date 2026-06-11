# Proposal: confirm-columns

## Why

The confirm-deletion screen lists worktrees as `name [badge] note` inline, so badges land wherever each name ends — ragged and slow to scan (user-reported from a real run). The user chose a badge-first gutter from the layout lab.

## What Changes

- Confirm rows become columnar: status badge in a fixed left gutter, names aligned after it (truncated to a stable width), risk notes trailing only on at-risk rows. Clean rows drop the redundant "clean" word — the green `[ok]` carries it.
- The one-vertical-sweep rule: the gutter answers "anything risky?" before any name is read, mirroring the worktree list's status column.

## Capabilities

### Modified

- `tui-design-system`: confirm screen layout.

## Impact

- `internal/tui/confirm.go` row loop; confirm/weight tests.
