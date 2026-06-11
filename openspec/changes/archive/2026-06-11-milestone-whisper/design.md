# Design: milestone-whisper

`recordRemovals(bareDir, n)` is a tea.Cmd: load → add → atomic save → report
`crossedPowerOfTen(before, after)` (largest p with before < p <= after). Any
state error returns an empty milestone — the whisper is garnish and never
surfaces failure; the counter is simply not advanced. CLI (non-TUI) removals
do not count; the whisper is a TUI moment. The copy lives in the voice
registry; powers of ten always take "th".
