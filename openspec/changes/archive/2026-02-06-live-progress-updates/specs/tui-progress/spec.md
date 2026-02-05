## MODIFIED Requirements

### Requirement: Display deletion progress
The TUI SHALL display a progress view during deletion showing a progress bar, percentage, and per-worktree status, updating in real-time as each deletion event is received from the deletion channel.

#### Scenario: Progress bar updates incrementally
- **WHEN** a worktree deletion completes (success or failure) and 2 out of 5 total have now finished
- **THEN** the progress bar SHALL immediately re-render showing 40% completion with "2/5" label

#### Scenario: Per-worktree status on deletion start
- **WHEN** a `DeletionStarted` event is received for a worktree
- **THEN** the TUI SHALL show "removing..." next to that worktree

#### Scenario: Per-worktree status on success
- **WHEN** a `DeletionCompleted` event is received for a worktree
- **THEN** the TUI SHALL show a checkmark and "removed" next to that worktree

#### Scenario: Per-worktree status on failure
- **WHEN** a `DeletionFailed` event is received for a worktree
- **THEN** the TUI SHALL show an error indicator and "failed" next to that worktree

#### Scenario: Events forwarded as Bubble Tea messages
- **WHEN** the deletion goroutine sends events over the progress channel
- **THEN** each event SHALL be delivered to the Bubble Tea Update loop as an individual message, not batched

## ADDED Requirements

### Requirement: Cmd-chained event consumption
The TUI SHALL consume deletion progress events using a Cmd-chaining pattern where each Cmd reads one event from the channel and returns it as a Msg, and the Update handler returns a new Cmd to read the next event.

#### Scenario: Channel stored on model
- **WHEN** deletion starts
- **THEN** the progress event channel SHALL be stored on the Model so it persists across Update calls

#### Scenario: Sequential event delivery
- **WHEN** multiple deletion events arrive on the channel
- **THEN** each event SHALL be read by a separate Cmd invocation and delivered as a separate Msg to Update

#### Scenario: Channel closure triggers completion
- **WHEN** the progress channel is closed (all deletions finished)
- **THEN** the wait Cmd SHALL return a completion message and the TUI SHALL transition to the summary view
