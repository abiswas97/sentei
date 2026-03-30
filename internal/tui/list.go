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

const (
	colCursor   = 0
	colCheckbox = 1
	colStatus   = 2
	colBranch   = 3
	colAge      = 4
	colSubject  = 5
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
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyMsg:
		if m.remove.filterActive {
			return m.updateFilterInput(msg)
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Back):
			if m.remove.filterText != "" {
				m.remove.filterText = ""
				m.reindex()
				return m, nil
			}
			if m.menuItems != nil {
				m.view = menuView
				return m, nil
			}
			return m, tea.Quit

		case key.Matches(msg, keys.Filter):
			m.remove.filterActive = true
			m.remove.filterInput.SetValue(m.remove.filterText)
			m.remove.filterInput.Focus()
			return m, m.remove.filterInput.Cursor.BlinkCmd()

		case key.Matches(msg, keys.Sort):
			m.remove.sortField = (m.remove.sortField + 1) % 2
			m.remove.cursor = 0
			m.remove.offset = 0
			m.reindex()

		case key.Matches(msg, keys.ReverseSort):
			m.remove.sortAscending = !m.remove.sortAscending
			m.remove.cursor = 0
			m.remove.offset = 0
			m.reindex()

		case key.Matches(msg, keys.Down):
			if m.remove.cursor < len(m.remove.visibleIndices)-1 {
				m.remove.cursor++
				if m.remove.cursor >= m.remove.offset+m.height {
					m.remove.offset = m.remove.cursor - m.height + 1
				}
			}

		case key.Matches(msg, keys.Up):
			if m.remove.cursor > 0 {
				m.remove.cursor--
				if m.remove.cursor < m.remove.offset {
					m.remove.offset = m.remove.cursor
				}
			}

		case key.Matches(msg, keys.PageDown):
			m.remove.cursor += m.height
			if m.remove.cursor >= len(m.remove.visibleIndices) {
				m.remove.cursor = len(m.remove.visibleIndices) - 1
			}
			if m.remove.cursor < 0 {
				m.remove.cursor = 0
			}
			if m.remove.cursor >= m.remove.offset+m.height {
				m.remove.offset = m.remove.cursor - m.height + 1
			}

		case key.Matches(msg, keys.PageUp):
			m.remove.cursor -= m.height
			if m.remove.cursor < 0 {
				m.remove.cursor = 0
			}
			if m.remove.cursor < m.remove.offset {
				m.remove.offset = m.remove.cursor
			}

		case key.Matches(msg, keys.Toggle):
			if len(m.remove.visibleIndices) > 0 {
				wt := m.remove.worktrees[m.remove.visibleIndices[m.remove.cursor]]
				if git.IsProtectedBranch(wt.Branch) {
					break
				}
				if m.remove.selected[wt.Path] {
					delete(m.remove.selected, wt.Path)
				} else {
					m.remove.selected[wt.Path] = true
				}
			}

		case key.Matches(msg, keys.All):
			allSelected := true
			for _, idx := range m.remove.visibleIndices {
				wt := m.remove.worktrees[idx]
				if git.IsProtectedBranch(wt.Branch) {
					continue
				}
				if !m.remove.selected[wt.Path] {
					allSelected = false
					break
				}
			}
			if allSelected {
				for _, idx := range m.remove.visibleIndices {
					wt := m.remove.worktrees[idx]
					if git.IsProtectedBranch(wt.Branch) {
						continue
					}
					delete(m.remove.selected, wt.Path)
				}
			} else {
				for _, idx := range m.remove.visibleIndices {
					wt := m.remove.worktrees[idx]
					if git.IsProtectedBranch(wt.Branch) {
						continue
					}
					m.remove.selected[wt.Path] = true
				}
			}

		case key.Matches(msg, keys.Confirm):
			if len(m.remove.selected) > 0 {
				m.view = confirmView
			}
		}
	}
	return m, nil
}

func (m Model) updateFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.remove.filterActive = false
		m.remove.filterText = ""
		m.remove.filterInput.SetValue("")
		m.remove.filterInput.Blur()
		m.reindex()
		return m, nil

	case key.Matches(msg, keys.Confirm):
		m.remove.filterActive = false
		m.remove.filterText = m.remove.filterInput.Value()
		m.remove.filterInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.remove.filterInput, cmd = m.remove.filterInput.Update(msg)
	m.remove.filterText = m.remove.filterInput.Value()
	m.reindex()
	return m, cmd
}

