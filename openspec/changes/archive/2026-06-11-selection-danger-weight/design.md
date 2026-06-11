# Weight: Design

## Decisions

### D1: One canonical label
`worktreeLabel` owns naming (branch / short-hash detached / prunable / dir fallback); the list's inline derivation is deleted. A worktree correlates across list, confirm, portal, and progress by construction.

### D2: Danger is the first binding
`viewFooterDanger(width, bindings)` styles binding zero with the warning token and the rest as a normal footer: declarative danger, no literals, reusable for future destructive confirms.

### D3: Weight without new color
Selection and focus reuse the accent; danger reuses warning. No new palette entries.
