# sentei: Git Worktree Cleanup Tool

## Product Requirements Document (PRD)

**Version:** 1.0
**Last Updated:** February 2026
**Status:** In Development

---

## 1. Executive Summary

**sentei** is a terminal user interface (TUI) tool for managing and cleaning up stale git worktrees in bare repositories. It provides an interactive, visual way to identify, select, and bulk-delete worktrees that are no longer needed, with parallel execution and clear progress feedback.

---

## 2. Problem Statement

### The Pain Point

Developers using git worktrees with bare repositories accumulate stale worktrees over time. These are worktrees created for:
- Feature branches that have been merged
- Experimental work that was abandoned
- Hotfixes that are long complete
- Code reviews that are finished

### Current Workflow (Manual)

```bash
# List worktrees
git worktree list

# For each stale worktree, manually:
# 1. Check if it has uncommitted changes
git -C /path/to/worktree status

# 2. Check last activity
git -C /path/to/worktree log -1 --format=%ai

# 3. Remove it
git worktree remove /path/to/worktree --force
```

### Problems with Current Workflow

| Problem | Impact |
|---------|--------|
| **No overview of worktree health** | Can't quickly see which worktrees are stale |
| **No visibility into uncommitted changes** | Risk of data loss when removing |
| **Manual, sequential deletion** | Time-consuming for many worktrees |
| **No bulk operations** | Must remove one at a time |
| **Easy to make mistakes** | Might delete active worktree by accident |

### Scale of Problem

In a typical bare repo setup:
- 10-30 worktrees accumulate over months
- 50-80% become stale after branch merges
- Manual cleanup takes 5-15 minutes
- Often postponed, leading to disk space issues

---

## 3. Solution Overview

**sentei** provides a TUI that:

1. **Scans** a bare repository for all worktrees
2. **Enriches** each worktree with metadata (last commit date, uncommitted changes, branch status)
3. **Displays** an interactive, sortable list with visual indicators
4. **Allows** multi-selection of worktrees for deletion
5. **Executes** parallel deletion with real-time progress
6. **Confirms** before any destructive action

### Value Proposition

| Before sentei | After sentei |
|-----------------|----------------|
| 5-15 min manual cleanup | 30 seconds interactive cleanup |
| Risk of losing uncommitted work | Clear warnings for dirty worktrees |
| Sequential, slow deletion | Parallel deletion with progress |
| No overview of worktree age | Sorted by last activity date |
| CLI-only, error-prone | Visual TUI, hard to make mistakes |

---

## 4. User Personas

### Primary: Senior Developer / Tech Lead

- Manages a monorepo with bare repo + worktree setup
- Creates 3-5 worktrees per week for features/reviews
- Needs to clean up periodically (weekly/monthly)
- Values speed and safety

### Secondary: DevOps / Platform Engineer

- Maintains CI/CD pipelines that create temporary worktrees
- Needs automated or scriptable cleanup
- May run in non-interactive mode

---

## 5. Functional Requirements

### 5.1 Core Features (MVP)

#### F1: Worktree Discovery
- Parse output of `git worktree list --porcelain`
- Support bare repositories and regular repositories with worktrees
- Handle edge cases: prunable worktrees, locked worktrees

#### F2: Metadata Enrichment
For each worktree, gather:

| Metadata | Source Command | Purpose |
|----------|---------------|---------|
| Branch name | `git worktree list --porcelain` | Identification |
| Last commit date | `git -C <path> log -1 --format=%ai` | Staleness indicator |
| Last commit title | `git -C <path> log -1 --format=%s` | Context |
| Uncommitted changes | `git -C <path> status --porcelain` | Safety warning |
| Untracked files | `git -C <path> status --porcelain` | Safety warning |
| Locked status | `git worktree list --porcelain` | Prevent accidental deletion |
| Worktree path | `git worktree list --porcelain` | Display and operations |

#### F3: Interactive TUI
- Display worktrees in a scrollable list
- Show columns: Branch | Last Activity | Last Commit | Status
- Status indicators:
  - 🟢 Clean (no uncommitted changes)
  - 🟡 Dirty (uncommitted changes)
  - 🔴 Untracked files present
  - 🔒 Locked
- Keyboard navigation (j/k or arrows)
- Multi-select with spacebar
- Select all / deselect all shortcuts

#### F4: Confirmation Dialog
- Show summary of selected worktrees
- Warn about worktrees with uncommitted changes
- Warn about worktrees with untracked files
- Require explicit confirmation (type "yes" or press specific key combo)

