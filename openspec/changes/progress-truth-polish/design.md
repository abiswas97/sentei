## Context

Four small honesty leaks measured by the June 2026 UX audit, independent of the structural progress work except where noted. All presentation-layer; all decided in the session FAQ.

## Goals / Non-Goals

**Goals**: every dim/meta element either tells the truth or is absent. **Non-Goals**: structural progress changes (other changes in this arc); any new layout vocabulary.

## Decisions

**D1: Suppress elapsed under 2s rather than render sub-second precision.**
Alternative (showing `0.4s`) rejected: sub-second elapsed on a TUI flow is noise that implies precision the 1Hz repaint ticker does not have; absence is honest. The reserve stays so appearance does not reflow the bar.

**D2: Failed phase headers drop the percent (FAQ: chosen over split `1 done · 1 failed` counts).**
Minimal change, no second header format to maintain; the ✗ indicator plus attempted/total counts carry the meaning.

**D3: Skip traces as dim per-step lines (FAQ: chosen over summary-count-only).**
Auditable per worktree/integration, which is the point after the ccc detection incident; reuses the existing dim `– skipped` vocabulary, so no new styling.

**D4: Errors title is a `copy.go` const, red count leads.**
Voice-registry rule: all titles in copy.go, one-line edits. Ordering follows the audit finding that green-first momentarily reads as success.

## Risks / Trade-offs

- [Golden churn] → regenerate only affected views, frame-review before commit.
- [Failed-header tests written against pre-declarations semantics] → sequenced after `progress-declarations` (noted in tasks 2.1).

## Migration Plan

Single PR after `progress-declarations` merges. Rollback: revert.

## Open Questions

None.
