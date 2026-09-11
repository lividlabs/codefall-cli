#!/usr/bin/env bash
# PreToolUse guard: deny any Bash command that would merge or push to the default branch.
#
# A human performs every merge to main; no codefall verb ever does. A denial from this
# hook is the system working as designed. Exit 2 blocks the tool call; stderr is shown
# to the model. Exit 0 raises no objection; the normal permission flow still applies.
#
# Deliberate limits: this inspects the command string plus, for `gh pr merge`, the PR's
# actual base branch. It prefers a rare false denial over a false allow, and it is a
# guard, not the only line — a repository ruleset protecting the default branch remains
# the backstop.
set -u

input=$(cat)
cmd=$(printf '%s' "$input" | python3 -c \
  'import json,sys; print(json.load(sys.stdin).get("tool_input",{}).get("command",""))' \
  2>/dev/null) || exit 0
[ -z "$cmd" ] && exit 0

deny() {
  echo "codefall: $1 A human performs every merge to '$2' — report the merge order and stop." >&2
  exit 2
}

# The protected branch: origin's default, else main.
protected=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||')
[ -z "$protected" ] && protected=main

# --- gh pr merge: the PR's base decides; an undetermined base is denied, not allowed. ---
if printf '%s' "$cmd" | grep -qE '(^|[;&|[:space:]])gh[[:space:]]+pr[[:space:]]+merge([[:space:]]|$)'; then
  ref=$(printf '%s' "$cmd" | awk '
    {for (i = 2; i <= NF; i++)
       if ($(i-1) == "pr" && $i == "merge") {
         for (j = i + 1; j <= NF; j++) if ($j !~ /^-/) { print $j; exit }
       }}')
  base=$(gh pr view $ref --json baseRefName --jq .baseRefName 2>/dev/null)
  if [ -z "$base" ] || [ "$base" = "$protected" ]; then
    deny "denied: 'gh pr merge' into '$protected' (or into a base this hook could not determine)." "$protected"
  fi
  exit 0  # merge into an epic branch or another non-default base: allowed
fi

current=$(git branch --show-current 2>/dev/null)

# --- git merge while standing on the protected branch. ---
if printf '%s' "$cmd" | grep -qE '(^|[;&|[:space:]])git[[:space:]]+merge([[:space:]]|$)'; then
  [ "$current" = "$protected" ] && \
    deny "denied: 'git merge' while on '$protected'." "$protected"
fi

# --- git push with a refspec naming the protected branch (origin main, HEAD:main). ---
if printf '%s' "$cmd" | grep -qE '(^|[;&|[:space:]])git[[:space:]]+push([[:space:]]|$)'; then
  if printf '%s' "$cmd" | grep -qE "[[:space:]:]${protected}([[:space:]]|\$)"; then
    deny "denied: 'git push' targeting '$protected'." "$protected"
  fi
  # A bare push (no refspec) while standing on the protected branch pushes it.
  if [ "$current" = "$protected" ] && \
     printf '%s' "$cmd" | grep -qE '(^|[;&|[:space:]])git[[:space:]]+push([[:space:]]+-[[:alnum:]=-]+)*([[:space:]]+[[:alnum:]._-]+)?[[:space:]]*$'; then
    deny "denied: bare 'git push' while on '$protected'." "$protected"
  fi
fi

exit 0
