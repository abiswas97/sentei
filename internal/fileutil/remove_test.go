package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveAllRetry_RemovesPopulatedDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub", "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveAllRetry(filepath.Join(dir, "sub")); err != nil {
		t.Fatalf("RemoveAllRetry: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub")); !os.IsNotExist(err) {
		t.Error("directory should be removed")
	}
}

func TestRemoveAllRetry_AbsentPathIsNoError(t *testing.T) {
	if err := RemoveAllRetry(filepath.Join(t.TempDir(), "does-not-exist")); err != nil {
		t.Errorf("removing an absent path should not error, got %v", err)
	}
}

func TestRemoveAllRetry_ReturnsLastErrorWhenStuck(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	parent := t.TempDir()
	child := filepath.Join(parent, "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A read-only parent prevents unlinking the child; restore perms first at
	// cleanup (LIFO) so t.TempDir's own RemoveAll can still succeed.
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })

	if err := RemoveAllRetry(child); err == nil {
		t.Error("expected a non-nil error when the directory cannot be removed")
	}
}