func (m Model) viewList() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("  sentei \u2500 Remove Worktrees"))
	b.WriteString("\n\n")

	if len(m.remove.worktrees) == 0 {
		b.WriteString(styleDim.Render("  No worktrees found."))
		b.WriteString("\n")
		return b.String()
	}

	if len(m.remove.visibleIndices) == 0 {
		b.WriteString(styleDim.Render("  No matches."))
		b.WriteString("\n\n")
		b.WriteString(m.viewStatusOrFilter())
		b.WriteString("\n")
		b.WriteString(m.viewBottomLine())
		return b.String()
	}

	end := min(m.remove.offset+m.height, len(m.remove.visibleIndices))

	arrow := " ▲"
	if !m.remove.sortAscending {
		arrow = " ▼"
	}
	hdrBranch := "Branch"
	hdrAge := "Age"
	hdrSubject := "Subject"
	switch m.remove.sortField {
	case SortByBranch:
		hdrBranch += arrow
	case SortByAge:
		hdrAge += arrow
	}

	t := table.New().
		BorderTop(false).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderColumn(false).BorderHeader(false).BorderRow(false).
		Headers("", "", "", hdrBranch, hdrAge, hdrSubject).
		Wrap(true)

	if m.width > 0 {
		t.Width(m.width)
	}

	fixedWidth := colWidthCursor + colWidthCheckbox + colWidthStatus + colWidthAge
	colPadding := 3
	remaining := max(m.width-fixedWidth-colPadding, 20)
	branchWidth := remaining / 2
	subjectWidth := remaining - branchWidth

	for i := m.remove.offset; i < end; i++ {
		wt := m.remove.worktrees[m.remove.visibleIndices[i]]

		cursor := "  "
		if i == m.remove.cursor {
			cursor = "> "
		}

		var checkbox string
		if git.IsProtectedBranch(wt.Branch) {
			checkbox = styleStatusProtected.Render("[P]")
		} else if m.remove.selected[wt.Path] {
			checkbox = "[x]"
		} else {
			checkbox = "[ ]"
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

		maxSubject := subjectWidth - 2
		if maxSubject > 3 && lipgloss.Width(subject) > maxSubject {
			runes := []rune(subject)
			if len(runes) > maxSubject-3 {
				subject = string(runes[:maxSubject-3]) + "..."
			}
		}

		t.Row(cursor, checkbox, status, branch, age, subject)
	}

	sortedCol := colAge
	if m.remove.sortField == SortByBranch {
		sortedCol = colBranch
	}

	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			base := styleColumnHeader
			if col == sortedCol {
				base = styleColumnHeaderSorted
			}
			return columnStyle(base, col, branchWidth, subjectWidth)
		}

		idx := m.remove.offset + row

		var base lipgloss.Style
		switch {
		case idx == m.remove.cursor:
			base = styleCursorRow
		case m.remove.selected[m.remove.worktrees[m.remove.visibleIndices[idx]].Path]:
			base = styleSelectedRow
		default:
			base = styleNormalRow
		}

		return columnStyle(base, col, branchWidth, subjectWidth)
	})

	b.WriteString(t.Render())
	b.WriteString("\n")

	b.WriteString(m.viewStatusOrFilter())
	b.WriteString("\n")
	b.WriteString(m.viewBottomLine())
	return b.String()
}

func (m Model) viewStatusOrFilter() string {
	if m.remove.filterActive {
		return m.remove.filterInput.View()
	}
	return m.viewStatusBar()
}

func (m Model) viewBottomLine() string {
	if m.remove.filterActive {
		return styleDim.Render("  enter apply \u00b7 esc cancel")
	}
	return m.viewLegend()
}

func (m Model) viewStatusBar() string {
	count := len(m.remove.selected)

	var filterInfo string
	if m.remove.filterText != "" {
		filterInfo = fmt.Sprintf(" \u00b7 filter: %q (%d/%d)", m.remove.filterText, len(m.remove.visibleIndices), len(m.remove.worktrees))
	}

	return styleStatusBar.Render(
		fmt.Sprintf("  %d selected%s \u00b7 space toggle \u00b7 a all \u00b7 enter delete \u00b7 / filter \u00b7 s sort \u00b7 q quit", count, filterInfo),
	)
}

func (m Model) viewLegend() string {
	return styleDim.Render("  ") +
		styleStatusClean.Render("[ok]") + styleDim.Render(" clean  ") +
		styleStatusDirty.Render("[~]") + styleDim.Render(" dirty  ") +
		styleStatusUntracked.Render("[!]") + styleDim.Render(" untracked  ") +
		styleStatusLocked.Render("[L]") + styleDim.Render(" locked  ") +
		styleStatusProtected.Render("[P]") + styleDim.Render(" protected")
}

func columnStyle(base lipgloss.Style, col, branchWidth, subjectWidth int) lipgloss.Style {
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
}
