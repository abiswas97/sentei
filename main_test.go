package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	bin := t.TempDir() + "/sentei"
	build := exec.Command("go", "build", "-ldflags", "-X main.version=v1.2.3 -X main.commit=abc1234 -X main.date=1970-01-01", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	t.Run("prints version and exits", func(t *testing.T) {
		out, err := exec.Command(bin, "--version").Output()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := strings.TrimSpace(string(out))
		want := "sentei v1.2.3 (abc1234, 1970-01-01)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("default version is dev", func(t *testing.T) {
		devBin := t.TempDir() + "/sentei"
		build := exec.Command("go", "build", "-o", devBin, ".")
		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("build failed: %v\n%s", err, out)
		}
		out, err := exec.Command(devBin, "--version").Output()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := strings.TrimSpace(string(out))
		want := "sentei dev (none, unknown)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("version takes precedence over other flags", func(t *testing.T) {
		out, err := exec.Command(bin, "--version", "--dry-run").Output()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := strings.TrimSpace(string(out))
		want := "sentei v1.2.3 (abc1234, 1970-01-01)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
