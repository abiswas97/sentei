## 1. External Setup

- [x] 1.1 Create `abiswas97/homebrew-tap` GitHub repo (public, empty, with a README)
- [x] 1.2 Create a fine-grained PAT with repo scope on `abiswas97/homebrew-tap` and add it as `HOMEBREW_TAP_GITHUB_TOKEN` secret in `abiswas97/sentei` repo settings

## 2. Project Files

- [x] 2.1 Add MIT `LICENSE` file at repo root
- [x] 2.2 Add `dist/` to `.gitignore`
- [x] 2.3 Expand `main.go` version vars to `version`, `commit`, `date` and update `--version` output to format `sentei <version> (<commit>, <date>)`

## 3. Makefile

- [x] 3.1 Create `Makefile` with targets: build, test, lint, snapshot, install-hooks

## 4. Linter Configuration

- [x] 4.1 Create `.golangci.yml` with linter set: govet, staticcheck, errcheck, ineffassign, unused, goimports, misspell, revive (exclude errcheck on test files)
- [x] 4.2 Run `make lint` locally and fix any violations

## 5. Commit Enforcement

- [x] 5.1 Create `scripts/commit-msg` hook script that validates conventional commit format via regex (allow merge commits)
- [x] 5.2 Create `scripts/install-hooks.sh` that symlinks the hook to `.git/hooks/commit-msg`
- [x] 5.3 Create `.commitlintrc.json` config for the CI commitlint action

## 6. CI Workflow

- [x] 6.1 Create `.github/workflows/ci.yml` with lint, test (race + coverage on ubuntu + macOS matrix), build, and commitlint jobs triggered on push to main and PRs. Enable Go module caching via actions/setup-go.
- [x] 6.2 Add Codecov upload step (tokenless) to the test job, configured to not fail on upload errors
- [x] 6.3 Add Codecov badge to README.md

## 7. Dependency Automation

- [x] 7.1 Create `.github/dependabot.yml` for gomodules and github-actions ecosystems (weekly schedule)

## 8. Release Automation

- [x] 8.1 Create `release-please-config.json` (release-type: go, component: sentei)
- [x] 8.2 Create `.release-please-manifest.json` with initial version `0.1.0`
- [x] 8.3 Create `.github/workflows/release-please.yml` triggered on push to main

## 9. GoReleaser

- [x] 9.1 Create `.goreleaser.yaml` with builds (CGO_ENABLED=0, ldflags, cross-compile matrix), archives (tar.gz + zip for windows), checksums, cosign signing, homebrew tap push, and changelog config
- [x] 9.2 Create `.github/workflows/release.yml` triggered on tag `v*` that runs GoReleaser with cosign
- [x] 9.3 Run `make snapshot` locally to validate the GoReleaser config

## 10. Branch Protection

- [x] 10.1 Configure GitHub branch protection on main via `gh api`: require status checks (lint, test, build, commitlint), block force pushes, block direct pushes
