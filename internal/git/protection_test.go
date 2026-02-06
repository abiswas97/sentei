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
		{"refs/heads/Main", false},
		{"refs/heads/MASTER", false},
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
