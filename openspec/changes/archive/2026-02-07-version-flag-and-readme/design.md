## Context

sentei is a single-binary CLI tool. It currently uses `flag` from the Go stdlib for CLI parsing. There is no version tracking or user-facing documentation beyond the PRD and CLAUDE.md (both developer-facing).

## Goals / Non-Goals

**Goals:**
- Users can check which version of sentei they're running
- Version is set at build time (no manual version string maintenance)
- README provides enough to install, run, and use sentei

**Non-Goals:**
- Automated release pipeline (goreleaser, CI/CD) — future work
- Shell completions generation
- Man page generation
- Changelog management

## Decisions

### D1: Version injection via `ldflags`

Use `go build -ldflags "-X main.version=<value>"` to set the version at build time.

- A `var version = "dev"` in `main.go` provides a sensible default for `go install` / dev builds
- No version file, no generated code, no build tags
- **Alternative**: embed a `VERSION` file — rejected because ldflags is the standard Go pattern and avoids an extra file

### D2: `--version` as a flag, not a subcommand

Add `--version` as a boolean flag in the existing `flag` package usage. When set, print `sentei <version>` to stdout and exit.

- Consistent with the existing flag-based CLI (no cobra/subcommands)
- **Alternative**: `sentei version` subcommand — rejected because the tool has no subcommands and adding one just for version would be inconsistent

### D3: README structure

Keep the README concise and action-oriented:
- One-line description + what it looks like (terminal screenshot placeholder)
- Installation (go install, binary download)
- Usage (basic commands, key bindings table)
- Link to PRD for deeper details

No badges, no extensive API docs, no contributing guide yet.

## Risks / Trade-offs

- [Risk] Users building with plain `go build` get "dev" as version → Acceptable; documented in README
- [Risk] README screenshots go stale → Use text-based examples instead of images for now
