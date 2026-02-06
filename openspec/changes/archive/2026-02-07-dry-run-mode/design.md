## Context

sentei currently always launches an interactive TUI. Users who want a quick preview of worktree state — or want to pipe output to other tools — have no non-interactive option. The existing enrichment and list-rendering logic already gathers all the data needed; we just need a text-output path that bypasses Bubble Tea.

## Goals / Non-Goals

**Goals:**
- Add `--dry-run` flag that prints a tabular worktree summary to stdout and exits
- Reuse existing discovery and enrichment pipeline (no duplication)
- Compose with `--playground` and repo path argument
- Output must be readable in a terminal and useful when piped

**Non-Goals:**
- Machine-parseable output (JSON, CSV) — future work
- Auto-selection of stale worktrees (`--older-than`) — that's F9
- Any deletion or confirmation flow in dry-run mode

## Decisions

**1. Plain text table to stdout, not TUI**

Dry-run prints a simple formatted table using `fmt` / `text/tabwriter` and exits before creating a Bubble Tea program. This keeps it composable with pipes and avoids alt-screen flicker.

Alternative: Render TUI in read-only mode — rejected because it still requires alt-screen and doesn't compose with pipes.

**2. Output format mirrors TUI list columns**

Use the same columns: Status, Branch, Age, Subject. Reuse `statusIndicator` text (without color), `relativeTime`, and `stripBranchPrefix` helpers from `internal/tui/list.go`. This keeps the mental model consistent between dry-run and interactive mode.

Since dry-run output goes to stdout (possibly piped), strip ANSI colors. Use plain text status indicators: `[ok]`, `[~]`, `[!]`, `[L]`.

**3. New `internal/dryrun` package**

A small package with a single `Print(worktrees []git.Worktree, w io.Writer)` function. This keeps main.go thin and the logic testable. It will import helpers from `internal/tui` (exported as needed) or duplicate the trivial ones (`relativeTime`, `stripBranchPrefix` are small enough to inline or export).

Alternative: Put it in `internal/tui` — rejected because dry-run has no TUI dependency and shouldn't pull in Bubble Tea.

**4. Sorting in dry-run uses the same default as TUI (age ascending)**

Apply the same default sort (oldest first) so the output matches what users would see on launch in interactive mode.

## Risks / Trade-offs

- [Duplicating helpers] → Accept minor duplication of `relativeTime` and `stripBranchPrefix` in the dryrun package rather than exporting from tui (which would couple dryrun to tui package). These are simple, stable functions.
