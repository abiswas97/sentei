# Terminal Tab Presence

## Why

A worktree-heavy user runs many terminal tabs; sentei is invisible among them. The audit's identity work and the charm-scout both ranked terminal-native presence as the highest payoff-per-line remaining: crush sets the window title; OSC 9;4 progress shows in tab/taskbar even when the window is buried.

## What Changes

- `tea.View.WindowTitle`: `sentei · <repo>` at rest, `sentei · <repo> · removing 2/3` (live verb + counts) during flows, `· scanning` during the cleanup scan.
- `tea.View.ProgressBar` (OSC 9;4): indeterminate during spinner waits, determinate value during progress flows, error state when any phase has failures, absent otherwise.
- The per-view operation names consolidate into one table feeding both the quit trace (`InterruptedFlow`) and the tab verb.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tui-design-system`: gains the terminal-presence requirement.

## Impact
internal/tui/model.go (View fields, op-name table); tests.
