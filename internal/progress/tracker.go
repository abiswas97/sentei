package progress

import "strings"

// StepStatus represents the current state of a step.
type StepStatus int

const (
	Pending StepStatus = iota
	Running
	Done
	Failed
)

// Step holds the display name and current status of a single step.
type Step struct {
	Name   string
	Status StepStatus
	Error  string
}

// Group is a named collection of steps (e.g., a phase or worktree).
type Group struct {
	Name  string
	Steps []Step
}

// Tracker tracks progress across groups of steps. It is initialized with
// a known total so the progress bar shows the correct denominator from
// the start, even before all steps have emitted events.
type Tracker struct {
	groups     []*Group
	groupIndex map[string]int
	stepIndex  map[string]int // "group:step" → index in group.Steps
	total      int
}

// New creates a Tracker with a known total step count.
// Use this when the total is calculable before execution starts.
func New(total int) *Tracker {
	return &Tracker{
		groupIndex: make(map[string]int),
		stepIndex:  make(map[string]int),
		total:      total,
	}
}

// NewFromGroups creates a Tracker from pre-declared groups and steps.
// Total is calculated automatically.
func NewFromGroups(groups []Group) *Tracker {
	t := &Tracker{
		groupIndex: make(map[string]int),
		stepIndex:  make(map[string]int),
	}
	for _, g := range groups {
		gCopy := g
		t.groupIndex[g.Name] = len(t.groups)
		for i, s := range gCopy.Steps {
			key := g.Name + ":" + s.Name
			t.stepIndex[key] = i
		}
		t.total += len(gCopy.Steps)
		t.groups = append(t.groups, &gCopy)
	}
	return t
}

// Update records a step status change. If the group or step doesn't exist
// yet, it is created dynamically (for event-driven progress where steps
// are discovered at runtime).
func (t *Tracker) Update(group, step string, status StepStatus, errMsg string) {
	key := group + ":" + step

	g := t.getOrCreateGroup(group)

	if idx, ok := t.stepIndex[key]; ok {
		g.Steps[idx].Status = status
		g.Steps[idx].Error = errMsg
		return
	}

	t.stepIndex[key] = len(g.Steps)
	g.Steps = append(g.Steps, Step{Name: step, Status: status, Error: errMsg})
}

// Done returns the count of completed steps (Done or Failed).
func (t *Tracker) Done() int {
	done := 0
	for _, g := range t.groups {
		for _, s := range g.Steps {
			if s.Status == Done || s.Status == Failed {
				done++
			}
		}
	}
	return done
}

// Total returns the known total step count. If steps were added dynamically
// beyond the initial total, it returns the larger of the two.
func (t *Tracker) Total() int {
	actual := 0
	for _, g := range t.groups {
		actual += len(g.Steps)
	}
	if actual > t.total {
		return actual
	}
	return t.total
}

// Groups returns the current groups with their steps, in insertion order.
func (t *Tracker) Groups() []Group {
	result := make([]Group, len(t.groups))
	for i, g := range t.groups {
		result[i] = *g
	}
	return result
}

// IsComplete returns true when all steps have terminal status (Done or Failed).
func (t *Tracker) IsComplete() bool {
	return t.Done() >= t.Total() && t.Total() > 0
}

// Percent returns the completion percentage (0-100).
func (t *Tracker) Percent() int {
	total := t.Total()
	if total == 0 {
		return 0
	}
	return (t.Done() * 100) / total
}

// FailedCount returns the number of steps with Failed status.
func (t *Tracker) FailedCount() int {
	count := 0
	for _, g := range t.groups {
		for _, s := range g.Steps {
			if s.Status == Failed {
				count++
			}
		}
	}
	return count
}

// Bar renders a progress bar string: ████░░░░ 5/9
func (t *Tracker) Bar(width int) string {
	total := t.Total()
	done := t.Done()

	filled := 0
	if total > 0 {
		filled = (done * width) / total
	}
	if filled > width {
		filled = width
	}

	var b strings.Builder
	for range filled {
		b.WriteRune('\u2588') // █
	}
	for range width - filled {
		b.WriteRune('\u2591') // ░
	}

	return b.String()
}

func (t *Tracker) getOrCreateGroup(name string) *Group {
	if idx, ok := t.groupIndex[name]; ok {
		return t.groups[idx]
	}
	g := &Group{Name: name}
	t.groupIndex[name] = len(t.groups)
	t.groups = append(t.groups, g)
	return g
}
