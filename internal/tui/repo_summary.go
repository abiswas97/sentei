package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/repo"
)

func (m Model) updateRepoSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = max(msg.Height-6, 5)
		return m, nil

	case tea.KeyMsg:
		var repoPath string
		switch r := m.repo.result.(type) {
		case repo.CreateResult:
			repoPath = r.RepoPath
		case repo.CloneResult:
			repoPath = r.RepoPath
		}

		switch {
		case key.Matches(msg, keys.Confirm):
			if repoPath != "" {
				return m, relaunchSenteiAt(repoPath)
			}
			return m, tea.Quit
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func relaunchSenteiAt(repoPath string) tea.Cmd {
	senteiPath, err := os.Executable()
	if err != nil {
		senteiPath = "sentei"
	}
	c := exec.Command(senteiPath, repoPath)
	c.Env = os.Environ()
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return tea.Quit()
	})
}

func (m Model) viewRepoSummary() string {
	var b strings.Builder

	switch result := m.repo.result.(type) {
	case repo.CreateResult:
		return m.viewCreateRepoSummary(result)
	case repo.CloneResult:
		return m.viewCloneRepoSummary(result)
	}

	b.WriteString("  Operation complete.\n")
	return b.String()
}

func (m Model) viewCreateRepoSummary(result repo.CreateResult) string {
	var b strings.Builder

	repoName := filepath.Base(result.RepoPath)

	// Check for GitHub phase failures
	ghFailed := false
	var ghErrMsg string
	for _, phase := range result.Phases {
		if phase.Name == "GitHub" && phase.HasFailures() {
			ghFailed = true
			for _, step := range phase.Steps {
				if step.Error != nil {
					ghErrMsg = step.Error.Error()
					break
				}
			}
			break
		}
	}

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Repository Created", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	if ghFailed {
		b.WriteString(fmt.Sprintf("  %s %s ready (local only)\n\n",
			styleIndicatorDone.Render(indicatorDone), repoName))
	} else {
		b.WriteString(fmt.Sprintf("  %s %s ready\n\n",
			styleIndicatorDone.Render(indicatorDone), repoName))
	}

	b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Path"), result.RepoPath))
	b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Branch"), "main"))

	if result.GitHubURL != "" {
		b.WriteString(fmt.Sprintf("    %-10s %s %s\n", styleDim.Render("GitHub"),
			result.GitHubURL, styleSuccess.Render("\u25cf")))
	} else if ghFailed {
		msg := "failed to publish"
		if ghErrMsg != "" {
			msg = ghErrMsg
		}
		b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("GitHub"),
			styleError.Render(indicatorFailed+" "+msg)))
	}

	worktreePath := result.WorktreePath
	if worktreePath == "" && result.RepoPath != "" {
		worktreePath = result.RepoPath + "/main"
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("    cd %s\n\n", styleDim.Render(worktreePath)))
	b.WriteString(styleDim.Render("  enter open in sentei \u00b7 q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewCloneRepoSummary(result repo.CloneResult) string {
	var b strings.Builder

	repoName := filepath.Base(result.RepoPath)

	b.WriteString(styleTitle.Render(fmt.Sprintf("  sentei %s Repository Cloned", "\u2500")))
	b.WriteString("\n\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  %s %s ready\n\n",
		styleIndicatorDone.Render(indicatorDone), repoName))

	b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Path"), result.RepoPath))
	b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Branch"), result.DefaultBranch))
	if result.OriginURL != "" {
		b.WriteString(fmt.Sprintf("    %-10s %s\n", styleDim.Render("Origin"), result.OriginURL))
	}

	worktreePath := result.WorktreePath
	if worktreePath == "" && result.RepoPath != "" && result.DefaultBranch != "" {
		worktreePath = result.RepoPath + "/" + result.DefaultBranch
	}

	b.WriteString("\n")
	b.WriteString(separator(m.width))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("    cd %s\n\n", styleDim.Render(worktreePath)))
	b.WriteString(styleDim.Render("  enter open in sentei \u00b7 q quit"))
	b.WriteString("\n")

	return b.String()
}
