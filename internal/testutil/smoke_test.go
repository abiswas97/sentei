package testutil

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestSmoke_LaunchAndQuit(t *testing.T) {
	repoPath := SetupBareRepo(t)

	tm := LaunchTUI(t, repoPath)

	// Wait for TUI to render something before sending quit.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("sentei"))
	}, teatest.WithDuration(10*time.Second), teatest.WithCheckInterval(50*time.Millisecond))

	// Send 'q' to quit.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	// Wait for clean exit.
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	// Read the final output to verify the TUI rendered something.
	out := readBytes(t, tm.FinalOutput(t))
	if len(out) == 0 {
		t.Error("expected non-empty final output from TUI")
	}
}

func readBytes(t *testing.T, r interface{ Read([]byte) (int, error) }) []byte {
	t.Helper()
	buf := make([]byte, 4096)
	var result []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return result
}
