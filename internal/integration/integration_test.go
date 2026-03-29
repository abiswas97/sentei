package integration

import "testing"

func TestAll(t *testing.T) {
	all := All()
	if len(all) != 2 {
		t.Fatalf("want 2 integrations, got %d", len(all))
	}
	names := map[string]bool{}
	for _, integ := range all {
		names[integ.Name] = true
	}
	if !names["code-review-graph"] {
		t.Error("missing code-review-graph")
	}
	if !names["cocoindex-code"] {
		t.Error("missing cocoindex-code")
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantNil bool
	}{
		{name: "existing", query: "code-review-graph"},
		{name: "existing ccc", query: "cocoindex-code"},
		{name: "not found", query: "nonexistent", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Get(tt.query)
			if tt.wantNil && got != nil {
				t.Error("want nil, got non-nil")
			}
			if !tt.wantNil && got == nil {
				t.Errorf("want %q, got nil", tt.query)
			}
		})
	}
}

func TestIntegrationFieldsComplete(t *testing.T) {
	for _, integ := range All() {
		t.Run(integ.Name, func(t *testing.T) {
			if integ.Name == "" {
				t.Error("empty Name")
			}
			if integ.Description == "" {
				t.Error("empty Description")
			}
			if integ.URL == "" {
				t.Error("empty URL")
			}
			if len(integ.Dependencies) == 0 {
				t.Error("no Dependencies")
			}
			if integ.Detect.Command == "" && integ.Detect.BinaryName == "" {
				t.Error("no Detect")
			}
			if integ.Install.Command == "" {
				t.Error("empty Install.Command")
			}
			if integ.Setup.Command == "" {
				t.Error("empty Setup.Command")
			}
			if integ.Setup.WorkingDir != "repo" && integ.Setup.WorkingDir != "worktree" {
				t.Errorf("Setup.WorkingDir must be 'repo' or 'worktree', got %q", integ.Setup.WorkingDir)
			}
			if integ.Teardown.Command == "" && len(integ.Teardown.Dirs) == 0 {
				t.Error("no Teardown")
			}
			if len(integ.GitignoreEntries) == 0 {
				t.Error("no GitignoreEntries")
			}
		})
	}
}

func TestDependencyFieldsComplete(t *testing.T) {
	for _, integ := range All() {
		for i, dep := range integ.Dependencies {
			t.Run(integ.Name+"/dep/"+dep.Name, func(t *testing.T) {
				if dep.Name == "" {
					t.Errorf("Dependency[%d]: empty Name", i)
				}
				if dep.Detect == "" {
					t.Errorf("Dependency[%d] %q: empty Detect", i, dep.Name)
				}
			})
		}
	}
}
