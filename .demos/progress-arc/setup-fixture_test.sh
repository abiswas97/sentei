#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)
PRODUCTION_ROOT=/tmp/sentei-vhs-progress-arc
SANDBOX=$(mktemp -d /tmp/sentei-vhs-progress-arc-test.XXXXXX)
trap 'rm -rf "$SANDBOX"' EXIT

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

make_test_setup() {
  local case_dir=$1
  local fixture_root="$case_dir/sentei-vhs-progress-arc"
  mkdir -p "$case_dir"
  sed "s|$PRODUCTION_ROOT|$fixture_root|g" \
    "$SOURCE_DIR/setup-fixture.sh" >"$case_dir/setup-fixture.sh"
  chmod +x "$case_dir/setup-fixture.sh"
}

assert_rejected() {
  local setup=$1
  if "$setup" >"$setup.stdout" 2>"$setup.stderr"; then
    fail "unsafe fixture was accepted: $setup"
  fi
}

test_symlinked_root_is_rejected_without_touching_target() {
  local case_dir="$SANDBOX/root-symlink"
  local external="$case_dir/external"
  make_test_setup "$case_dir"
  mkdir -p "$external/repo"
  printf 'keep\n' >"$external/repo/sentinel"
  ln -s "$external" "$case_dir/sentei-vhs-progress-arc"

  assert_rejected "$case_dir/setup-fixture.sh"

  [[ -f "$external/repo/sentinel" ]] || fail 'symlinked root target was mutated'
}

test_retained_symlinks_are_rejected_without_touching_targets() {
  local child
  for child in outputs frames; do
    local case_dir="$SANDBOX/$child-symlink"
    local fixture_root="$case_dir/sentei-vhs-progress-arc"
    local external="$case_dir/external"
    make_test_setup "$case_dir"
    mkdir -p "$fixture_root" "$external"
    printf 'keep\n' >"$external/sentinel"
    ln -s "$external" "$fixture_root/$child"

    assert_rejected "$case_dir/setup-fixture.sh"

    [[ -f "$external/sentinel" ]] || fail "$child symlink target was deleted"
    [[ ! -e "$external/unexpected" ]] || fail "$child symlink target was written"
  done
}

test_symlinks_beneath_retained_directories_are_rejected() {
  local case_dir="$SANDBOX/retained-descendant-symlinks"
  local fixture_root="$case_dir/sentei-vhs-progress-arc"
  local external="$case_dir/external"
  make_test_setup "$case_dir"
  mkdir -p "$fixture_root/outputs" "$fixture_root/frames" "$external/directory"
  printf 'keep-file\n' >"$external/render.gif"
  printf 'keep-directory\n' >"$external/directory/sentinel"
  ln -s "$external/render.gif" "$fixture_root/outputs/removal-success.gif"
  ln -s "$external/directory" "$fixture_root/frames/nested"

  assert_rejected "$case_dir/setup-fixture.sh"

  [[ $(cat "$external/render.gif") == keep-file ]] || fail 'retained file symlink target was overwritten'
  [[ -f "$external/directory/sentinel" ]] || fail 'nested retained symlink target was mutated'
}

test_non_owned_root_is_rejected_before_reset() {
  local case_dir="$SANDBOX/non-owned"
  local fixture_root="$case_dir/sentei-vhs-progress-arc"
  local fake_bin="$case_dir/fake-bin"
  make_test_setup "$case_dir"
  mkdir -p "$fixture_root/repo" "$fake_bin"
  printf 'keep\n' >"$fixture_root/repo/sentinel"
  cat >"$fake_bin/id" <<'SHIM'
#!/usr/bin/env bash
if [[ "${1:-}" == -u ]]; then
  printf '2147483647\n'
else
  exec /usr/bin/id "$@"
fi
SHIM
  chmod +x "$fake_bin/id"

  if PATH="$fake_bin:$PATH" "$case_dir/setup-fixture.sh" \
      >"$case_dir/setup-fixture.sh.stdout" 2>"$case_dir/setup-fixture.sh.stderr"; then
    fail 'non-owned fixture root was accepted'
  fi

  [[ -f "$fixture_root/repo/sentinel" ]] || fail 'non-owned root was reset'
}

test_valid_fixture_remains_idempotent_and_retains_outputs() {
  local case_dir="$SANDBOX/idempotent"
  local fixture_root="$case_dir/sentei-vhs-progress-arc"
  make_test_setup "$case_dir"

  "$case_dir/setup-fixture.sh" >/dev/null
  printf 'retained\n' >"$fixture_root/outputs/existing.gif"
  printf 'stale\n' >"$fixture_root/logs/stale.log"
  "$case_dir/setup-fixture.sh" >/dev/null

  [[ -f "$fixture_root/outputs/existing.gif" ]] || fail 'outputs were not retained'
  [[ ! -e "$fixture_root/logs/stale.log" ]] || fail 'reset child was not recreated'
  [[ -f "$fixture_root/seed/README.md" ]] || fail 'fixture was not recreated'
}

test_symlinked_root_is_rejected_without_touching_target
test_retained_symlinks_are_rejected_without_touching_targets
test_symlinks_beneath_retained_directories_are_rejected
test_non_owned_root_is_rejected_before_reset
test_valid_fixture_remains_idempotent_and_retains_outputs
printf 'setup-fixture safety tests passed\n'
