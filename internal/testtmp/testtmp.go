// Package testtmp provides temp-dir helpers for tests. It is a dependency-free
// leaf package so any package's tests can use it without import cycles.
//
// On macOS, Spotlight (mds/mdworker) indexes files under the default TMPDIR
// (/var/folders/...). When a test writes git object files and then removes the
// dir, the indexer can still be holding a file, so RemoveAll fails with
// ENOTEMPTY and the test flakes. These helpers address that two ways: a
// .metadata_never_index marker (the documented Spotlight exclusion) and a
// bounded retry on RemoveAll.
package testtmp

import (
	"os"
	"path/filepath"
	"testing"
)

// markNoIndex excludes dir's subtree from macOS Spotlight indexing.
func markNoIndex(dir string) {
	_ = os.WriteFile(filepath.Join(dir, ".metadata_never_index"), nil, 0o644)
}

// removeWithRetry removes dir, retrying to absorb the transient ENOTEMPTY race.
// Bounded and condition-based (retry until success), with no sleep.
func removeWithRetry(dir string) {
	for i := 0; i < 50; i++ {
		if err := os.RemoveAll(dir); err == nil {
			return
		}
	}
}

// RobustTempDir is like t.TempDir but excludes its subtree from Spotlight and
// retries cleanup, so a macOS indexing race can't fail an otherwise-passing
// test. Cleanup is best-effort and never fails the test.
func RobustTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sentei-test-")
	if err != nil {
		t.Fatal(err)
	}
	markNoIndex(dir)
	t.Cleanup(func() { removeWithRetry(dir) })
	return dir
}

// RunWithIsolatedTemp runs a package's tests with TMPDIR pointed at a dedicated
// Spotlight-excluded dir, so every t.TempDir in the package is safe from the
// macOS indexing race without changing individual tests. Use from TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(testtmp.RunWithIsolatedTemp(m)) }
func RunWithIsolatedTemp(m *testing.M) int {
	base, err := os.MkdirTemp("", "sentei-tests-")
	if err != nil {
		panic(err)
	}
	markNoIndex(base)
	old, hadOld := os.LookupEnv("TMPDIR")
	_ = os.Setenv("TMPDIR", base)

	code := m.Run()

	if hadOld {
		_ = os.Setenv("TMPDIR", old)
	} else {
		_ = os.Unsetenv("TMPDIR")
	}
	removeWithRetry(base)
	return code
}
