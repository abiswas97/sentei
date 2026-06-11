# Proposal: milestone-whisper

## Why

Wave 3b of the joy redesign: sentei should quietly mark the long arc of use. A lifetime counter exists nowhere; milestones pass silently.

## What Changes

- `State` gains `LifetimeRemoved`, incremented by each TUI removal run's success count (atomic save, degrades silently on error — garnish never alarms).
- When a run crosses a power of ten (10, 100, 1000, …), the removal summary whispers one dim line: `that was your 100th worktree, pruned`.
- The milestone message is handled globally: the hold may already have advanced the view when the recording lands.

## Capabilities

### Modified

- `tui-design-system`: summary whisper rule.

## Impact

- internal/state (one field), internal/tui/milestone.go (new), progress.go wiring, summary.go render, copy.go line.