#### F5: Parallel Deletion
- Execute `git worktree remove --force <path>` for each selected worktree
- Run deletions concurrently (configurable parallelism)
- Show real-time progress bar
- Display per-worktree status (pending/in-progress/success/failed)
- Collect and display any errors at the end

#### F6: Post-Deletion Summary
- Show count of successfully removed worktrees
- List any failures with error messages
- Suggest `git worktree prune` if orphaned metadata exists

### 5.2 Extended Features (Post-MVP)

#### F7: Sorting and Filtering
- Sort by: last activity (default), branch name
- Filter by: branch name substring (interactive `/` search)

#### F8: Dry Run Mode
- `--dry-run` flag to preview what would be deleted
- No confirmation required, just shows the plan

#### F9: Non-Interactive Mode
- `--yes` flag to skip confirmation
- `--older-than=30d` to auto-select stale worktrees
- For CI/automation use cases

#### F10: Configuration File
- `.sentei.yaml` in repo root or home directory
- Configure: default sort, parallelism, protected branches

#### F11: Branch Protection
- Never show/select certain branches (e.g., main, develop)
- Configurable protected branch patterns

---

## 6. Non-Functional Requirements

### 6.1 Performance
- Initial scan: < 2 seconds for 50 worktrees
- Metadata enrichment: < 5 seconds for 50 worktrees (parallelized)
- Deletion: Limited by git, but parallel execution should maximize throughput

### 6.2 Reliability
- Never delete without explicit user confirmation (interactive mode)
- Gracefully handle partially failed deletions
- No data corruption under any circumstances

### 6.3 Usability
- Zero configuration required for basic usage
- Clear, readable TUI with adequate contrast
- Intuitive keyboard shortcuts (vim-style + arrows)
- Helpful error messages with suggested fixes

### 6.4 Portability
- macOS (primary)
- Linux (full support)
- Windows (best effort, WSL recommended)

### 6.5 Distribution
- Single binary with no runtime dependencies
- Available via: direct download, Homebrew, go install

---

## 7. Technical Architecture

### 7.1 Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go 1.21+ | Single binary, excellent concurrency, CLI ecosystem |
| TUI Framework | Bubble Tea | Elm architecture, composable, active community |
| UI Components | Bubbles | Progress bars, spinners, lists |
| Styling | Lip Gloss | Terminal styling, colors |

### 7.2 Project Structure

```
sentei/
├── main.go                 # Entry point, CLI parsing
├── go.mod
├── go.sum
├── internal/
│   ├── git/
│   │   ├── worktree.go     # Worktree struct and methods
│   │   ├── parser.go       # Parse git worktree list output
│   │   ├── commands.go     # Execute git commands
│   │   └── git_test.go
│   ├── tui/
│   │   ├── model.go        # Main Bubble Tea model
│   │   ├── list.go         # Worktree list component
│   │   ├── confirm.go      # Confirmation dialog
│   │   ├── progress.go     # Deletion progress view
│   │   ├── styles.go       # Lip Gloss styles
│   │   └── keys.go         # Key bindings
│   └── worktree/
│       ├── enricher.go     # Parallel metadata fetching
│       ├── deleter.go      # Parallel deletion logic
│       └── filter.go       # Filtering and sorting
├── cmd/
│   └── root.go             # CLI command definitions (if using cobra)
└── docs/
    ├── PRD.md
    └── TASKS.md
```

### 7.3 Data Flow

```
┌─────────────────┐
│ git worktree    │
│ list --porcelain│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Parser          │──────▶ []Worktree (basic info)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Enricher        │──────▶ []Worktree (with metadata)
│ (parallel)      │        - last commit date
└────────┬────────┘        - uncommitted changes
         │                 - etc.
         ▼
┌─────────────────┐
│ TUI List View   │◀─────▶ User interaction
└────────┬────────┘        - navigation
         │                 - selection
         ▼
┌─────────────────┐
│ Confirmation    │──────▶ User confirms
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Deleter         │──────▶ Parallel deletion
│ (parallel)      │        with progress updates
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Summary View    │──────▶ Results displayed
└─────────────────┘
```

---

## 8. User Experience

### 8.1 CLI Interface

```bash
# Basic usage - run in a bare repo or repo with worktrees
sentei

# Specify repo path
sentei /path/to/bare/repo

# Dry run
sentei --dry-run

# Non-interactive (for scripts)
sentei --yes --older-than=30d

# Show version
sentei --version
```

### 8.2 TUI Wireframe

