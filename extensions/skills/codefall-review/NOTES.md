# codefall-review — where it came from

Why `SKILL.md` looks the way it does: what was taken from elsewhere, what was rejected, and what was
tried and dropped. None of this is instruction — the skill is the instruction. This exists so nobody
re-adds something that was removed on purpose.

## Taken

**OpenCode's built-in `/review`** — target inference from the argument's shape, the rule that diffs
alone are not enough, and most of Calibration: be certain, no hypothetical edge cases, state the
conditions, do not be a zealot about style, no flattery. That template is the best short statement
of review discipline in the open and there was no reason to write a worse one.

What it does not have, and this skill added: document targets, the lens decomposition, structured
findings, and everything after the review — triage, fixes, a file.

**Anthropic's `pr-review-toolkit`** — one pass per question rather than one reviewer looking for
everything, and the lenses its agents cover: silent failures, test coverage, type design, comment
accuracy, simplification. Dropped from it: the "Strengths" and "Positive Observations" sections,
which are flattery with a heading.

**Cross-model review wrappers** — shelling out to another harness's headless mode rather than trying
to switch providers inside the host. No harness offers a per-subagent provider switch, and a
session-wide base URL override changes the host model too, so a subprocess is the only way.

**`codefall-graft`** — the report as a deliverable, and stopping after it.

## Rejected

**An external reviewer that edits.** Codex and the rest can write, and some cross-model tools let the
second model fix what it found. A subprocess writing files bypasses the `PreToolUse` hooks that are
codefall's enforcement and the checkpoints that make a session reversible, so a change it made would
be neither guarded nor undoable. Every external reviewer runs read-only.

**The reviewer fixing what it found.** The point of a second reading is different blind spots. A
model that finds and fixes grades its own work on the next pass. The reviewer is a subagent or
another harness; the session triages and fixes.

**A lens filter as a skill parameter.** `focus=<lens,…>` meant getting the argument right before
seeing what it would have done. The confirmation step names the lenses that will run and asks what
to drop, which puts the choice where the information is.

## Tried and dropped

**Rounds.** An earlier draft numbered reviews per target — `round-001`, `round-002` — reconciled the
previous round's findings against the current code, and marked them `resolved` or `stale`. It bought
stable finding ids, a `status` lifecycle, staleness rules, and reconciliation logic. All of that
only matters when findings sit unresolved between reviews, and they do not: a review ends with every
finding fixed, dismissed, or deferred. One invocation, one file, and a later review of the same
thing is a new review.

**Review as a read-only verb.** The first draft changed nothing at all and left every fix to the
user. It kept the reviewer honest by construction, and it meant a review of twelve typos produced
twelve things to type by hand. The subagent boundary gets the same separation without that cost.

**Inline `> [!REVIEW]` callouts in documents.** Document findings were going to be written into the
document beside the passage they were about. That is an edit to the reviewed artifact made before
anyone agreed to it. Findings go in the findings file, and accepted ones become real edits after
triage.

**Filing findings as beads.** A finding that outlives the review would have become a
`discovered-from` bead. The graph is `codefall-design`'s, and a review that quietly adds work to it
is making a decision that is not its own. Deferred findings stay in the file.

**Specific commits, merged pull requests, merged branches, superseded ADRs.** All were targets in the
first draft. Reviewing history produces findings about code that has moved on, and there is nowhere
for a fix to land. Every target is something live.

**A self-contained review packet.** The external reviewer was going to be handed a copy of
everything it needed — the diff, the full files, the conventions, the upstream document — because a
subprocess starts with no context. It runs inside the repository instead and reads what it needs,
which is less to build and less to go stale.
