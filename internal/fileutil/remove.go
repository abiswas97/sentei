package fileutil

import "os"

// RemoveAllRetry removes path, retrying on transient failures. On macOS,
// Spotlight/fseventsd can briefly hold newly written files in git object dirs,
// making a single RemoveAll fail with ENOTEMPTY; retrying lets it drain. The
// loop is bounded and condition-based (retry until success), with no sleep.
func RemoveAllRetry(path string) error {
	var err error
	for i := 0; i < 50; i++ {
		if err = os.RemoveAll(path); err == nil {
			return nil
		}
	}
	return err
}
