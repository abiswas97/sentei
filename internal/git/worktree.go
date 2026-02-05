package git

import "time"

type Worktree struct {
	Path        string
	HEAD        string
	Branch      string
	IsBare      bool
	IsLocked    bool
	LockReason  string
	IsPrunable  bool
	PruneReason string
	IsDetached  bool

	LastCommitDate        time.Time
	LastCommitSubject     string
	HasUncommittedChanges bool
	HasUntrackedFiles     bool
	IsEnriched            bool
	EnrichmentError       string
}
