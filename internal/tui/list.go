package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abiswas/wt-sweep/internal/git"
)

const (
	colWidthBranch  = 30
	colWidthAge     = 16
	colWidthSubject = 40
)

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

func stripBranchPrefix(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

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
		m.height = msg.Height - 4
		if m.height < 5 {
			m.height = 5
		}
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

	b.WriteString(styleHeader.Render("wt-sweep - Git Worktree Cleanup"))
	b.WriteString("\n\n")

	if len(m.worktrees) == 0 {
		b.WriteString(styleDim.Render("  No worktrees found."))
		b.WriteString("\n")
		return b.String()
	}

	end := m.offset + m.height
	if end > len(m.worktrees) {
		end = len(m.worktrees)
	}

	for i := m.offset; i < end; i++ {
		wt := m.worktrees[i]

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		branch := stripBranchPrefix(wt.Branch)
		if branch == "" && wt.IsDetached {
			branch = wt.HEAD[:7]
		} else if branch == "" && wt.IsPrunable {
			branch = "(prunable)"
		}

		status := statusIndicator(wt)

		age := relativeTime(wt.LastCommitDate)
		subject := wt.LastCommitSubject
		if wt.EnrichmentError != "" {
			age = "error"
			subject = wt.EnrichmentError
		}

		branchCol := fmt.Sprintf("%-*s", colWidthBranch, branch)
		ageCol := fmt.Sprintf("%-*s", colWidthAge, age)

		if len(subject) > colWidthSubject {
			subject = subject[:colWidthSubject-3] + "..."
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			cursor, checkbox, " ", status, " ",
			branchCol, " ", ageCol, " ", subject,
		)

		if i == m.cursor {
			row = styleCursorRow.Render(row)
		} else if m.selected[i] {
			row = styleSelectedRow.Render(row)
		} else {
			row = styleNormalRow.Render(row)
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString(m.viewStatusBar())
	return b.String()
}

func (m Model) viewStatusBar() string {
	count := len(m.selected)
	return styleStatusBar.Render(
		fmt.Sprintf("  %d selected | space: toggle | a: all | enter: delete | q: quit", count),
	)
}
