## Context

sentei is a single-binary Go CLI tool with no CI/CD infrastructure. The repo is now public on GitHub at `abiswas97/sentei`. The project uses conventional commits consistently. Current version embedding is a single `var version = "dev"` set via ldflags. There are no existing tags.

The goal is a fully automated pipeline: push to feature branch → CI validates → merge PR → release-please creates a Release PR → merge Release PR → GoReleaser cross-compiles, signs, and distributes.

## Goals / Non-Goals

**Goals**:
- PR-based workflow with branch protection on main
- Automated CI on every PR (lint, test on ubuntu + macOS, build, coverage)
- Automated version bumping and changelog via release-please
- Cross-platform binary distribution via GoReleaser (linux/darwin/windows × amd64/arm64)
- Homebrew tap distribution
- Cosign keyless signing of checksums
- Codecov coverage reporting (tokenless) with badge in README
- Conventional commit enforcement (local hook + CI commitlint)
- Dependabot for dependency updates
- Makefile for local dev commands

**Non-Goals**:
- Docker image distribution (no container use case for a CLI tool)
- Homebrew-core submission (premature — pursue after project matures)
- Windows ARM64 builds (low demand, can add later)
- Nix/Scoop/AUR distribution (future consideration)
- Auto-merge of Release PRs (want human review gate)
- CONTRIBUTING.md (skip for now)

## Decisions

### D1: release-please for version automation
**Choice**: Google's release-please over svu or semantic-release.
**Why**: Creates a reviewable Release PR with changelog before tagging. Maintains CHANGELOG.md automatically. No Node.js runtime dependency (runs as GitHub Action). Well-maintained by Google. The PR review gate is valuable — you see exactly what version bump and changelog will be produced before merging.
**Alternatives**: svu (simpler but no changelog, no review gate), semantic-release (Node.js dependency, heavier).
**Bootstrap**: The `.release-please-manifest.json` with `{"." : "0.1.0"}` is the baseline. No manual tag is needed — release-please uses the manifest version as its starting point.

### D2: GoReleaser for build and distribution
**Choice**: GoReleaser (free/OSS tier).
**Why**: Industry standard for Go CLI tools. Handles cross-compilation, archives, checksums, Homebrew formula generation, and cosign signing in a single declarative config. Used by lazygit, charm tools, and most popular Go CLIs.
**Alternatives**: Manual `GOOS/GOARCH` build scripts (fragile, high maintenance).

### D3: Cosign keyless signing from day one
**Choice**: Include cosign signing via GitHub Actions OIDC.
**Why**: Zero key management overhead. Uses GitHub's identity provider for attestation. Industry is moving toward sigstore/cosign for supply chain security. Only ~5 lines of GoReleaser config. Produces a `.sigstore.json` bundle alongside checksums.
**Trade-off**: Users need cosign installed to verify signatures (checksums.txt alone is still useful without it).

### D4: golangci-lint with moderate rule set
**Choice**: Moderate linter config — govet, staticcheck, errcheck, ineffassign, unused, goimports, misspell, revive.
**Why**: Catches real issues without being noisy. Matches what charm tools use. Can tighten over time.
**Alternatives**: Minimal (just govet + errcheck) — too permissive. Exhaustive (20+ linters) — too noisy for a small project.

### D5: Commit enforcement via local hook + CI
**Choice**: Shell-based `commit-msg` git hook (installed via script) + `wagoid/commitlint-github-action` in CI on PRs.
**Why**: Local hook catches format issues before push (fast feedback). CI commitlint validates all commits in PRs (safety net). Shell-based hook has zero dependencies (no Node.js/commitlint install needed locally). Works with PR-based workflow — commitlint runs on PR commits.
**Alternatives**: `pre-commit` framework (heavier dependency), PR title linting (only validates squash-merge title, not individual commits).

### D6: Codecov tokenless for public repo
**Choice**: Codecov with tokenless upload + badge in README.
**Why**: Free for public repos. No token or account setup needed. GitHub Action auto-detects public repos. Badge gives visibility into coverage trends.

### D7: Version output format
**Choice**: `sentei v0.1.0 (abc1234, 2026-02-07)` — version + short commit + date.
**Why**: Helps debug user-reported issues. Standard pattern in Go CLIs. Three ldflags vars: `version`, `commit`, `date`.

### D8: Dependabot for dependency updates
**Choice**: GitHub Dependabot over Renovate.
**Why**: GitHub-native, zero config beyond a YAML file, covers both `gomodules` and `github-actions` ecosystems. Simpler than Renovate for a single-maintainer project. Creates PRs automatically for dependency updates.
**Alternatives**: Renovate (more powerful grouping, but more complex setup — overkill here).

### D9: Makefile for local development
**Choice**: Include a Makefile with common targets: `build`, `test`, `lint`, `snapshot`, `install-hooks`.
**Why**: Industry standard for Go projects. Single consistent interface for all dev commands. `make snapshot` runs `goreleaser build --snapshot --clean` for local testing of GoReleaser config without publishing.

### D10: PR-based workflow with branch protection
**Choice**: All changes go through PRs to main. Branch protection requires CI status checks (lint, test, build, commitlint) to pass before merge. No force pushes or direct pushes to main.
**Why**: Required for CI commitlint to work (only runs on PRs). Pairs naturally with release-please (which creates Release PRs). Prevents accidental direct pushes that bypass CI. Standard for any project with CI/CD.

## Risks / Trade-offs

- **[Risk] release-please Release PR merge conflicts** → Rare in practice; close and let release-please recreate if it happens.
- **[Risk] HOMEBREW_TAP_GITHUB_TOKEN secret expires** → Use a fine-grained PAT with only the tap repo scope. Document renewal.
- **[Risk] GoReleaser OSS limitations** → OSS version covers everything we need. Pro features (Docker manifest, custom publishers) are not needed.
- **[Risk] Cosign verification friction for users** → Checksums.txt works without cosign. Signing is an additional layer, not a replacement.
- **[Risk] Branch protection blocks emergency fixes** → Can temporarily disable rules for genuine emergencies. This is rare and the protection value outweighs it.

## Migration Plan

1. Create `abiswas97/homebrew-tap` repo on GitHub (empty, public)
2. Create `HOMEBREW_TAP_GITHUB_TOKEN` secret in sentei repo settings (PAT with repo scope on the tap repo)
3. Add all config files, workflows, LICENSE, Makefile in a single PR
4. Configure GitHub branch protection rules on main
5. Subsequent PRs merged to main will auto-generate Release PRs via release-please

**Rollback**: Delete workflows and config files. Remove branch protection. No application code is affected.

## Open Questions

None — all decisions resolved during explore and review sessions.
