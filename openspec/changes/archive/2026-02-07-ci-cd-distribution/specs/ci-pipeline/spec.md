## ADDED Requirements

### Requirement: CI runs lint, test, and build on every push and PR
The CI workflow SHALL trigger on pushes to `main` and on pull requests targeting `main`. It SHALL run three jobs: lint (golangci-lint), test (go test with race detector and coverage), and build (go build). All three jobs MUST pass for the workflow to succeed.

#### Scenario: Push to main triggers CI
- **WHEN** a commit is pushed to the `main` branch
- **THEN** the CI workflow runs lint, test, and build jobs

#### Scenario: PR targeting main triggers CI
- **WHEN** a pull request is opened or updated against `main`
- **THEN** the CI workflow runs lint, test, and build jobs

#### Scenario: Lint failure blocks CI
- **WHEN** golangci-lint reports violations
- **THEN** the lint job fails and the overall workflow fails

#### Scenario: Test failure blocks CI
- **WHEN** any test fails or the race detector finds issues
- **THEN** the test job fails and the overall workflow fails

### Requirement: Tests run on multiple operating systems
The CI test job SHALL run on a matrix of ubuntu-latest and macos-latest. Both OS runs MUST pass for the test job to succeed.

#### Scenario: Tests pass on both OSes
- **WHEN** tests are run via CI
- **THEN** they execute on both ubuntu-latest and macos-latest

#### Scenario: macOS-only failure blocks CI
- **WHEN** a test fails on macos-latest but passes on ubuntu-latest
- **THEN** the test job fails

### Requirement: Go modules are cached in CI
The CI workflow SHALL cache Go modules between runs using `actions/setup-go` with caching enabled. This reduces CI run time by avoiding redundant dependency downloads.

#### Scenario: Cached modules used on subsequent runs
- **WHEN** CI runs and a cache exists from a previous run
- **THEN** Go modules are restored from cache instead of re-downloaded

### Requirement: golangci-lint configuration exists
The project SHALL include a `.golangci.yml` configuration file with at minimum these linters enabled: govet, staticcheck, errcheck, ineffassign, unused, goimports, misspell, revive. Test files SHALL be excluded from errcheck rules.

#### Scenario: Linter config is present and valid
- **WHEN** `golangci-lint run` is executed
- **THEN** it loads `.golangci.yml` and applies the configured linter set

### Requirement: Coverage is uploaded to Codecov
The CI test job SHALL generate a coverage profile and upload it to Codecov using tokenless upload. The upload SHALL NOT fail the workflow if Codecov is unreachable.

#### Scenario: Coverage uploaded on push to main
- **WHEN** tests pass on a push to `main`
- **THEN** coverage data is uploaded to Codecov without requiring a token

#### Scenario: Codecov upload failure does not block CI
- **WHEN** the Codecov upload fails (service unreachable)
- **THEN** the overall CI workflow still succeeds

### Requirement: Dependabot is configured for dependency updates
The project SHALL include a `.github/dependabot.yml` configuration that monitors two ecosystems: `gomodules` (Go dependencies) and `github-actions` (workflow action versions). Dependabot SHALL create PRs for updates on a weekly schedule.

#### Scenario: Go dependency update PR is created
- **WHEN** a new version of a Go dependency is available
- **THEN** Dependabot creates a PR updating `go.mod` and `go.sum`

#### Scenario: GitHub Actions version update PR is created
- **WHEN** a new version of a GitHub Action used in workflows is available
- **THEN** Dependabot creates a PR updating the action version

### Requirement: Branch protection rules are configured on main
The `main` branch SHALL have branch protection rules requiring: all CI status checks (lint, test, build, commitlint) to pass before merge, no force pushes, and no direct pushes (all changes via PR). Administrators SHALL NOT be exempt from these rules.

#### Scenario: PR with failing CI cannot be merged
- **WHEN** a PR has failing CI status checks
- **THEN** the merge button is blocked on GitHub

#### Scenario: Direct push to main is rejected
- **WHEN** a developer attempts to push directly to main
- **THEN** the push is rejected by branch protection

#### Scenario: Force push to main is rejected
- **WHEN** a developer attempts to force push to main
- **THEN** the push is rejected by branch protection
