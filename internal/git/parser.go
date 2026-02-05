package git

import (
	"strings"
)

func ParsePorcelain(input string) ([]Worktree, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return []Worktree{}, nil
	}

	blocks := splitBlocks(input)
	worktrees := make([]Worktree, 0, len(blocks))

	for _, block := range blocks {
		wt := parseBlock(block)
		worktrees = append(worktrees, wt)
	}

	return worktrees, nil
}

func splitBlocks(input string) []string {
	var blocks []string
	var current []string

	for _, line := range strings.Split(input, "\n") {
		if line == "" {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}

	return blocks
}

func parseBlock(block string) Worktree {
	var wt Worktree

	for _, line := range strings.Split(block, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			wt.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			wt.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			wt.Branch = strings.TrimPrefix(line, "branch ")
		case line == "bare":
			wt.IsBare = true
		case line == "detached":
			wt.IsDetached = true
		case line == "locked":
			wt.IsLocked = true
		case strings.HasPrefix(line, "locked "):
			wt.IsLocked = true
			wt.LockReason = strings.TrimPrefix(line, "locked ")
		case line == "prunable":
			wt.IsPrunable = true
		case strings.HasPrefix(line, "prunable "):
			wt.IsPrunable = true
			wt.PruneReason = strings.TrimPrefix(line, "prunable ")
		}
	}

	return wt
}
