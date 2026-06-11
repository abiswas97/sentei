# Cleanup Truth Chain: Design

## Decisions

### D1: The headline states what WILL happen
`DeletableAggressiveCount()` (candidates minus unmerged) drives the headline, the `a` gate, and the footer hint. The bold/warning emphasis always sits on the truth; the dim --force note carries the remainder. A confirm prompt for zero deletions can no longer exist.

### D2: Preview teaches, result mirrors
The clean preview renders the same five no-op check lines the result shows, in the same order, so the user learns coverage before running and can scan-compare after.

### D3: Provenance is labeled
The command echo reads `ran: sentei cleanup --mode …`; the removal tip names the --force caveat instead of promising a no-op.
