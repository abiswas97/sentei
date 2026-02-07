### Requirement: Version flag prints version and exits
The CLI SHALL accept a `--version` flag. When provided, it SHALL print the version string to stdout in the format `sentei <version> (<commit>, <date>)` and exit with code 0. The commit SHALL be the short (7-char) git commit SHA. The date SHALL be the build date in `YYYY-MM-DD` format. No other output SHALL be produced.

#### Scenario: User passes --version
- **WHEN** user runs `sentei --version`
- **THEN** stdout contains `sentei v0.1.0 (abc1234, 2026-02-07)` and the process exits with code 0

#### Scenario: Version not set at build time
- **WHEN** the binary is built without `-ldflags` version injection
- **THEN** `sentei --version` prints `sentei dev (none, unknown)`

### Requirement: Version flag takes precedence over other flags
When `--version` is combined with other flags, the version output SHALL take precedence and the tool SHALL exit immediately without running any other logic.

#### Scenario: --version combined with --dry-run
- **WHEN** user runs `sentei --version --dry-run`
- **THEN** only the version string is printed and the process exits with code 0

#### Scenario: --version combined with --playground
- **WHEN** user runs `sentei --version --playground`
- **THEN** only the version string is printed and no playground is created
