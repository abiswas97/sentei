package integration

func codeReviewGraph() Integration {
	return Integration{
		Name:        "code-review-graph",
		Description: "Build code graph for AI-assisted code review",
		URL:         "https://github.com/tirth8205/code-review-graph",
		Dependencies: []Dependency{
			{
				Name:   "python3.10+",
				Detect: `python3 -c "import sys; assert sys.version_info >= (3,10)"`,
			},
			{
				Name:    "pipx",
				Detect:  "pipx --version",
				Install: "brew install pipx",
			},
		},
		Detect: DetectSpec{
			Command: "code-review-graph --version",
		},
		Install: InstallSpec{
			Command: "pipx install code-review-graph",
		},
		Setup: SetupSpec{
			Command:    "code-review-graph build --repo {path}",
			WorkingDir: "repo",
		},
		Teardown: TeardownSpec{
			Dirs: []string{".code-review-graph/"},
		},
		GitignoreEntries: []string{".code-review-graph/"},
	}
}
