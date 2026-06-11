# Canvas Overlay Compositor

## Why

The audit caught a stray list-cursor `>` glyph bleeding through the portal: an artifact of sentei's hand-rolled ANSI splicer (`compositeOverlay`). lipgloss v2 ships first-class compositing (`Canvas` + `Compositor`/`Layer`) built on ultraviolet's cell buffer, handling wide runes and SGR state by construction, and its layer IDs later enable mouse hit-testing.

## What Changes

- `compositeOverlay` keeps its signature but its body becomes a lipgloss `Canvas` composing a background layer and a centered, z-raised foreground layer via `Compositor`.
- The bespoke splice logic (manual ANSI truncation, SGR reset injection) is deleted.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tui-portal`: the compositing requirement notes library-grade cell compositing (no behavioral change to the contract scenarios).

## Impact
internal/tui/overlay.go only; existing overlay/portal tests define the contract and pass unchanged.
