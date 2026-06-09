// Package testtmp provides a temp-dir helper for tests. It is a dependency-free
// leaf package so any package's tests can use it without import cycles.
package testtmp

import (
	"os"
	"testing"
)

// RobustTempDir is like t.TempDir but tolerates the macOS Spotlight/fseventsd
// race where background indexers transiently hold newly written files in git's
// object dirs (e.g. tmp_objdir-incoming-*), making RemoveAll fail with ENOTEMPTY
// mid-sweep. Cleanup is best-effort and never fails an otherwise-passing test.
// The retry loop is condition-based (retry until RemoveAll succeeds), no sleep.
func RobustTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sentei-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for i := 0; i < 50; i++ {
			if err := os.RemoveAll(dir); err == nil {
				return
			}
		}
	})
	return dir
}
