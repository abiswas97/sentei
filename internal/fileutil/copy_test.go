package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0644)

	dst := filepath.Join(t.TempDir(), "copy")
	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(got) != "hello" {
		t.Errorf("a.txt = %q, want hello", got)
	}
	got, _ = os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if string(got) != "world" {
		t.Errorf("sub/b.txt = %q, want world", got)
	}
}
