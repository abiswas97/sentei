package integration

func cocoindexCode() Integration {
	return Integration{
		Name:        "cocoindex-code",
		Description: "Indexes your codebase for semantic search, letting AI agents find code by meaning rather than keywords. Supports incremental updates.",
		URL:         "https://github.com/cocoindex-io/cocoindex-code",
		Dependencies: []Dependency{
			{
				Name:   "python3.11+",
				Detect: `python3 -c "import sys; assert sys.version_info >= (3,11)"`,
			},
			{
				Name:    "uv",
				Detect:  "uv --version",
				Install: "brew install uv",
			},
		},
		Detect: DetectSpec{
			BinaryName: "ccc",
		},
		Install: InstallSpec{
			Command:      `uv tool install --upgrade cocoindex-code --prerelease explicit --with "cocoindex>=1.0.0a24"`,
			FirstRunNote: "Downloads ~87MB embedding model on first use",
		},
		Setup: SetupSpec{
			Command:    "ccc init && ccc index",
			WorkingDir: "worktree",
		},
		Teardown: TeardownSpec{
			Command: "ccc reset --all --force",
			Dirs:    []string{".cocoindex_code/"},
		},
		GitignoreEntries: []string{".cocoindex_code/"},
		IndexCopyDir:     ".cocoindex_code",
	}
}
