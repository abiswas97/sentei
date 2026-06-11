# Design: integration-robustness

## Detection

`detectTool` default: `command -v <binary>` (was `<binary> --version`). One
line; explicit `Detect.Command` still wins. POSIX `command -v` is supported
by every sh sentei shells through.

## Error peek

New pure helper:

```go
// errorPeekLines bounds a multi-line error: first line, last non-empty
// line (CLI installers end with `error: …`), and the elided count.
func errorPeekLines(errText string, width int) []string
```

Summaries (apply + create) render the peek under the failed step: first
line dim, last line error-styled, `… N more — ? for full output` dim. The
live progress row uses only the last non-empty line, ellipsis-truncated
(one-line rows law). Portal paths render the untrimmed error.

`renderIntegrationOutcomes` gains a width/full flag separating the inline
peek from the portal's full rendering.
