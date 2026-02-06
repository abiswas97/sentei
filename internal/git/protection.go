package git

import "strings"

var protectedBranches = map[string]bool{
	"main":    true,
	"master":  true,
	"develop": true,
	"dev":     true,
}

func IsProtectedBranch(branch string) bool {
	name := strings.TrimPrefix(branch, "refs/heads/")
	return protectedBranches[name]
}
