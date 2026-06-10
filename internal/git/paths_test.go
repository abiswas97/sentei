package git

import (
	"path/filepath"
	"testing"
)

func TestWorktreeDirName(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{name: "slash replaced with dash", branch: "feature/auth", want: "feature-auth"},
		{name: "multiple slashes", branch: "bugfix/login/redirect", want: "bugfix-login-redirect"},
		{name: "no slash unchanged", branch: "hotfix", want: "hotfix"},
		{name: "trailing slash stripped", branch: "feature/", want: "feature-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WorktreeDirName(tt.branch)
			if got != tt.want {
				t.Errorf("WorktreeDirName(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestWorktreePath(t *testing.T) {
	got := WorktreePath("/repo", "feature/auth")
	want := filepath.Join("/repo", "feature-auth")
	if got != want {
		t.Errorf("WorktreePath = %q, want %q", got, want)
	}
}
