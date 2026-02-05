package git

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
}
