package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/abiswas97/sentei/internal/git"
)

// Column indices for the worktree list table. These map to the
// table.Row() argument order and the StyleFunc col parameter.
const (
	colCursor   = 0
	colCheckbox = 1
	colStatus   = 2
	colBranch   = 3
	colAge      = 4
	colSubject  = 5
)

// relativeTime formats a timestamp as a human-readable relative duration
// (e.g. "3 days ago", "just now"). Returns "unknown" for zero-value times.
func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// stripBranchPrefix removes the "refs/heads/" prefix from a git branch reference.
func stripBranchPrefix(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

// statusIndicator returns a colored status badge for a worktree.
// Priority: locked [L] > dirty [~] > untracked [!] > clean [ok].
func statusIndicator(wt git.Worktree) string {
	switch {
	case wt.IsLocked:
		return styleStatusLocked.Render("[L]")
	case wt.HasUncommittedChanges:
		return styleStatusDirty.Render("[~]")
	case wt.HasUntrackedFiles:
		return styleStatusUntracked.Render("[!]")
	default:
		return styleStatusClean.Render("[ok]")
	}
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-4, 5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.worktrees)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.height {
					m.offset = m.cursor - m.height + 1
				}
			}

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}

		case key.Matches(msg, keys.PageDown):
			m.cursor += m.height
			if m.cursor >= len(m.worktrees) {
				m.cursor = len(m.worktrees) - 1
			}
			if m.cursor >= m.offset+m.height {
				m.offset = m.cursor - m.height + 1
			}

		case key.Matches(msg, keys.PageUp):
			m.cursor -= m.height
			if m.cursor < 0 {
				m.cursor = 0
			}
			if m.cursor < m.offset {
				m.offset = m.cursor
			}

		case key.Matches(msg, keys.Toggle):
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}

		case key.Matches(msg, keys.All):
			if len(m.selected) == len(m.worktrees) {
				m.selected = make(map[int]bool)
			} else {
				for i := range m.worktrees {
					m.selected[i] = true
				}
			}

		case key.Matches(msg, keys.Confirm):
			if len(m.selected) > 0 {
				m.view = confirmView
			}
		}
	}
	return m, nil
}

func (m Model) viewList() string {
	var b strings.Builder

	b.WriteString(styleHeader.Render("sentei - Git Worktree Cleanup"))
	b.WriteString("\n\n")

	if len(m.worktrees) == 0 {
		b.WriteString(styleDim.Render("  No worktrees found."))
		b.WriteString("\n")
		return b.String()
	}

	end := min(m.offset+m.height, len(m.worktrees))

	t := table.New().
		BorderTop(false).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderColumn(false).BorderHeader(false).BorderRow(false).
		Wrap(true)

	if m.width > 0 {
		t.Width(m.width)
	}

	fixedWidth := colWidthCursor + colWidthCheckbox + colWidthStatus + colWidthAge
	// 3 data columns (branch, age, subject) each have Padding(0,1) which eats
	// 1 char of their Width budget. Subtract here so the proportional split
	// accounts for it and the total doesn't overshoot the terminal width.
	colPadding := 3
	remaining := max(m.width-fixedWidth-colPadding, 20)
	branchWidth := remaining / 2
	subjectWidth := remaining - branchWidth

	for i := m.offset; i < end; i++ {
		wt := m.worktrees[i]

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}

		status := statusIndicator(wt)

		branch := stripBranchPrefix(wt.Branch)
		if branch == "" {
			switch {
			case wt.IsDetached:
				branch = wt.HEAD
				if len(branch) >= 7 {
					branch = branch[:7]
				}
			case wt.IsPrunable:
				branch = "(prunable)"
			}
		}

		age := relativeTime(wt.LastCommitDate)
		subject := wt.LastCommitSubject
		if wt.EnrichmentError != "" {
			age = "error"
			subject = wt.EnrichmentError
		}

		// Pre-truncate subject so it never wraps. Only branch names wrap.
		// -2 accounts for Padding(0,1) inside the column plus a safety char.
		maxSubject := subjectWidth - 2
		if maxSubject > 3 && lipgloss.Width(subject) > maxSubject {
			runes := []rune(subject)
			if len(runes) > maxSubject-3 {
				subject = string(runes[:maxSubject-3]) + "..."
			}
		}

		t.Row(cursor, checkbox, status, branch, age, subject)
	}

	t.StyleFunc(func(row, col int) lipgloss.Style {
		idx := m.offset + row

		var base lipgloss.Style
		switch {
		case idx == m.cursor:
			base = styleCursorRow
		case m.selected[idx]:
			base = styleSelectedRow
		default:
			base = styleNormalRow
		}

		switch col {
		case colCursor:
			return base.Width(colWidthCursor)
		case colCheckbox:
			return base.Width(colWidthCheckbox)
		case colStatus:
			return base.Width(colWidthStatus)
		case colBranch:
			return base.Width(branchWidth).Padding(0, 1)
		case colAge:
			return base.Width(colWidthAge).Padding(0, 1)
		case colSubject:
			return base.Width(subjectWidth).Padding(0, 1)
		default:
			return base
		}
	})

	b.WriteString(t.Render())
	b.WriteString("\n")

	b.WriteString(m.viewStatusBar())
	return b.String()
}

func (m Model) viewStatusBar() string {
	count := len(m.selected)
	return styleStatusBar.Render(
		fmt.Sprintf("  %d selected | space: toggle | a: all | enter: delete | q: quit", count),
	)
}
