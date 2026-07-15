#!/usr/bin/env bash
set -euo pipefail

ROOT=/tmp/sentei-vhs-progress-arc
case "$ROOT" in
  /tmp/sentei-vhs-progress-arc) ;;
  *) printf 'refusing unsafe fixture root: %s\n' "$ROOT" >&2; exit 64 ;;
esac

mkdir -p "$ROOT"
for child in home xdg-config xdg-cache xdg-data repo seed shims logs; do
  rm -rf "${ROOT:?}/$child"
  mkdir -p "$ROOT/$child"
done
mkdir -p "$ROOT/outputs" "$ROOT/frames"

export HOME="$ROOT/home"
export XDG_CONFIG_HOME="$ROOT/xdg-config"
export XDG_CACHE_HOME="$ROOT/xdg-cache"
export XDG_DATA_HOME="$ROOT/xdg-data"
export GIT_CONFIG_GLOBAL="$ROOT/home/gitconfig"
export GIT_CONFIG_SYSTEM=/dev/null
export GIT_AUTHOR_NAME='Sentei Demo'
export GIT_AUTHOR_EMAIL='sentei-demo@example.invalid'
export GIT_COMMITTER_NAME="$GIT_AUTHOR_NAME"
export GIT_COMMITTER_EMAIL="$GIT_AUTHOR_EMAIL"
export LC_ALL=C
export TZ=UTC
export TERM=xterm-256color

GIT=/usr/bin/git
"$GIT" config --global init.defaultBranch main
"$GIT" init "$ROOT/seed" >/dev/null
printf '# deterministic sentei demo\n' >"$ROOT/seed/README.md"
"$GIT" -C "$ROOT/seed" add README.md
"$GIT" -C "$ROOT/seed" commit -m 'fixture: initial content' >/dev/null
"$GIT" clone --bare "$ROOT/seed" "$ROOT/repo/demo.git" >/dev/null

for branch in alpha beta gamma; do
  path="$ROOT/repo/worktrees/$branch"
  mkdir -p "$(dirname "$path")"
  "$GIT" --git-dir="$ROOT/repo/demo.git" worktree add -b "demo/$branch" "$path" main >/dev/null
  printf '%s\n' "$branch" >"$path/$branch.txt"
  "$GIT" -C "$path" add "$branch.txt"
  "$GIT" -C "$path" commit -m "fixture: $branch" >/dev/null
done

cat >"$ROOT/shims/git" <<'SHIM'
#!/usr/bin/env bash
set -euo pipefail
ROOT=/tmp/sentei-vhs-progress-arc
printf 'git %q ' "$@" >>"$ROOT/logs/git.log"
printf '\n' >>"$ROOT/logs/git.log"
for arg in "$@"; do
  case "$arg" in
    *://*|git@*) printf 'network git argument denied: %s\n' "$arg" >&2; exit 90 ;;
  esac
done
case " ${*} " in
  *' fetch '*|*' pull '*|*' push '*|*' clone '*|*' ls-remote '*)
    printf 'network-capable git verb denied\n' >&2
    exit 91
    ;;
esac
if [[ " ${*} " == *' worktree remove '* ]]; then
  for arg in "$@"; do
    case "$arg" in
      /*)
        case "$arg" in "$ROOT"/*) ;; *) printf 'worktree escape denied: %s\n' "$arg" >&2; exit 92 ;; esac
        ;;
    esac
  done
  sleep 0.35
fi
exec /usr/bin/git "$@"
SHIM

cat >"$ROOT/shims/ccc" <<'SHIM'
#!/usr/bin/env bash
set -euo pipefail
ROOT=/tmp/sentei-vhs-progress-arc
printf 'ccc %s\n' "$*" >>"$ROOT/logs/ccc.log"
case "${1:-}" in
  init|--version) exit 0 ;;
  index) printf 'deterministic index failure\n' >&2; exit 17 ;;
  *) exit 0 ;;
esac
SHIM

cat >"$ROOT/shims/python3" <<'SHIM'
#!/usr/bin/env bash
exit 0
SHIM
cat >"$ROOT/shims/uv" <<'SHIM'
#!/usr/bin/env bash
exit 0
SHIM

for denied in gh curl wget brew npm npx pnpm yarn pip pip3 pipx; do
  cat >"$ROOT/shims/$denied" <<'SHIM'
#!/usr/bin/env bash
printf 'network or package-manager command denied: %s\n' "$0" >&2
exit 93
SHIM
done
chmod +x "$ROOT/shims/"*

cat >"$ROOT/demo.env" <<EOF
export HOME='$HOME'
export XDG_CONFIG_HOME='$XDG_CONFIG_HOME'
export XDG_CACHE_HOME='$XDG_CACHE_HOME'
export XDG_DATA_HOME='$XDG_DATA_HOME'
export GIT_CONFIG_GLOBAL='$GIT_CONFIG_GLOBAL'
export GIT_CONFIG_SYSTEM=/dev/null
export GIT_AUTHOR_NAME='$GIT_AUTHOR_NAME'
export GIT_AUTHOR_EMAIL='$GIT_AUTHOR_EMAIL'
export GIT_COMMITTER_NAME='$GIT_COMMITTER_NAME'
export GIT_COMMITTER_EMAIL='$GIT_COMMITTER_EMAIL'
export LC_ALL=C
export TZ=UTC
export TERM=xterm-256color
export SENTEI_MOTION=off
export PATH='$ROOT/shims:/usr/bin:/bin'
EOF

printf 'fixture ready: %s\n' "$ROOT/repo/demo.git"
