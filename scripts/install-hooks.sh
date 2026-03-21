#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

git config core.hooksPath "$SCRIPT_DIR"

echo "Git hooks installed (hooksPath → $SCRIPT_DIR)"
