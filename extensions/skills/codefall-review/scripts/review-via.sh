#!/usr/bin/env bash
#
# Run a review packet through another harness, headless and read-only.
#
# codefall-review writes the packet — the diff or document, the full files, the
# conventions, the upstream artifact, the lenses, and the findings schema. This
# script only runs the harness and puts its final message in a file. It parses
# nothing and decides nothing: what the reviewer said is the skill's to read.
#
# Every harness runs in its own read-only mode. A reviewer does not edit, and a
# subprocess that edited would bypass the host's PreToolUse hooks and its
# checkpoints, so nothing it did would be guarded or reversible.
#
# Usage: review-via.sh <harness> <model|-> <packet-file> <schema-file> <out-file>
#
#   harness      codex | claude | opencode | gemini
#   model        the model to use, or - for the harness's own default
#   packet-file  the review packet, already written
#   schema-file  findings.schema.json, beside this script's parent
#   out-file     where the harness's final message is written
#
# Environment
#   CODEFALL_REVIEW_TIMEOUT   seconds before the run is killed (default 900)
#
# Exit codes
#   0   the harness ran and out-file holds its final message
#   64  the arguments are wrong
#   69  the harness is not on PATH
#   73  out-file could not be written, or the harness wrote nothing
#   75  the harness was still running at the timeout and was killed
#   varies  the harness's own non-zero exit, passed through

set -uo pipefail

readonly USAGE="usage: review-via.sh <codex|claude|opencode|gemini> <model|-> <packet> <schema> <out>"

fail() {
  local code=$1
  shift
  printf 'review-via: %s\n' "$*" >&2
  exit "$code"
}

if [ "$#" -ne 5 ]; then
  fail 64 "$USAGE"
fi

harness=$1
model=$2
packet=$3
schema=$4
out=$5

case $harness in
  codex | claude | opencode | gemini) ;;
  *) fail 64 "unknown harness \"$harness\" ($USAGE)" ;;
esac

[ -r "$packet" ] || fail 64 "cannot read packet $packet"
[ -r "$schema" ] || fail 64 "cannot read schema $schema"

command -v "$harness" >/dev/null 2>&1 || fail 69 "$harness is not on PATH"

timeout_seconds=${CODEFALL_REVIEW_TIMEOUT:-900}

# The harness may write out-file itself (codex) or have its stdout redirected
# into it (the rest), so the directory has to exist either way.
out_dir=$(dirname "$out")
mkdir -p "$out_dir" || fail 73 "cannot create $out_dir"
: >"$out" || fail 73 "cannot write $out"

# `-` means the harness picks. Building the model flag as an array keeps the
# empty case from becoming an empty argument, which some CLIs read as a model
# named "".
model_flag=()
if [ "$model" != "-" ]; then
  model_flag=(--model "$model")
fi

# Not every harness takes a JSON Schema, so the packet carries the schema inline
# for all of them and this line is the extra guarantee where one is available.
run_codex() {
  codex exec \
    "${model_flag[@]}" \
    --sandbox read-only \
    --output-schema "$schema" \
    --output-last-message "$out" \
    - <"$packet"
}

# Plan mode is Claude Code's read-only mode. Text output puts the final message
# on stdout with no envelope, which is what keeps this script free of a JSON
# dependency.
run_claude() {
  claude --print \
    "${model_flag[@]}" \
    --permission-mode plan \
    --output-format text \
    <"$packet" >"$out"
}

# OpenCode's `plan` agent is its read-only one. The packet is attached rather
# than passed as the message: a diff plus its full files runs past the argument
# length a shell will take.
run_opencode() {
  opencode run \
    "${model_flag[@]}" \
    --agent plan \
    --file "$packet" \
    "Review the attached packet and reply with findings JSON only." >"$out"
}

# Gemini appends -p to whatever arrived on stdin, so the packet is the input and
# the prompt is the instruction that closes it.
run_gemini() {
  gemini \
    "${model_flag[@]}" \
    --approval-mode plan \
    --output-format text \
    -p "Review the packet above and reply with findings JSON only." \
    <"$packet" >"$out"
}

# A review that hangs is the failure this guards against. There is no portable
# `timeout` on macOS, so the run goes to the background and is polled — and the
# whole process group is killed, because the harnesses spawn children of their
# own that outlive a signal sent to the parent alone.
run_with_timeout() {
  set -m

  "run_$harness" &
  local pid=$!
  local waited=0

  while kill -0 "$pid" 2>/dev/null; do
    if [ "$waited" -ge "$timeout_seconds" ]; then
      kill -TERM -"$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null
      sleep 2
      kill -KILL -"$pid" 2>/dev/null || kill -KILL "$pid" 2>/dev/null
      wait "$pid" 2>/dev/null
      return 75
    fi

    sleep 1
    waited=$((waited + 1))
  done

  wait "$pid"
}

run_with_timeout
status=$?

if [ "$status" -eq 75 ]; then
  fail 75 "$harness did not finish within ${timeout_seconds}s"
fi

if [ "$status" -ne 0 ]; then
  fail "$status" "$harness exited $status"
fi

if [ ! -s "$out" ]; then
  fail 73 "$harness exited 0 and wrote nothing to $out"
fi

exit 0
