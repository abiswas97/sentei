package tui

import (
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abiswas97/sentei/internal/git"
)

func relaunchSentei(repoPath string) tea.Cmd {
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

// worktreeLabel returns the display name for a worktree: its branch, or the
// directory name for a detached HEAD so the label is never blank.
func worktreeLabel(wt git.Worktree) string {
	if branch := stripBranchPrefix(wt.Branch); branch != "" {
		return branch
	}
	return filepath.Base(wt.Path)
}
