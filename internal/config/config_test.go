package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func boolPtr(b bool) *bool {
	return &b
}

func assertConfigEqual(t *testing.T, got, want Config) {
	t.Helper()

	if len(got.Ecosystems) != len(want.Ecosystems) {
		t.Fatalf("Ecosystems length: got %d, want %d", len(got.Ecosystems), len(want.Ecosystems))
	}
	for i, gotE := range got.Ecosystems {
		wantE := want.Ecosystems[i]
		if gotE.Name != wantE.Name {
			t.Errorf("Ecosystems[%d].Name: got %q, want %q", i, gotE.Name, wantE.Name)
		}
		switch {
		case gotE.Enabled == nil && wantE.Enabled == nil:
			// both nil — ok
		case gotE.Enabled == nil || wantE.Enabled == nil:
			t.Errorf("Ecosystems[%d].Enabled: got %v, want %v", i, gotE.Enabled, wantE.Enabled)
		case *gotE.Enabled != *wantE.Enabled:
			t.Errorf("Ecosystems[%d].Enabled: got %v, want %v", i, *gotE.Enabled, *wantE.Enabled)
		}
		if len(gotE.Detect.Files) != len(wantE.Detect.Files) {
			t.Errorf("Ecosystems[%d].Detect.Files length: got %d, want %d", i, len(gotE.Detect.Files), len(wantE.Detect.Files))
		} else {
			for j, f := range gotE.Detect.Files {
				if f != wantE.Detect.Files[j] {
					t.Errorf("Ecosystems[%d].Detect.Files[%d]: got %q, want %q", i, j, f, wantE.Detect.Files[j])
				}
			}
		}
		if gotE.Install.Command != wantE.Install.Command {
			t.Errorf("Ecosystems[%d].Install.Command: got %q, want %q", i, gotE.Install.Command, wantE.Install.Command)
		}
		if gotE.Install.WorkspaceDetect != wantE.Install.WorkspaceDetect {
			t.Errorf("Ecosystems[%d].Install.WorkspaceDetect: got %q, want %q", i, gotE.Install.WorkspaceDetect, wantE.Install.WorkspaceDetect)
		}
		if gotE.Install.WorkspaceInstall != wantE.Install.WorkspaceInstall {
			t.Errorf("Ecosystems[%d].Install.WorkspaceInstall: got %q, want %q", i, gotE.Install.WorkspaceInstall, wantE.Install.WorkspaceInstall)
		}
		switch {
		case gotE.Install.Parallel == nil && wantE.Install.Parallel == nil:
			// both nil — ok
		case gotE.Install.Parallel == nil || wantE.Install.Parallel == nil:
			t.Errorf("Ecosystems[%d].Install.Parallel: got %v, want %v", i, gotE.Install.Parallel, wantE.Install.Parallel)
		case *gotE.Install.Parallel != *wantE.Install.Parallel:
			t.Errorf("Ecosystems[%d].Install.Parallel: got %v, want %v", i, *gotE.Install.Parallel, *wantE.Install.Parallel)
		}
		if len(gotE.EnvFiles) != len(wantE.EnvFiles) {
			t.Errorf("Ecosystems[%d].EnvFiles length: got %d, want %d", i, len(gotE.EnvFiles), len(wantE.EnvFiles))
		} else {
			for j, f := range gotE.EnvFiles {
				if f != wantE.EnvFiles[j] {
					t.Errorf("Ecosystems[%d].EnvFiles[%d]: got %q, want %q", i, j, f, wantE.EnvFiles[j])
				}
			}
		}
		if len(gotE.PostInstall) != len(wantE.PostInstall) {
			t.Errorf("Ecosystems[%d].PostInstall length: got %d, want %d", i, len(gotE.PostInstall), len(wantE.PostInstall))
		} else {
			for j, cmd := range gotE.PostInstall {
				if cmd != wantE.PostInstall[j] {
					t.Errorf("Ecosystems[%d].PostInstall[%d]: got %q, want %q", i, j, cmd, wantE.PostInstall[j])
				}
			}
		}
	}

	if len(got.ProtectedBranches) != len(want.ProtectedBranches) {
		t.Errorf("ProtectedBranches length: got %d, want %d", len(got.ProtectedBranches), len(want.ProtectedBranches))
	} else {
		for i, b := range got.ProtectedBranches {
			if b != want.ProtectedBranches[i] {
				t.Errorf("ProtectedBranches[%d]: got %q, want %q", i, b, want.ProtectedBranches[i])
			}
		}
	}

	if len(got.IntegrationsEnabled) != len(want.IntegrationsEnabled) {
		t.Errorf("IntegrationsEnabled length: got %d, want %d", len(got.IntegrationsEnabled), len(want.IntegrationsEnabled))
	} else {
		for i, s := range got.IntegrationsEnabled {
			if s != want.IntegrationsEnabled[i] {
				t.Errorf("IntegrationsEnabled[%d]: got %q, want %q", i, s, want.IntegrationsEnabled[i])
			}
		}
	}
}

func TestConfigUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Config
	}{
		{
			name: "full config",
			input: `
ecosystems:
  - name: node
    enabled: true
    detect:
      files:
        - package.json
    install:
      command: npm install
      workspace_detect: package.json
      workspace_install: npm install --workspaces
      parallel: true
    env_files:
      - .env
      - .env.local
    post_install:
      - npm run build
protected_branches:
  - main
  - develop
integrations_enabled:
  - github
  - slack
`,
			want: Config{
				Ecosystems: []EcosystemConfig{
					{
						Name:    "node",
						Enabled: boolPtr(true),
						Detect:  DetectConfig{Files: []string{"package.json"}},
						Install: InstallConfig{
							Command:          "npm install",
							WorkspaceDetect:  "package.json",
							WorkspaceInstall: "npm install --workspaces",
							Parallel:         boolPtr(true),
						},
						EnvFiles:    []string{".env", ".env.local"},
						PostInstall: []string{"npm run build"},
					},
				},
				ProtectedBranches:   []string{"main", "develop"},
				IntegrationsEnabled: []string{"github", "slack"},
			},
		},
		{
			name:  "empty config",
			input: `{}`,
			want:  Config{},
		},
		{
			name: "ecosystem with enabled false",
			input: `
ecosystems:
  - name: python
    enabled: false
    detect:
      files:
        - requirements.txt
    install:
      command: pip install -r requirements.txt
`,
			want: Config{
				Ecosystems: []EcosystemConfig{
					{
						Name:    "python",
						Enabled: boolPtr(false),
						Detect:  DetectConfig{Files: []string{"requirements.txt"}},
						Install: InstallConfig{
							Command: "pip install -r requirements.txt",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got Config
			if err := yaml.Unmarshal([]byte(tc.input), &got); err != nil {
				t.Fatalf("yaml.Unmarshal error: %v", err)
			}
			assertConfigEqual(t, got, tc.want)
		})
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled *bool
		want    bool
	}{
		{"nil means enabled", nil, true},
		{"explicit true", boolPtr(true), true},
		{"explicit false", boolPtr(false), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := EcosystemConfig{Enabled: tc.enabled}
			if got := e.IsEnabled(); got != tc.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsParallel(t *testing.T) {
	tests := []struct {
		name     string
		parallel *bool
		want     bool
	}{
		{"nil means not parallel", nil, false},
		{"explicit true", boolPtr(true), true},
		{"explicit false", boolPtr(false), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := InstallConfig{Parallel: tc.parallel}
			if got := i.IsParallel(); got != tc.want {
				t.Errorf("IsParallel() = %v, want %v", got, tc.want)
			}
		})
	}
}
