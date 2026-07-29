# Worktree File Propagation Design

## Goal

Sentei will copy explicitly configured, repository-private files from the bare-repository container into every worktree it creates. This is a separate capability from ecosystem-specific `env_files`.

## Configuration

`.sentei.yaml` gains a top-level `worktree_files` list:

```yaml
worktree_files:
  - source: .sentei/private/codex-config.toml
    destination: .codex/config.toml
    overwrite: false
```

- `source` is relative to the bare-repository container that owns `.sentei.yaml`.
- `destination` is relative to the newly created worktree.
- `overwrite` is optional and defaults to `false`.
- A non-empty value in a later config layer replaces the earlier layer, matching Sentei's existing scalar-list merge behavior.
- An absent or empty list adds no work and preserves existing creation behavior.

`worktree_files` applies automatically to both interactive and non-interactive creation. It has no opt-in CLI flag because the repository owner has declared the files part of every new worktree's setup contract.

## Architecture

`internal/worktreefile` owns the rule contract, static path validation, containment checks, and file copying. `internal/config` parses, merges, and validates rules. `internal/creator` adds one shared Setup step after worktree creation and merge, before dependency and integration setup. CLI and TUI pass the same loaded rules into `creator.Options`.

The copy package accepts only the bare-repository root, new-worktree root, and validated rules. It does not know about ecosystems, CLI flags, or TUI state.

## Safety and Copy Semantics

Static validation rejects blank paths, absolute paths, paths that clean to `.`, and any `..` component. Runtime validation:

1. Resolves the source against the real bare-repository root and rejects symlink resolution outside it.
2. Requires the source to be a regular file.
3. Rejects symlinks in every existing destination component, including the destination itself.
4. Creates missing destination parents with directory mode `0755`.
5. Rechecks containment after parent creation.
6. Copies through a temporary file in the destination directory, applies the source permission bits, and renames atomically.

Existing destinations are preserved by default. Their rule is reported as skipped. `overwrite: true` replaces only a regular, contained destination and preserves the source file's permission bits. File contents are never included in progress messages or errors.

## Failure Behavior

A missing source, unsafe path, or copy failure marks the `Copy worktree files` Setup step failed with a path-only diagnostic. The worktree remains available and independent dependency or integration steps continue, matching the creator pipeline's established partial-failure contract. CLI creation exits non-zero through `Result.HasFailures`; TUI creation shows the failed Setup step.

An invalid configured path fails configuration loading before creation. This avoids creating a worktree while silently dropping declared private setup.

## Compatibility

Existing `env_files`, ecosystem detection and installation, integrations, merge behavior, progress events, CLI flags, and TUI options remain unchanged. With no `worktree_files`, the creator plan and output are byte-for-byte equivalent apart from documentation.

No sync command for existing worktrees is added.

## Verification

Tests cover:

- YAML parsing, config-layer replacement, empty config, and every invalid path class.
- Nested destinations, parent creation, default no-overwrite, explicit overwrite, permissions, missing source, copy errors, absolute/traversal rejection, source symlink escape, and destination symlink escape.
- Creator planning, ordering, failure visibility, continuation of independent steps, and empty-rule regression behavior.
- CLI and TUI forwarding to the shared creator contract.
- A real temporary bare repository whose private container template is copied into a created worktree.
- Existing `env_files` regression behavior.

