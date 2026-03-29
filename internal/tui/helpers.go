package tui

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
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
