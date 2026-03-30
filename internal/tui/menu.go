package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/git"
	"github.com/abiswas97/sentei/internal/repo"
	"github.com/abiswas97/sentei/internal/worktree"
)

type worktreeContextMsg struct {
	worktrees []git.Worktree
	err       error
}

func loadWorktreeContext(runner git.CommandRunner, repoPath string) tea.Cmd {
	return func() tea.Msg {
		wts, err := git.ListWorktrees(runner, repoPath)
		if err != nil {
			return worktreeContextMsg{err: err}
		}
		wts = worktree.EnrichWorktrees(runner, wts, 10)
		var filtered []git.Worktree
		for _, wt := range wts {
			if !wt.IsBare {
				filtered = append(filtered, wt)
			}
		}
		return worktreeContextMsg{worktrees: filtered}
	}
}

func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case worktreeContextMsg:
		if msg.err == nil {
			m.remove.worktrees = msg.worktrees
			m.reindex()
			m.updateMenuHints()
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Back):
			return m, tea.Quit

		case key.Matches(msg, keys.Down):
			for {
				m.menuCursor++
				if m.menuCursor >= len(m.menuItems) {
					m.menuCursor = len(m.menuItems) - 1
					break
				}
				if m.menuItems[m.menuCursor].enabled {
					break
				}
			}

		case key.Matches(msg, keys.Up):
			for {
				m.menuCursor--
				if m.menuCursor < 0 {
					m.menuCursor = 0
					break
				}
				if m.menuItems[m.menuCursor].enabled {
					break
				}
			}

		case key.Matches(msg, keys.Confirm):
			if m.menuCursor >= 0 && m.menuCursor < len(m.menuItems) && m.menuItems[m.menuCursor].enabled {
				label := m.menuItems[m.menuCursor].label
				switch label {
				case "Create new worktree":
					m.view = createBranchView
					return m, m.create.branchInput.Cursor.BlinkCmd()
				case "Manage integrations":
					m.view = integrationListView
					return m, m.loadIntegrationState()
				case "Remove worktrees":
					m.view = listView
					if len(m.remove.worktrees) == 0 {
						return m, loadWorktreeContext(m.runner, m.repoPath)
					}
				case "Cleanup & exit":
					m.view = cleanupResultView
					m.remove.cleanupResult = nil
					return m, runStandaloneCleanup(m.runner, m.repoPath)
				case "Create new repository":
					m.repo.nameInput.SetValue("")
					m.repo.locationInput.SetValue(m.repoPath)
					m.repo.focusedField = 0
					m.repo.validationErr = ""
					m.view = repoNameView
					return m, m.repo.nameInput.Focus()
				case "Clone repository as bare":
					m.repo.urlInput.SetValue("")
					m.repo.cloneNameInput.SetValue("")
					m.repo.cloneFocusedField = 0
					m.repo.nameManuallyEdited = false
					m.view = cloneInputView
					return m, m.repo.urlInput.Focus()
				case "Migrate to bare repository":
					m.view = migrateConfirmView
					return m, loadMigrateInfo(m.runner, m.repoPath)
				}
			}
		}
	}
	return m, nil
}

func (m *Model) updateMenuHints() {
	if m.context != repo.ContextBareRepo {
		return
	}
	if len(m.menuItems) < 3 {
		return
	}
	count := len(m.remove.worktrees)
	if count > 0 {
		m.menuItems[2].hint = fmt.Sprintf("%d available", count)
		m.menuItems[2].enabled = true
	} else {
		m.menuItems[2].hint = "none"
		m.menuItems[2].enabled = false
	}
}

func (m Model) viewMenu() string {
	var b strings.Builder

	repoName := filepath.Base(m.repoPath)
	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Git Worktree Manager", "\u2500")))
	b.WriteString("\n\n")

	switch m.context {
	case repo.ContextBareRepo:
		b.WriteString(styleDim.Render(fmt.Sprintf("  %s (bare) %s %s", repoName, "\u00b7", m.repoPath)))
		b.WriteString("\n")
		if len(m.remove.worktrees) > 0 {
			clean, dirty, locked := 0, 0, 0
			for _, wt := range m.remove.worktrees {
				switch {
				case wt.IsLocked:
					locked++
				case wt.HasUncommittedChanges || wt.HasUntrackedFiles:
					dirty++
				default:
					clean++
				}
			}
			b.WriteString(styleDim.Render(fmt.Sprintf("  %d worktrees %s %d clean, %d dirty, %d locked",
				len(m.remove.worktrees), "\u00b7", clean, dirty, locked)))
			b.WriteString("\n")
		}
	case repo.ContextNonBareRepo:
		b.WriteString(styleDim.Render(fmt.Sprintf("  %s %s %s", repoName, "\u00b7", m.repoPath)))
		b.WriteString("\n")
	case repo.ContextNoRepo:
		b.WriteString(styleDim.Render(fmt.Sprintf("  %s", m.repoPath)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	for i, item := range m.menuItems {
		cursor := "  "
		if i == m.menuCursor {
			cursor = "> "
		}

		label := item.label
		if !item.enabled {
			label = styleDim.Render(label)
		}

		hint := ""
		if item.hint != "" {
			hint = "  " + styleDim.Render(item.hint)
		}

		if i == m.menuCursor {
			b.WriteString(styleAccent.Render(cursor) + label + hint)
		} else {
			b.WriteString("  " + label + hint)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleDim.Render("  j/k navigate \u00b7 enter select \u00b7 q quit"))
	b.WriteString("\n")

	return b.String()
}
