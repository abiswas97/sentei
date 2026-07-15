#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
common_dir="$(git rev-parse --git-common-dir)"
case "$common_dir" in
/*) ;;
*) common_dir="$(cd "$common_dir" && pwd -P)" ;;
esac
hooks_dir="$common_dir/sentei-hooks"

mkdir -p "$hooks_dir"
for hook in pre-commit commit-msg; do
	install -m 0755 "$SCRIPT_DIR/$hook" "$hooks_dir/$hook"
done
git config --local core.hooksPath "$hooks_dir"

echo "Git hooks installed (hooksPath → $hooks_dir)"
