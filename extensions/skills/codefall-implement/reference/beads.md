# Beads: the commands a run issues

Every `bd` command in a run is the root session's, executed against the primary checkout — with
`-C` when the session's working directory is elsewhere. The default beads setup is an embedded Dolt
database, one writer at a time, whose sync channel is the repo's own git remote under
`refs/dolt/data`. Read at step 5, before the first claim.

## Session start

```bash
bd dolt pull            # teammates' claims and closes, when a remote is wired
bd gate check           # resolve gates for PRs merged since last session
bd ready --mol <epic>   # the claimable frontier
```

A project with no Dolt remote works identically; state is this-machine-only. Say so once; `bd dolt
pull` failing for lack of a remote is that statement's trigger, not a stop.

`bd gate check` fails safely. With `gh` missing, unauthenticated, or no GitHub remote, every gate
stays open, the command reports per-gate errors, and it still exits 0: a report, never a falsely
resolved gate. An `ESCALATE` line is different — the gate was checked and its PR is missing — and
is surfaced to the user.

## Claim, work, close

```bash
bd update <epic> <bead> --claim     # epic scope: the epic itself is claimed on the run's first claim
bd dolt push                        # the claim is visible to the team before the work, not after
```

Single-bead scope claims only its bead — the epic stays unclaimed so coworkers can take siblings —
and creates no landed bead; the PR-link comment is the merge trail.

Work happens in git; commits carry the bead ID — `feat: add StageContext type (bd-unz)`.

```bash
bd comment <bead> "PR #101 · feat/bd-unz-stage-context · built with opus/high"
bd close <bead> -r "done: criteria R1,R2 verified, checks green, PR #101 open" --suggest-next
bd dolt push
```

**The close is the engine.** Closing a bead removes it as a blocker, so `--suggest-next` prints the
next wave straight from the graph. **Closed means done, not merged**; the merge is tracked by the
gates below.

## The landed bead and its gates

An epic cannot be gated directly and refuses to close while children are open. So an epic run's
first claim also creates one extra child task — `Land: all PRs merged to main` — with gates
blocking it:

```bash
bd gate create --type=gh:pr --blocks <land-bead> --await-id=<pr-number> -r "PR #<n>"
```

`bd close` refuses an issue with unsatisfied gates, so the landed bead cannot close — and therefore
the epic cannot close — until every PR is merged. A later session's `bd gate check` clears the
gates as merges land. **Epic closed always means the code is on `main`.**

**Which PRs get gates depends on the strategy.** A stacked run gates every PR as it opens. An
epic-branch run gates none of its worker PRs — they target the epic branch and the root merges them
at the wave boundary; its one gate is created at integration, for the aggregate PR. A standalone
bead gets no gate.

## Discovered work

```bash
bd create "Parser drops trailing comma" --deps discovered-from:<bead> -p 2
bd dolt push
```

Workers report discoveries in their result JSON; the root files them, and pushes as after every
other write.

## Session end

Final `bd dolt push`. Every write already landed when its command ran; the database is the handoff,
and the next session opens with `bd ready`.
