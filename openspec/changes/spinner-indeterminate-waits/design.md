# Spinner: Design

## Context

Two indeterminate waits exist: `cleanupScan == nil` in the cleanup preview (scan running) and the bare-repo menu before `worktreeContextMsg` arrives (`Remove worktrees` item disabled with a `loading…` hint). bubbles/spinner animates via a self-scheduling `Tick` command producing `spinner.TickMsg`.

## Goals / Non-Goals

**Goals**: visible motion during both waits; ticks only while a wait is visible; zero change to determinate progress and the indicator vocabulary.

**Non-Goals**: spinners anywhere else; configurable spinner styles.

## Decisions

### D1: One shared spinner on Model
Both waits are the same concept (indeterminate activity), never visible simultaneously with conflicting needs; one `spinner.Model` field serves both. MiniDot style, accent foreground (wired through `applyPalette` so theming holds).

### D2: Tick lifecycle gated by visibility
`spinner.TickMsg` is handled at the top of `Update`: when an indeterminate wait is visible (`m.indeterminateWaitActive()`), update the spinner and return the next tick; otherwise swallow the message, ending the cycle. The tick starts when a wait begins: entering the cleanup flow (scan start) and menu init/load kick (bare-repo context before worktrees arrive). Restarting on re-entry is handled by issuing `m.spinner.Tick` alongside the existing state-entry commands.

### D3: Menu loading is a flag, not a string match
`menuItem` gains `loading bool`; the builder sets it on the worktrees item until context arrives; `viewMenu` renders `<spinner> loading…` for loading items. No render-site string matching.

## Risks / Trade-offs

- [Tick keeps firing after the wait ends] → the gate swallows orphan TickMsgs; a test asserts no cmd is returned once the wait is over.
- [teatest timing assumptions] → assertions target stable text ("Scanning repository"), not the animated glyph.

## Migration Plan

Single PR `feat/spinner`. Playground: enter cleanup preview and observe animation during the scan hold; menu loading state. Rollback: revert merge commit.
