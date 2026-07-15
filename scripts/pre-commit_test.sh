#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
tmp="$(mktemp -d "${TMPDIR:-/tmp}/sentei-hooks.XXXXXX")"
repo="$tmp/repo"
worktree="$tmp/worktree"
trap 'rm -rf "$tmp"' EXIT

# The fixture must not discover or mutate the caller's repository, hooks, or
# identity configuration. Establish the isolated process environment before
# the first fixture Git command.
mkdir -p "$tmp/home" "$tmp/xdg"
export HOME="$tmp/home"
export XDG_CONFIG_HOME="$tmp/xdg"
export GIT_CONFIG_GLOBAL="$tmp/gitconfig-global"
export GIT_CONFIG_SYSTEM=/dev/null
export GIT_CONFIG_NOSYSTEM=1
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY
unset GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_COMMON_DIR GIT_PREFIX
unset GIT_CONFIG_COUNT GIT_CONFIG_PARAMETERS

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

run_hook() {
	(
		cd "$repo"
		env "$@" "$SCRIPT_DIR/pre-commit"
	)
}

expect_reject() {
	role="$1"
	shift
	if output="$(run_hook "$@" 2>&1)"; then
		fail "invalid effective $role identity was accepted"
	fi
	case "$output" in
	*"$role"*) ;;
	*) fail "$role rejection did not identify the invalid role: $output" ;;
	esac
}

resolve_common_dir() {
	root="$1"
	common="$(git -C "$root" rev-parse --git-common-dir)"
	case "$common" in
	/*) printf '%s\n' "$common" ;;
	*) (cd "$root/$common" && pwd -P) ;;
	esac
}

git init -q "$repo"
git -C "$repo" config user.name "Valid Config"
git -C "$repo" config user.email "avishek.biswas.1997@gmail.com"
git -C "$repo" config core.hooksPath /dev/null
printf 'fixture\n' >"$repo/README.md"
git -C "$repo" add README.md
git -C "$repo" commit -q -m "chore: initialize fixture"

expect_reject author \
	GIT_AUTHOR_NAME=sentei-test GIT_AUTHOR_EMAIL=test@sentei.invalid \
	GIT_COMMITTER_NAME="Valid User" GIT_COMMITTER_EMAIL=valid@example.com
expect_reject committer \
	GIT_AUTHOR_NAME="Valid User" GIT_AUTHOR_EMAIL=valid@example.com \
	GIT_COMMITTER_NAME=sentei-test GIT_COMMITTER_EMAIL=test@sentei.invalid

git -C "$repo" config user.name sentei-test
git -C "$repo" config user.email test@sentei.invalid
if output="$(run_hook 2>&1)"; then
	fail "configured reserved identity was accepted"
fi
case "$output" in
*author*) ;;
*) fail "configured identity rejection did not identify the author role: $output" ;;
esac
run_hook \
	GIT_AUTHOR_NAME="Valid Author" GIT_AUTHOR_EMAIL=author@example.com \
	GIT_COMMITTER_NAME="Valid Committer" GIT_COMMITTER_EMAIL=committer@example.com \
	>/dev/null || fail "valid environment identities did not override invalid repository config"

git -C "$repo" config user.name "Valid Config"
git -C "$repo" config user.email valid-config@example.com
git -C "$repo" worktree add -q -b hook-test "$worktree"

(
	cd "$repo"
	"$SCRIPT_DIR/install-hooks.sh" >/dev/null
)
common_dir="$(resolve_common_dir "$repo")"
expected_hooks="$common_dir/sentei-hooks"
configured_hooks="$(git -C "$repo" config --local --get core.hooksPath)"
[ "$configured_hooks" = "$expected_hooks" ] || fail "hooksPath = $configured_hooks, want $expected_hooks"
case "$configured_hooks" in
"$common_dir"/*) ;;
*) fail "hooksPath is not beneath the common Git directory: $configured_hooks" ;;
esac
cmp -s "$SCRIPT_DIR/pre-commit" "$expected_hooks/pre-commit" || fail "installed pre-commit differs from checked-in hook"
cmp -s "$SCRIPT_DIR/commit-msg" "$expected_hooks/commit-msg" || fail "installed commit-msg differs from checked-in hook"
[ -x "$expected_hooks/pre-commit" ] || fail "installed pre-commit is not executable"
[ -x "$expected_hooks/commit-msg" ] || fail "installed commit-msg is not executable"

printf 'stale\n' >"$expected_hooks/pre-commit"
(
	cd "$worktree"
	"$SCRIPT_DIR/install-hooks.sh" >/dev/null
)
[ "$(git -C "$worktree" config --local --get core.hooksPath)" = "$expected_hooks" ] || fail "second worktree changed the shared hooksPath"
cmp -s "$SCRIPT_DIR/pre-commit" "$expected_hooks/pre-commit" || fail "second worktree install did not refresh hook content"

(
	cd "$worktree"
	if GIT_AUTHOR_NAME=sentei-test GIT_AUTHOR_EMAIL=test@sentei.invalid \
		GIT_COMMITTER_NAME="Valid User" GIT_COMMITTER_EMAIL=valid@example.com \
		git hook run pre-commit >/dev/null 2>&1; then
		fail "active installed hook accepted an invalid author"
	fi
	GIT_AUTHOR_NAME="Valid Author" GIT_AUTHOR_EMAIL=author@example.com \
		GIT_COMMITTER_NAME="Valid Committer" GIT_COMMITTER_EMAIL=committer@example.com \
		git hook run pre-commit >/dev/null
)

echo "hook tests passed"
