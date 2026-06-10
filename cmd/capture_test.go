package cmd

import (
	"io"
	"os"
	"testing"
)

// captureStdout runs fn while capturing everything written to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	return captureFile(t, &os.Stdout, fn)
}

// captureStderr runs fn while capturing everything written to os.Stderr.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	return captureFile(t, &os.Stderr, fn)
}

func captureFile(t *testing.T, target **os.File, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := *target
	*target = w
	defer func() { *target = orig }()

	done := make(chan string)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	fn()
	w.Close()
	return <-done
}
