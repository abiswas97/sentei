package git

import "testing"

func TestIsProtectedBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"refs/heads/main", true},
		{"refs/heads/master", true},
		{"refs/heads/develop", true},
		{"refs/heads/dev", true},
		{"main", true},
		{"master", true},
		{"develop", true},
		{"dev", true},

		{"refs/heads/feature/dev-tools", false},
		{"refs/heads/feature/main-page", false},
		{"refs/heads/development", false},
		// Case-insensitive: a branch conceptually the mainline is protected
		// regardless of capitalization.
		{"refs/heads/Main", true},
		{"refs/heads/MASTER", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			got := IsProtectedBranch(tt.branch)
			if got != tt.want {
				t.Errorf("IsProtectedBranch(%q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}

func TestIsProtectedBranchWith_ProtectsNonStandardDefault(t *testing.T) {
	// A non-standard default branch must be protected even though it is not in
	// the built-in convention set.
	if !IsProtectedBranchWith("refs/heads/production", "production") {
		t.Error("the detected default branch 'production' must be protected")
	}
	if !IsProtectedBranchWith("trunk", "trunk") {
		t.Error("the detected default branch 'trunk' must be protected")
	}
	// Still protects the static set even when a default is supplied.
	if !IsProtectedBranchWith("refs/heads/main", "production") {
		t.Error("'main' must remain protected")
	}
	// Non-default, non-convention feature branch is not protected.
	if IsProtectedBranchWith("refs/heads/feature/x", "production") {
		t.Error("a feature branch must not be protected")
	}
	// Empty default falls back to the static set only.
	if IsProtectedBranchWith("refs/heads/production", "") {
		t.Error("with no default supplied, 'production' is not in the static set")
	}
}
