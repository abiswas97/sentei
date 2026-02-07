# sentei

[![codecov](https://codecov.io/gh/abiswas97/sentei/branch/main/graph/badge.svg)](https://codecov.io/gh/abiswas97/sentei)

A TUI tool for cleaning up stale git worktrees. Scan, select, and bulk-delete worktrees with parallel execution and clear progress feedback.

## Features

- Interactive list with metadata (last commit date, branch, status)
- Multi-select with keyboard navigation
- Parallel deletion with real-time progress
- Safety: confirmation dialogs, warnings for dirty worktrees, branch protection
- Sorting, filtering, and dry-run mode

## Installation

### go install

```bash
go install github.com/abiswas97/sentei@latest
```

### Build from source

```bash
git clone https://github.com/abiswas97/sentei.git
cd sentei
go build -ldflags "-X main.version=$(git describe --tags --always)" -o sentei .
```

## Usage

Run inside a git repository with worktrees:

```bash
sentei                          # current directory
sentei /path/to/bare/repo       # specify repo path
sentei --dry-run                # print summary, no interactive TUI
sentei --version                # print version and exit
sentei --playground             # launch with a temporary test repo
```

### CLI Flags

| Flag | Description |
|------|-------------|
| `--version` | Print version and exit |
| `--dry-run` | Print worktree summary to stdout and exit |
| `--playground` | Create a temporary test repo with sample worktrees |

### Key Bindings

| Key | Action |
|-----|--------|
| `j` / `k` / arrows | Navigate up/down |
| `PgUp` / `PgDn` | Page up/down |
| `Space` | Toggle selection |
| `a` | Select/deselect all |
| `s` | Cycle sort (age, branch) |
| `S` | Reverse sort direction |
| `/` | Filter by branch name |
| `Enter` | Confirm deletion of selected |
| `y` / `n` | Yes/no in confirmation dialog |
| `Esc` | Go back / clear filter |
| `q` / `Ctrl+C` | Quit |

### Status Indicators

| Indicator | Meaning |
|-----------|---------|
| `[ok]` | Clean — no uncommitted changes |
| `[~]` | Dirty — has uncommitted changes |
| `[!]` | Has untracked files |
| `[L]` | Locked |
| `[P]` | Protected branch (cannot be deleted) |

Protected branches: `main`, `master`, `develop`, `dev`.

## License

MIT
