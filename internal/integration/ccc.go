package integration

func cocoindexCode() Integration {
	return Integration{
		Name:             "cocoindex-code",
		ShortDescription: "Semantic code search for AI agents",
		Description:      "Indexes your codebase for semantic search, letting AI agents find code by meaning rather than keywords. Supports incremental updates.",
		URL:              "https://github.com/cocoindex-io/cocoindex-code",
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
			// [full] is upstream's documented batteries-included extra: it
			// pulls cocoindex[sentence-transformers] so the default local-
			// embedding `ccc index` has its runtime deps. --python pins
			// resolution: uv otherwise honors any .python-version in the
			// cwd, and a 3.10 pin makes the >=3.11 requirement
			// unsatisfiable.
			Command:      `uv tool install --upgrade --python 3.11 "cocoindex-code[full]"`,
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
