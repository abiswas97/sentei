## ADDED Requirements

### Requirement: Bar fill derives from checkpoints
The overall progress bar's fill target SHALL derive from checkpoints reached over checkpoints declared across all phases (equal checkpoint weights), while phase headers SHALL continue to count steps, so the bar gains resolution from declared sub-stages without changing the human-readable counts.

#### Scenario: Bar moves during parallel work
- **WHEN** three parallel removals have started (start checkpoints reached) and none has completed
- **THEN** the overall bar's fill target is greater than zero while every phase header still reads 0 of its step total done

#### Scenario: Bar never overstates
- **WHEN** all checkpoints of all declared steps are reached
- **THEN** the bar's fill target is exactly 100%, and at no prior moment does it exceed the reached/declared ratio

### Requirement: Done styling requires settled phases
Phase collapse, the done indicator on a phase headline, and success styling SHALL render only for phases whose settled predicate is true (closed and fully resolved), and a settled phase SHALL never return to an in-progress rendering.

#### Scenario: No premature done styling
- **WHEN** a phase's done count equals its currently known total but more declared work exists or the phase is open
- **THEN** the phase renders with in-progress treatment, not the done indicator

#### Scenario: Settled phases stay settled
- **WHEN** a phase has rendered as settled
- **THEN** no subsequent event causes that phase to render in-progress again in any later frame
