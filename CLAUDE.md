# sentei: Git Worktree Cleanup Tool

## Project Overview

sentei is a TUI (Terminal User Interface) tool for managing and cleaning up stale git worktrees in bare repositories. It provides an interactive way to identify, select, and bulk-delete worktrees with parallel execution and progress feedback.

## Tech Stack

- **Language**: Go 1.21+
- **TUI Framework**: Bubble Tea (Elm architecture)
- **UI Components**: Bubbles (progress bars, spinners, lists)
- **Styling**: Lip Gloss

## Development Guidelines

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (should be automatic)
- Keep functions focused and single-purpose
- Prefer explicit error handling over panic
- Use meaningful variable names (no single-letter except for standard Go idioms like `i` in loops)

### Error Handling

- All errors must be handled explicitly
- Error messages should:
  - Clearly state what went wrong
  - Suggest a fix or next step
  - Not expose internal implementation details
- Use `fmt.Errorf` with `%w` for error wrapping to maintain error chains

### Testing

- Write tests for all git parsing logic
- Test edge cases documented in PRD section 9.1
- Use table-driven tests for multiple scenarios

### Git Commands

Always use these exact commands for consistency:

```bash
# List worktrees
git worktree list --porcelain

# Last commit date
git -C <path> log -1 --format=%ai

# Last commit title
git -C <path> log -1 --format=%s

# Check status
git -C <path> status --porcelain

# Remove worktree
git worktree remove --force <path>

# Clean metadata
git worktree prune
```

### Project Structure

```
sentei/
├── main.go                 # Entry point, CLI parsing
├── internal/
│   ├── git/               # Git command execution and parsing
│   ├── tui/               # Bubble Tea UI components
│   └── worktree/          # Business logic (enrichment, deletion)
├── cmd/                   # CLI commands (if using cobra)
└── docs/                  # Documentation
```

Keep internal packages truly internal - they are implementation details.

## Key Design Principles

### Safety First

- Never delete without explicit confirmation in interactive mode
- Clearly warn about uncommitted changes and untracked files
- Prevent deletion of current working directory
- Use status indicators: 🟢 Clean, 🟡 Dirty, 🔴 Untracked files, 🔒 Locked

### Performance

- Parallelize metadata enrichment (target: <5s for 50 worktrees)
- Parallelize deletion operations
- Keep initial scan fast (<2s for 50 worktrees)

### User Experience

- Zero configuration required for basic usage
- Vim-style keyboard shortcuts (j/k) plus arrow keys
- Clear, helpful error messages
- Show progress for long operations

## Common Tasks

### Adding New Git Metadata

1. Add field to `Worktree` struct
2. Add git command to fetch the data
3. Update enrichment logic to populate it (ensure parallelization)
4. Update TUI to display it

### Modifying TUI

1. Update model state in `internal/tui/model.go`
2. Handle messages/events in Update method
3. Render changes in View method
4. Update key bindings if needed

### Testing Git Parsing

Create test fixtures with sample `git worktree list --porcelain` output and verify parsing.

## Edge Cases to Handle

Refer to PRD section 9.1 for complete list. Key ones:

- Not in a git repository
- No worktrees found
- Worktree directory missing (prunable)
- Locked worktrees
- Worktree is current directory
- Permission denied

## Before Committing

- Run `go fmt ./...`
- Run `go vet ./...`
- Ensure all tests pass: `go test ./...`
- Check binary builds: `go build`
- Test with a real bare repo setup

## Dependencies

Use `go get` to add dependencies. Keep dependencies minimal and well-maintained.

Current major dependencies:
- github.com/charmbracelet/bubbletea (TUI framework)
- github.com/charmbracelet/bubbles (UI components)
- github.com/charmbracelet/lipgloss (styling)

## Documentation

- Keep PRD.md as source of truth for requirements
- Update examples when CLI interface changes
- Document all exported functions and types with Go doc comments
