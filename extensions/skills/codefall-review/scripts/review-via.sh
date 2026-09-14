#!/usr/bin/env bash
#
# Run a review through another harness, headless and read-only.
#
# codefall-review writes the prompt — the target, the lenses, the calibration
# rules, and the findings schema. The reviewer runs inside the repository and
# reads the files itself, so nothing is copied for it. This script only starts
# the harness and puts its final message in a file. It parses nothing and
# decides nothing: what the reviewer said is the skill's to read.
#
# Every harness runs in its own read-only mode. A reviewer does not edit, and a
# subprocess that edited would bypass the host's PreToolUse hooks and its
# checkpoints, so nothing it did would be guarded or reversible.
#
# Usage: review-via.sh [--model <model>] <harness> <prompt-file> <schema-file> <out-file>
#
#   --model      the model to use; omitted, the harness picks its own
#   harness      codex | claude | opencode | gemini
#   prompt-file  the review prompt, already written
#   schema-file  findings.schema.json, beside this script's parent
#   out-file     where the harness's final message is written
#
# The prompt carries the schema for every harness. Codex and Claude Code also
# take it as a flag, which constrains their output instead of requesting it; the
# other two have no equivalent and rely on the prompt alone.
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

readonly USAGE="usage: review-via.sh [--model <model>] <codex|claude|opencode|gemini> <prompt> <schema> <out>"

fail() {
  local code=$1
  shift
  printf 'review-via: %s\n' "$*" >&2
  exit "$code"
}

# The model is a flag rather than a positional, so leaving it out is how the
# harness's own default is chosen. A positional would need a placeholder in the
# slot, and a reader would have to open this file to learn what the placeholder
# meant.
model=""

while [ "$#" -gt 0 ]; do
  case $1 in
    --model)
      [ "$#" -ge 2 ] || fail 64 "--model needs a value ($USAGE)"
      model=$2
      shift 2
      ;;
    --model=*)
      model=${1#--model=}
      shift
      ;;
    --)
      shift
      break
      ;;
    -*)
      fail 64 "unknown flag \"$1\" ($USAGE)"
      ;;
    *)
      break
      ;;
  esac
done

if [ "$#" -ne 4 ]; then
  fail 64 "$USAGE"
fi

harness=$1
prompt=$2
schema=$3
out=$4

case $harness in
  codex | claude | opencode | gemini) ;;
  *) fail 64 "unknown harness \"$harness\" ($USAGE)" ;;
esac

[ -r "$prompt" ] || fail 64 "cannot read prompt $prompt"
[ -r "$schema" ] || fail 64 "cannot read schema $schema"

command -v "$harness" >/dev/null 2>&1 || fail 69 "$harness is not on PATH"

timeout_seconds=${CODEFALL_REVIEW_TIMEOUT:-900}

# The harness may write out-file itself (codex) or have its stdout redirected
# into it (the rest), so the directory has to exist either way.
out_dir=$(dirname "$out")
mkdir -p "$out_dir" || fail 73 "cannot create $out_dir"
: >"$out" || fail 73 "cannot write $out"

# Building the flag as an array keeps the no-model case from becoming an empty
# argument, which some CLIs read as a model named "".
#
# Every expansion below is written `"${model_flag[@]+"${model_flag[@]}"}"` rather
# than `"${model_flag[@]}"`. Bash before 4.4 — which is what macOS ships as
# /bin/bash — treats an empty array as unset under `set -u` and aborts, and this
# script is run by whichever bash is first on PATH.
model_flag=()
if [ -n "$model" ]; then
  model_flag=(--model "$model")
fi

# --output-schema constrains the final message to the findings shape.
# --output-last-message is codex writing its own output, which the read-only
# sandbox does not govern: the sandbox covers the model's tool calls, not the
# program's plumbing.
run_codex() {
  codex exec \
    "${model_flag[@]+"${model_flag[@]}"}" \
    --sandbox read-only \
    --output-schema "$schema" \
    --output-last-message "$out" \
    - <"$prompt"
}

# Plan mode is Claude Code's read-only mode. Text output puts the final message
# on stdout with no envelope, which is what keeps this script free of a JSON
# dependency, and --json-schema constrains that message to the findings shape.
#
# Two differences from codex's --output-schema. The flag takes the schema itself
# rather than a path to it, so the file is read in here; and its validator does
# not resolve the 2020-12 meta-schema, rejecting the document outright while
# `$schema` and `$id` are present — so those two lines are dropped first. They
# identify the published artifact and say nothing about the shape, so a schema
# without them constrains exactly the same output.
run_claude() {
  claude --print \
    "${model_flag[@]+"${model_flag[@]}"}" \
    --permission-mode plan \
    --output-format text \
    --json-schema "$(sed -e '/^  *"\$schema": /d' -e '/^  *"\$id": /d' "$schema")" \
    <"$prompt" >"$out"
}

# OpenCode's `plan` agent is its read-only one. It has no schema flag, so the
# copy of the schema in the prompt is all it gets.
run_opencode() {
  opencode run \
    "${model_flag[@]+"${model_flag[@]}"}" \
    --agent plan \
    --file "$prompt" \
    "Review as the attached prompt says. Reply with findings JSON only." >"$out"
}

# Gemini appends -p to whatever arrived on stdin, so the prompt file is the
# input and -p is the instruction that closes it. No schema flag here either.
run_gemini() {
  gemini \
    "${model_flag[@]+"${model_flag[@]}"}" \
    --approval-mode plan \
    --output-format text \
    -p "Review as the text above says. Reply with findings JSON only." \
    <"$prompt" >"$out"
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
