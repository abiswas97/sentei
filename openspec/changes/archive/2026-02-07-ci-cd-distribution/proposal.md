## Why

sentei has no CI/CD pipeline, no automated releases, and no distribution beyond `go install` from source. Every release would require manual cross-compilation, manual GitHub Release creation, and manual Homebrew formula updates. This blocks the project from having reproducible, verifiable releases and makes distribution to users impractical.

## What Changes

- Add GitHub Actions CI workflow (lint, test, build, coverage) on PRs, with multi-OS test matrix (ubuntu + macOS)
- Add automated versioning via release-please (conventional commit analysis → Release PR → tag)
- Add GoReleaser for cross-compilation, archives, checksums, cosign signing, and Homebrew formula generation
- Add golangci-lint configuration for consistent code quality
- Add Codecov integration for coverage reporting (tokenless, public repo) with badge in README
- Expand version embedding in binary (version + commit SHA + date)
- Add local git commit-msg hook and CI commitlint check to enforce conventional commits
- Create `abiswas97/homebrew-tap` repository for Homebrew distribution
- Switch to PR-based workflow with GitHub branch protection rules on main
- Add Dependabot for automated dependency updates (Go modules + GitHub Actions)
- Add MIT LICENSE file (required for Homebrew)
- Add Makefile for local development commands (build, test, lint, snapshot, install-hooks)
- Add `dist/` to `.gitignore` for GoReleaser output

## Capabilities

### New Capabilities
- `ci-pipeline`: GitHub Actions workflows for linting, testing, building, coverage reporting, Dependabot config, and branch protection rules
- `automated-releases`: release-please for version bumping and changelog generation, GoReleaser for cross-compilation and distribution (Homebrew, archives, checksums, cosign signing)
- `commit-enforcement`: Local git commit-msg hook and CI commitlint check to enforce conventional commit format
- `developer-workflow`: Makefile for local dev commands, LICENSE file, `.gitignore` updates, Codecov badge in README

### Modified Capabilities
- `version-flag`: Expand version output to include commit SHA and build date (format: `sentei v0.1.0 (abc1234, 2026-02-07)`)

## Impact

- **New files**: `.github/workflows/ci.yml`, `.github/workflows/release-please.yml`, `.github/workflows/release.yml`, `.goreleaser.yaml`, `.golangci.yml`, `release-please-config.json`, `.release-please-manifest.json`, `scripts/commit-msg`, `scripts/install-hooks.sh`, `.commitlintrc.json`, `.github/dependabot.yml`, `Makefile`, `LICENSE`
- **Modified files**: `main.go` (version vars expansion), `.gitignore` (add `dist/`), `README.md` (add Codecov badge)
- **New dependencies**: None (all tooling is CI-side or shell scripts)
- **External**: New GitHub repo `abiswas97/homebrew-tap`, Codecov integration (tokenless), GitHub branch protection rules on main
- **Secrets needed**: `HOMEBREW_TAP_GITHUB_TOKEN` (PAT with repo scope for pushing to tap repo)
