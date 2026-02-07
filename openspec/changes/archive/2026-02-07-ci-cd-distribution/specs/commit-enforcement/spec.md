## ADDED Requirements

### Requirement: Local commit-msg hook validates conventional commit format
An installable git `commit-msg` hook SHALL validate that the commit message follows conventional commit format. The hook SHALL accept messages matching the pattern `<type>[optional scope][optional !]: <description>` where type is one of: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert. The hook SHALL reject non-conforming messages with a helpful error showing the expected format.

#### Scenario: Valid conventional commit is accepted
- **WHEN** a developer commits with message `feat: add new feature`
- **THEN** the commit-msg hook passes and the commit succeeds

#### Scenario: Valid scoped commit is accepted
- **WHEN** a developer commits with message `fix(parser): handle empty input`
- **THEN** the commit-msg hook passes and the commit succeeds

#### Scenario: Valid breaking change commit is accepted
- **WHEN** a developer commits with message `feat!: change CLI args`
- **THEN** the commit-msg hook passes and the commit succeeds

#### Scenario: Invalid commit message is rejected
- **WHEN** a developer commits with message `updated the thing`
- **THEN** the commit-msg hook fails with an error showing the expected format

#### Scenario: Merge commits are allowed
- **WHEN** a developer commits with message `Merge branch 'feature' into main`
- **THEN** the commit-msg hook passes (merge commits are exempt)

### Requirement: Hook install script exists
A `scripts/install-hooks.sh` script SHALL install the commit-msg hook by symlinking or copying it to `.git/hooks/commit-msg`. The script SHALL be idempotent (safe to run multiple times). The script SHALL make the hook executable.

#### Scenario: Running install script sets up the hook
- **WHEN** a developer runs `./scripts/install-hooks.sh`
- **THEN** `.git/hooks/commit-msg` exists and is executable

#### Scenario: Running install script twice is safe
- **WHEN** a developer runs `./scripts/install-hooks.sh` twice
- **THEN** no error occurs and the hook is still functional

### Requirement: CI validates conventional commits on PRs
A GitHub Actions job SHALL validate that all commit messages in pull requests follow the conventional commit format using `wagoid/commitlint-github-action`. This job runs as part of the CI workflow on PRs targeting `main`.

#### Scenario: PR with valid commits passes
- **WHEN** a PR targeting `main` has all conventional commit messages
- **THEN** the commitlint check passes

#### Scenario: PR with invalid commits fails
- **WHEN** a PR targeting `main` has a commit with message `updated stuff`
- **THEN** the commitlint check fails with details about which commit is invalid
