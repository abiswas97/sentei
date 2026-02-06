## Context

sentei currently allows selecting any non-bare worktree for deletion. There is no mechanism to prevent selecting worktrees for well-known long-lived branches like `main` or `develop`. The confirmation dialog warns about dirty/locked worktrees but does not distinguish important branches from ephemeral ones.

## Goals / Non-Goals

**Goals:**
- Prevent accidental deletion of worktrees for well-known long-lived branches
- Zero configuration — works out of the box with sensible defaults

**Non-Goals:**
- User-configurable protected branch patterns (no config file)
- Protecting branches based on remote tracking or merge status

## Decisions

### Decision 1: Pure function over struct field

Add a `IsProtectedBranch(branch string) bool` function in `internal/git/` rather than adding an `IsProtected` field to the `Worktree` struct.

**Rationale:** Protection is a UI concern derived from the branch name, not intrinsic worktree metadata. A pure function is simpler, has no state to keep in sync, and is trivially testable. The caller (TUI) calls it when needed.

**Alternative considered:** Adding `IsProtected bool` to the `Worktree` struct and populating it during enrichment. Rejected because it couples a UI policy to the data model and requires threading the logic through enrichment.

### Decision 2: Protected branch list

Protect exact matches against the short branch name (after stripping `refs/heads/`): `main`, `master`, `develop`, `dev`.

**Rationale:** These are the most common long-lived branches across Gitflow and trunk-based workflows. The list is intentionally small to avoid false positives.

### Decision 3: `[P]` indicator replaces checkbox

Protected worktrees display `[P]` in place of the checkbox column (`[ ]`/`[x]`). This makes it visually clear the worktree cannot be selected without adding a new column.

### Decision 4: Spacebar and select-all silently skip protected

Pressing spacebar on a protected worktree does nothing. Select-all (`a`) skips protected worktrees. No error message or toast — the `[P]` indicator is sufficient context.

## Risks / Trade-offs

- [False positives] A user with a feature branch literally named `dev` cannot select it → Acceptable given how rare this is and the safety benefit
- [Missing protection] Teams using non-standard names like `development` or `trunk` are not protected → Acceptable for v1; config file is a future option if needed
- [Detached HEAD] A detached worktree checked out at the same commit as `main` is NOT protected (protection is branch-name-based, not commit-based) → Correct behavior
