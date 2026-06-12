package testtmp

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The guard for the isolation itself: under HermeticGitEnv, git must not be
// able to see ANY global configuration, even when one demonstrably exists.
// This is the invariant that makes a test path bug fail loudly instead of
// silently committing with (or overwriting) the developer's real identity.
func TestHermeticGitEnv_GitCannotSeeOutsideConfig(t *testing.T) {
	// A decoy "global" config proves the void: point an un-hardened git at
	// it, confirm it reads; then confirm the hermetic env reads nothing.
	decoy := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(decoy, []byte("[user]\n\tname = Decoy\n\temail = decoy@example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	read := func(env []string) string {
		cmd := exec.Command("git", "config", "--global", "--get", "user.name")
		cmd.Dir = t.TempDir()
		cmd.Env = env
		out, _ := cmd.CombinedOutput()
		return strings.TrimSpace(string(out))
	}

	if got := read(append(os.Environ(), "GIT_CONFIG_GLOBAL="+decoy)); got != "Decoy" {
		t.Fatalf("control failed: expected decoy global config to be readable, got %q", got)
	}
	if got := read(append(HermeticGitEnv(), "HOME="+filepath.Dir(decoy))); got != "" {
		t.Fatalf("hermetic git read a global config: %q", got)
	}
}

func TestHermeticGitEnv_CommitsCarryTestIdentity(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = HermeticGitEnv()
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "f")
	git("commit", "-qm", "probe")

	author := git("log", "--format=%an <%ae> %cn <%ce>", "-1")
	want := "sentei-test <test@sentei.invalid> sentei-test <test@sentei.invalid>"
	if author != want {
		t.Fatalf("hermetic commit identity = %q, want %q (a real identity leaking in here is the bug class this guards against)", author, want)
	}
}
