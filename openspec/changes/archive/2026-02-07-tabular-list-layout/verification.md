# Verification Report: tabular-list-layout

## Summary

**Status**: PASS (with minor doc drift)

16/18 tasks complete. 2 manual testing tasks (4.2, 4.3) were performed during development but not formally checked off. All spec requirements are implemented. One design.md statement is stale.

---

## 1. Completeness — Tasks vs Implementation

| Task | Status | Notes |
|------|--------|-------|
| 1.1 Add `width` field to Model | DONE | `model.go:24` |
| 1.2 Capture `msg.Width` in WindowSizeMsg | DONE | `list.go:94` |
| 2.1 Import `lipgloss/table` | DONE | `list.go:11` |
| 2.2 Remove old width constants | DONE | `colWidthBranch`/`colWidthAge`/`colWidthSubject` gone |
| 2.3 Rewrite viewList with lipgloss/table | DONE | `list.go:162-279` |
| 2.4 StyleFunc with fixed widths + row styles | DONE | `list.go:243-272` |
| 2.5 Set Width/Wrap on table | DONE | `list.go:183`, `list.go:180` |
| 2.6 Render only visible slice | DONE | `list.go:174`, loop `m.offset..end` |
| 3.1 Column width constants | DONE | `styles.go:59-62` |
| 3.2 Proportional branch/subject widths | DONE | `list.go:186-193` |
| 3.3 Padding on columns | DONE | Data columns only (see drift below) |
| 4.1 Update test assertions | DONE | Tests pass |
| 4.2 Manual playground test | **NOT CHECKED** | Was done live during dev |
| 4.3 Manual real-repo test | **NOT CHECKED** | Was done live during dev |
| 4.4 go vet + go test | DONE | Clean |
| 5.1 Update design.md | DONE | |
| 5.2 Update spec.md | DONE | |
| 5.3 Update tasks.md | DONE | |

## 2. Correctness — Spec Requirements vs Implementation

### MODIFIED: Display enriched worktrees in a scrollable list

| Scenario | Verdict | Evidence |
|----------|---------|----------|
| Normal worktree row | PASS | 6-column table with cursor, checkbox, status, branch, age, subject (`list.go:240`) |
| Dirty worktree row | PASS | `statusIndicator` returns `[~]` for `HasUncommittedChanges` (`list.go:82-83`) |
| Untracked files row | PASS | `statusIndicator` returns `[!]` for `HasUntrackedFiles` (`list.go:84-85`) |
| Locked worktree row | PASS | `statusIndicator` returns `[L]` for `IsLocked` (`list.go:80-81`) |
| Bare repository excluded | PASS | Filtering happens upstream before Model receives worktrees |
| Prunable worktree row | PASS | Shows `(prunable)` as branch name (`list.go:218-219`) |
| Enrichment error | PASS | Shows `"error"` age and error message as subject (`list.go:225-228`) |

### ADDED: Responsive column layout with wrapping

| Scenario | Verdict | Evidence |
|----------|---------|----------|
| Wide terminal | PASS | Proportional split gives ample room; confirmed in 120-col test |
| Narrow terminal | PASS | `max(..., 20)` floor prevents negative widths (`list.go:191`) |
| Long branch name wrapping | PASS | `.Wrap(true)` on table, branch column has no pre-truncation |
| Terminal resize | PASS | `WindowSizeMsg` updates `m.width`, table rebuilt each render |

### ADDED: Terminal width tracking

| Scenario | Verdict | Evidence |
|----------|---------|----------|
| Initial width | PASS | `m.width = msg.Width` in `updateList` (`list.go:94`) |
| Width update on resize | PASS | Same handler fires on every resize |

### Subject truncation (evolved requirement, not in original spec)

| Behavior | Verdict | Evidence |
|----------|---------|----------|
| Subject pre-truncated with `...` | PASS | `list.go:230-238`, uses `lipgloss.Width()` + `[]rune` slicing |
| Only branch wraps, subject does not | PASS | Confirmed in manual testing with saaf-monorepo |

## 3. Coherence — Artifacts vs Implementation

### design.md

| Statement | Current? | Issue |
|-----------|----------|-------|
| "All columns get `Padding(0, 1)` via StyleFunc" (line 90) | **STALE** | Only data columns (branch, age, subject) get `Padding(0, 1)`. Prefix columns (cursor, checkbox, status) bake gaps into their fixed Width instead. |
| "Total padding budget: 6 chars (6 columns x 1 char)" (line 90) | **STALE** | Actual padding budget is 3 chars (3 data columns x 1 char). `colPadding := 3` in `list.go:190`. |
| Column width table (lines 36-44) shows cursor=2, checkbox=3, status=4 | **STALE** | Actual: cursor=3, checkbox=5, status=6 (include inter-column gap). |

### tasks.md

| Statement | Current? | Issue |
|-----------|----------|-------|
| Task 3.1 says "cursor=2, checkbox=4, status=5" | **STALE** | Actual: cursor=3, checkbox=5, status=6. |
| Task 3.3 says "Add `Padding(0, 1)` to all column styles" | **STALE** | Only data columns get padding. |

### spec.md

| Statement | Current? | Issue |
|-----------|----------|-------|
| "All columns SHALL have consistent 1-char right padding" (line 37) | **SLIGHTLY INACCURATE** | Prefix columns achieve gap via extra Width chars, not via Padding. Net visual effect is the same (consistent gaps), but mechanism differs. |
| No mention of subject truncation | **MISSING** | Subject pre-truncation with `...` is an evolved requirement not captured in spec. |

### proposal.md

No issues — describes the problem and impact accurately.

## 4. Recommendations

1. **Fix design.md line 90**: Change "All columns" to "Data columns (branch, age, subject)" and update padding budget from 6 to 3.
2. **Fix design.md column width table**: Update cursor=3, checkbox=5, status=6 to match implementation.
3. **Fix tasks.md 3.1/3.3**: Update constants and padding description to match.
4. **Add subject truncation to spec.md**: Add a scenario under "Responsive column layout" for subject truncation behavior.
5. **Mark tasks 4.2/4.3 complete**: They were performed during development.
