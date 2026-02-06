## 1. Version Flag

- [x] 1.1 Add `var version = "dev"` to `main.go` and a `--version` boolean flag
- [x] 1.2 Add early exit: when `--version` is set, print `sentei <version>` to stdout and `os.Exit(0)` before any other logic (playground, git commands, TUI)
- [x] 1.3 Add test for version output format and early exit behavior

## 2. README

- [x] 2.1 Create `README.md` with project description and overview
- [x] 2.2 Add installation section: `go install`, build from source with ldflags
- [x] 2.3 Add usage section: CLI flags table (`--version`, `--dry-run`, `--playground`, `--playground-keep`, positional repo path)
- [x] 2.4 Add key bindings table (navigation, selection, sorting, filtering, confirmation, quit)
- [x] 2.5 Add status indicators legend (`[ok]`, `[~]`, `[!]`, `[L]`, `[P]`)
