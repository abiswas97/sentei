# Design: completion-flash

`renderProgressLayout` already re-applies bar colors per render; the branch
keys off `l.Completed`: accent gradient while working, `barDoneStart→barDoneEnd`
(hex twins of success green) once the result arrives. The integration flow
never set `Completed`; it gains `integ.finalized`, set at
`integrationFinalizedMsg` and cleared when a new apply starts. The audit
sketch said "flash (~400ms accent→success)"; settle-to-green was chosen
instead — a revert reads as a glitch, and green-stays matches the
✦-crystallizes story.
