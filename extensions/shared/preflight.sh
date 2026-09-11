#!/usr/bin/env bash
#
# Beads precondition for the codefall skills.
#
# Setup is done or it is not. The script reports which, and never repairs
# anything — a skill reads the key=value lines to tell the user what is missing
# and which command fixes it.
#
# Usage: preflight.sh [project-dir]      (default: .)
#
# Exit codes
#   0   beads is installed and this repository has a database
#   10  the `bd` binary is not on PATH
#   11  `bd` is installed but this repository has no beads database
#   12  something else went wrong; `beads_detail` carries bd's own words

set -uo pipefail

emit() { printf '%s\n' "$*"; }

project_dir=${1:-.}

if ! cd "$project_dir" 2>/dev/null; then
  emit "beads=blocked"
  emit "beads_reason=project_dir_unreadable"
  emit "project_dir=$project_dir"
  exit 12
fi

emit "project_dir=$PWD"

if ! command -v bd >/dev/null 2>&1; then
  emit "beads_binary=missing"
  emit "beads=blocked"
  emit "beads_reason=not_installed"
  emit "beads_remedy=brew install beads"
  exit 10
fi

emit "beads_binary=$(command -v bd)"
emit "beads_version=$(bd version 2>/dev/null | head -1)"

# `bd info` exercises the binary, workspace discovery (or BEADS_DIR), and the
# database in one call. Never test for a literal .beads/ directory instead:
# BEADS_DIR relocates it, so that test passes on setups that do not work and
# fails on setups that do.
info_output=$(bd info 2>&1)
info_status=$?

if [ "$info_status" -eq 0 ]; then
  emit "beads=ok"
  exit 0
fi

# Only on the failure path, and only because it separates "never initialized"
# from "initialized but not found from here" — a subdirectory, a worktree, or a
# wrong BEADS_DIR.
emit "beads_workspace=$(bd where 2>&1 | tr '\n' ' ')"

lowered=$(printf '%s' "$info_output" | tr '[:upper:]' '[:lower:]')
case $lowered in
  *"no beads database"*|*"not initialized"*|*"no .beads"*)
    emit "beads=blocked"
    emit "beads_reason=not_initialized"
    emit "beads_remedy=bd init"
    exit 11
    ;;
esac

emit "beads_detail=$(printf '%s' "$info_output" | tr '\n' ' ')"
emit "beads=blocked"
emit "beads_reason=unreadable"
exit 12