```
┌─────────────────────────────────────────────────────────────────────┐
│ sentei - Git Worktree Cleanup                          [?] Help   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   [ ] 🟢 feature/auth-refactor     3 days ago    "Add OAuth2 flow" │
│   [x] 🟢 bugfix/login-redirect     45 days ago   "Fix redirect..."  │
│   [x] 🟡 experiment/new-ui         120 days ago  "WIP: new design"  │
│ > [ ] 🔒 release/v2.1              2 days ago    "Bump version"     │
│   [x] 🟢 chore/deps-update         90 days ago   "Update deps"      │
│   [ ] 🔴 tmp/code-review-123       200 days ago  "Review changes"   │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ 3 selected │ Space: toggle │ a: all │ Enter: delete │ q: quit      │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.3 Confirmation Dialog

```
┌─────────────────────────────────────────────────────────────────────┐
│                        ⚠️  Confirm Deletion                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   You are about to delete 3 worktrees:                              │
│                                                                     │
│   • bugfix/login-redirect (clean)                                   │
│   • experiment/new-ui ⚠️  HAS UNCOMMITTED CHANGES                   │
│   • chore/deps-update (clean)                                       │
│                                                                     │
│   ⚠️  1 worktree has uncommitted changes that will be LOST          │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│              [Y] Yes, delete │ [N] No, go back                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.4 Progress View

```
┌─────────────────────────────────────────────────────────────────────┐
│                       Removing Worktrees                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ████████████████████░░░░░░░░░░░░░░░░░░░░  2/3 (67%)              │
│                                                                     │
│   ✓ bugfix/login-redirect    removed                                │
│   ✓ experiment/new-ui        removed                                │
│   ⏳ chore/deps-update        removing...                            │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 9. Edge Cases and Error Handling

### 9.1 Edge Cases

| Scenario | Behavior |
|----------|----------|
| Not in a git repository | Show error: "Not a git repository. Run from a repo or specify path." |
| No worktrees found | Show message: "No worktrees found (only the main working tree exists)." |
| Worktree directory missing | Mark as "prunable", suggest `git worktree prune` |
| Locked worktree selected | Show lock icon, warn in confirmation, use `--force` |
| Worktree is current directory | Prevent deletion, show error |
| Branch checked out elsewhere | Show warning, may fail to delete |
| Bare repo without worktrees | Show message explaining bare repos need worktrees |
| Permission denied on delete | Show error, continue with others, summarize failures |
| Git command not found | Show error: "git not found in PATH" |

### 9.2 Error Messages

All error messages should:
- Clearly state what went wrong
- Suggest a fix or next step
- Not expose internal implementation details

Example:
```
❌ Failed to remove 'feature/old-branch'

   Error: Worktree is locked

   To unlock and retry:
     git worktree unlock /path/to/feature/old-branch

   Or select "force delete locked" option in sentei
```

---

## 10. Success Metrics

| Metric | Target |
|--------|--------|
| Time to clean 10 stale worktrees | < 30 seconds (vs 5+ min manual) |
| User errors (accidental deletions) | 0 (due to confirmations) |
| Supported edge cases | 100% of documented cases |
| Binary size | < 15 MB |
| Startup time | < 500ms |

---

## 11. Future Considerations

- **Integration with git hooks**: Auto-suggest cleanup after merge
- **Remote tracking**: Show if branch has been merged to main
- **Worktree creation**: Full lifecycle management (not just deletion)
- **TUI themes**: Light/dark mode, custom color schemes
- **Shell completions**: Bash, Zsh, Fish

---

## 12. Glossary

| Term | Definition |
|------|------------|
| **Bare repository** | A git repo without a working tree, containing only the `.git` directory contents |
| **Worktree** | A linked working tree that allows checking out a branch in a separate directory |
| **Stale worktree** | A worktree for a branch that is no longer actively being worked on |
| **Prunable worktree** | A worktree whose directory has been deleted but git metadata remains |
| **Locked worktree** | A worktree marked as locked to prevent accidental removal |

---

## Appendix A: Git Worktree Commands Reference

```bash
# List all worktrees (human readable)
git worktree list

# List all worktrees (machine parseable)
git worktree list --porcelain

# Remove a worktree
git worktree remove <path>

# Force remove (even with uncommitted changes)
git worktree remove --force <path>

# Clean up stale worktree metadata
git worktree prune

# Lock a worktree
git worktree lock <path>

# Unlock a worktree
git worktree unlock <path>
```

## Appendix B: Sample `git worktree list --porcelain` Output

```
worktree /Users/dev/repo
bare

worktree /Users/dev/repo/main
HEAD abc123def456...
branch refs/heads/main

worktree /Users/dev/repo/feature-x
HEAD def456abc789...
branch refs/heads/feature-x

worktree /Users/dev/repo/locked-branch
HEAD 789abc123def...
branch refs/heads/locked-branch
locked
```
