# Design: p3-polish

Mechanical, behavior-narrow fixes. The one judgment call: the sort arrow.
The underlying sort is by commit date; the column shows ages. Arrow-follows-
data would point up for date-ascending while the visible ages descend —
the audit's confusion. Rule adopted: the arrow describes the displayed
values' order. Implemented as an Age-column flip; contract-tested. The
GoldenListView regeneration is this intentional flip and nothing else.
