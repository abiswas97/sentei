## Why

sentei has no `--version` flag and no README. Users can't check which version they're running, and there's no user-facing documentation for installation or usage. Both are table-stakes for distributing a CLI tool.

## What Changes

- Add `--version` flag that prints the version string and exits
- Version injected at build time via `-ldflags` so there's a single source of truth
- Create README.md with project description, installation instructions, usage examples, and key bindings reference

## Capabilities

### New Capabilities
- `version-flag`: `--version` CLI flag with build-time version injection
- `readme`: User-facing README.md with installation, usage, and reference docs

### Modified Capabilities

_(none — no existing spec-level behavior changes)_

## Impact

- `main.go`: New `--version` flag handling, version variable
- `README.md`: New file at repo root
- Build process: Requires `-ldflags "-X main.version=..."` for release builds
- No dependency changes
