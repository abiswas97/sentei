# Design: progress-motion

## Breathing dot

A second `spinner.Model` (`m.breath`) with custom frames:

```go
spinner.Spinner{Frames: []string{"·", "∙", "●", "∙"}, FPS: time.Second / 3}
```

Four frames at 3 fps gives a ~1.3s calm heartbeat. All frames are single-cell, so the
`●`/`·` status column stays aligned. The styled current frame (`styleIndicatorActive`)
is injected into rendering the same way the bar and elapsed readout already are:
`ProgressLayout` gains an `ActiveGlyph string` field; pure constructions leave it empty
and renderers fall back to the static `●`-family midpoint `∙` so layout functions stay
pure and table-testable. `renderProgressLayout` injects the live frame.

Tick gating reuses the existing pattern: `spinner.TickMsg` for the breath spinner is
forwarded only while `determinateProgressActive()` (or the cleanup result's running
state) holds, exactly as `indeterminateWaitActive` gates `m.spin`. The two spinners
have distinct tick IDs, so the gates do not interfere.

Consumers of `indicatorActive` and their replacements:

| Site | Today | After |
| --- | --- | --- |
| Phase line (`renderPhase`) | static `◐` | breath frame |
| Step line (`renderPhase`) | static `◐` | breath frame |
| Stat line (`viewStatLine`) | static `◐` | breath frame (counts live steps) |
| Cleanup result running line | static `◐` | breath frame |

`indicatorActive = "◐"` is deleted; the constant becomes the frame set. Tests assert
on the fallback `∙` or strip frames via the spinner's initial frame.

## Responsive bar width

`progressBarWidth = 20` becomes the floor, not the width:

```go
barWidth = max(minBarWidth, contentWidth - metaWidth)
```

where `contentWidth = l.Width - 2` (left margin) and `metaWidth` is the rendered width
of everything right of the bar (`"  55%"` plus `"  4s"` elapsed when present), measured
with `lipgloss.Width` at render time. The animated model's width is set on
WindowSizeMsg in the global Update (same hoist as view sizing); the static fallback
`renderProgressBar` takes the width as a parameter. The percentage label keeps the
existing truth contract (follows displayed fill).

## Gradient fill

Two new palette tokens (hex so blends interpolate smoothly; the v2 renderer
downsamples for non-truecolor terminals):

| token | dark | light |
| --- | --- | --- |
| `barStart` | `#5f5fd7` (accent 62) | `#5f00d7` (accent 56) |
| `barEnd` | `#ff87d7` (selected 212) | `#d75f87` (selected 168) |

`newOverallBar` adds `progress.WithColors(colorBarStart, colorBarEnd)` and
`progress.WithScaled(true)` so the blend always spans exactly the filled portion.
Fill characters stay `█`/`░`. The static fallback bar keeps solid accent: it renders
only in pure/static contexts (tests, pre-first-frame), and the spec records that.

## What does not change

Spring physics (harmonica via bubbles/progress), the percentage-follows-fill contract,
elapsed readout, completion holds, `– skipped` semantics, and the OSC 9;4 mirror all
stay as shipped.

## Revision: smoothing (post-#73 GIF review)

User review found the breath staccato (4 frames at 3fps = a visible jump every
333ms) and the bar snappy rather than smooth, and flagged the two animated
vocabularies (braille waits vs dot rows). The dot's smoothness is capped by
glyph resolution — Unicode offers ~4 usable dot sizes — so rather than tune it,
the active indicator standardizes on the heavy braille dot, single-cell frames
at 10fps:

```go
spinner.Spinner{Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}, FPS: time.Second / 10}
```

(The bubbles builtin `Dot` ships two-cell frames with a trailing space; the
custom set strips it so the status column stays aligned.)

With one vocabulary the two spinner instances merge into one: `m.spin` carries
every working surface, gated by `spinnerActive()` (indeterminate waits ∪
determinate progress ∪ cleanup running line). The dispatch wrapper starts the
tick chain on any transition into a working state; the explicit tick in the
cleanup-preview entry is removed so no path double-starts the chain (two
chains advance frames at double speed — the spinner dedupes by ID, not by
chain).

The bar spring drops from frequency 30 to 8 (damping 1): fills glide over
most of a second instead of snapping in ~0.2s, still settling within the
1.5s completion hold.

Transitions stay instant cuts by decision: motion belongs to state-driven
elements (spring, spinner, holds), never to keypress navigation. Slide
transitions via Canvas+harmonica were considered and declined — every
navigation would pay latency in a productivity tool.
