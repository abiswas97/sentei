# integration-apply-summary Delta

## ADDED Requirements

### Requirement: Presence-based tool detection
The default integration detect probe SHALL test PATH presence (`command -v <binary>`) rather than invoking the tool with a flag it may not implement. An integration MAY still declare an explicit `Detect.Command` that overrides the default. A tool already installed by any manager (pipx, uv, brew, manual) SHALL be detected and its install steps skipped.

#### Scenario: Existing install is detected regardless of CLI flags
- **WHEN** the integration's binary is on PATH but does not implement `--version`
- **THEN** detection SHALL succeed and the install and dependency steps SHALL be skipped

#### Scenario: Missing tool still installs
- **WHEN** the binary is not on PATH
- **THEN** detection SHALL fail and the install step SHALL run

### Requirement: Bounded failure rendering
A failed step's error SHALL render as a bounded peek, never raw: in summaries, at most three lines — the error's first line (dim), its last non-empty line (error color), and a dim `… N more — ? for full output` marker when lines were elided — each truncated to the view width. In live progress rows the error SHALL clamp to a single truncated line. The complete untrimmed output SHALL be available in the detail portal.

#### Scenario: Multi-hundred-line tool output stays bounded
- **WHEN** a failed step's error contains an installer's full output
- **THEN** the summary SHALL show at most three lines for that step and the chrome SHALL remain intact

#### Scenario: Full output preserved
- **WHEN** the user opens the detail portal from a summary with failures
- **THEN** the failed step's complete error output SHALL be readable there
