## ADDED Requirements

### Requirement: Makefile provides local development commands
The project SHALL include a `Makefile` at the repo root with these targets: `build` (compile binary), `test` (run tests with race detector), `lint` (run golangci-lint), `snapshot` (run goreleaser build --snapshot --clean for local testing), `install-hooks` (run scripts/install-hooks.sh). Each target SHALL work without arguments.

#### Scenario: make build compiles the binary
- **WHEN** a developer runs `make build`
- **THEN** the `sentei` binary is compiled in the project root

#### Scenario: make test runs tests
- **WHEN** a developer runs `make test`
- **THEN** `go test -race ./...` executes

#### Scenario: make lint runs linter
- **WHEN** a developer runs `make lint`
- **THEN** `golangci-lint run` executes using `.golangci.yml`

#### Scenario: make snapshot builds without publishing
- **WHEN** a developer runs `make snapshot`
- **THEN** `goreleaser build --snapshot --clean` runs and outputs to `dist/`

#### Scenario: make install-hooks sets up git hooks
- **WHEN** a developer runs `make install-hooks`
- **THEN** the commit-msg hook is installed via `scripts/install-hooks.sh`

### Requirement: GoReleaser output directory is gitignored
The `.gitignore` file SHALL include `dist/` to prevent GoReleaser snapshot output from being committed.

#### Scenario: dist directory is ignored by git
- **WHEN** a developer runs `make snapshot` which creates `dist/`
- **THEN** `git status` does not show `dist/` as untracked

### Requirement: MIT LICENSE file exists
The project SHALL include a `LICENSE` file at the repo root containing the MIT license text. This is required for Homebrew distribution.

#### Scenario: LICENSE file is present
- **WHEN** the repo is checked
- **THEN** a `LICENSE` file exists at the repo root with MIT license text

### Requirement: README includes Codecov badge
The `README.md` SHALL include a Codecov coverage badge that links to the project's Codecov dashboard. The badge SHALL display the current coverage percentage.

#### Scenario: Codecov badge is visible in README
- **WHEN** a user views the README on GitHub
- **THEN** a coverage badge is displayed showing the current coverage percentage
