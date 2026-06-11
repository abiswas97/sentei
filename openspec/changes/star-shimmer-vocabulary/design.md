# Design: star-shimmer-vocabulary

## Motion clock

One message, one counter, one gate:

```go
type motionTickMsg struct{}

// model field
motionTick int
```

A `tea.Tick(motionResolution, …)` chain (60ms) runs while `motionActive()`
(renamed `spinnerActive`: indeterminate waits ∪ determinate progress ∪ cleanup
running line) and increments `motionTick`. The dispatch wrapper starts the
chain on transitions into a working state exactly as it started spinner ticks;
Init covers models that start working. `pumpCmds` drops `motionTickMsg` like
it drops the old animation messages.

Everything animated is a pure function of the tick:

- `starFrame(tick)` — `starFrames[(tick*motionResolution/starInterval) % len]`,
  frames `· ✢ ✳ ✻ ✽ ✻ ✳ ✢` at 120ms (two ticks per frame).
- `starColor(tick)` — accent→bright ramp synced to frame size (small=dim,
  full=bright) for standalone stars (stat line, fallback contexts).
- `shimmerLine(text, ramp, tick)` — per-rune coloring: band center sweeps
  `[-pad, len+pad]` cyclically every 2.5s; rune intensity is a triangular
  falloff around the center; color = lerp(ramp.base, ramp.peak, intensity);
  output is bold. Rune count and stripped text are invariant.

bubbles/spinner leaves the dependency set for the TUI (stopwatch stays).

## Shimmer ramps as palette data

Ramps live on the palette like every other color decision:

| ramp | dark base → peak | light base → peak |
| --- | --- | --- |
| accent (phase headlines, scans) | `#5f5fd7 → #d3d3ff` | `#5f00d7 → #8b5fe8` |
| body (working step labels) | `#9a9a9a → #ffffff` | `#6c6c6c → #1a1a1a` |
| dim (menu loading hints) | `#6c6c6c → #b0b0b0` | `#9e9e9e → #555555` |

Color lerp is a small hex-parse/interpolate helper; the v2 renderer downsamples
for non-truecolor terminals.

## Vocabulary

| state | glyph | motion |
| --- | --- | --- |
| pending | `·` dim | still |
| working | star morph, in its line's shimmer band | moving |
| done (row) | `✦` green | still |
| success (verdict headline) | `✦` green bold | still |
| failed | `✗` red | still |
| warning | `⚠` orange | still |
| would-act (cleanup preview) | `▸` accent | still |

`indicatorDone = "✦"`; `indicatorSuccess` is deleted and headline sites return
to `indicatorDone`. `indicatorActiveFallback = "✻"` for pure layouts. The core
rule: anything moving is being worked on; anything still is settled. An
invariant test renders every major view and asserts `●` appears nowhere.

## Render sites

- Active phase line: `shimmerLine(star+" "+name, accentRamp, tick)`; counts
  stay dim and static.
- Running step rows: `shimmerLine(star+" "+label, bodyRamp, tick)`.
- Stat line: `starColor(tick)`-styled `starFrame(tick)` (count lines do not
  shimmer).
- Cleanup scan + running line: accent shimmer.
- Menu loading hint: dim shimmer.
- ProgressLayout stays pure: it gains a `Motion` presentation struct (glyph
  plus shimmer closures) injected by the model's render path; nil falls back
  to static styles and `✻`.

## Out of scope

Bar gradient, spring, settle floor, hold mechanics, OSC presence: unchanged.
