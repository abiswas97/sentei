package tui

import (
	"path/filepath"

	"github.com/abiswas97/sentei/internal/testutil/mock"
)

// bareDirRunner returns a runner that answers `git rev-parse --git-common-dir`
// for root with root/.bare, letting flows that resolve the bare dir run without
// a real git repository on disk.
func bareDirRunner(root string) *mock.Runner {
	return &mock.Runner{Responses: map[string]mock.Response{
		root + ":[rev-parse --git-common-dir]": {Output: filepath.Join(root, ".bare")},
	}}
}

func stripAnsi(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
			continue
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}
