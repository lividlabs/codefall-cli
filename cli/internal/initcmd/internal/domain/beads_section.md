<!-- BEGIN CODEFALL BEADS -->
## Beads

This project tracks work in Beads (`bd`). `bd prime` prints the full workflow and command
reference; Claude Code gets it at session start through the hook in `.claude/settings.json`.

Issues live in a local Dolt database under `.beads/`. The team's copy is `refs/dolt/data` on the
git remote, and `bd dolt pull` and `bd dolt push` are what move it; a git pull or push moves
nothing in it. `.beads/issues.jsonl` is an export, never the copy that is synced.

- Track every task in `bd`, never in TodoWrite, TaskCreate, or a markdown TODO list.
- `bd dolt pull` before reading the graph at the start of a session. `/codefall-refresh` does it
  when it runs.
- `bd ready` before asking what to work on; `bd update <id> --claim` before starting;
  `bd close <id>` when it is done.
- `bd dolt push` after every write — a create, a claim, a close, a dependency change — and
  before the work that follows it, so a teammate sees the claim and not only the result.
- Work discovered along the way becomes a linked issue:
  `bd create "<title>" --description "<context>" --deps discovered-from:<id>`.
- A project with no Dolt remote: `bd dolt pull` fails with `no remote`, and `bd dolt push` says it
  skipped. Say once that the database is this machine's and carry on. `bd dolt push --yes` adopts
  the git origin as the Dolt remote, and running it is the user's decision.
- `.beads/interactions.jsonl` is bd's append-only interaction log, off unless `audit.enabled` in
  `.beads/config.yaml` says otherwise. When it is on, the file is committed and comes along with
  whatever commit follows a bead write; that is expected, not a stray change. `.gitattributes`
  merges it by union, so never resolve a conflict in it by hand, never edit it, and never read it
  as part of a change under review.

### Quick reference

```bash
bd dolt pull             # the team's claims and closes, first
bd ready                 # unblocked work
bd show <id>             # one issue, with its dependencies
bd update <id> --claim   # take it
bd close <id>            # finish it
bd dolt push             # after every write
```

### Ending a session

1. File an issue for anything left undone or found along the way.
2. Run the repository's checks if code changed.
3. Close what is finished; update what is not.
4. `bd dolt push`, so the next session on any machine opens on this state.
5. Report what changed, what was verified, and which issues moved. Do not commit or push git
   unless the current instructions say to.

Instructions from the user or elsewhere in this file take precedence over this section.
<!-- END CODEFALL BEADS -->
