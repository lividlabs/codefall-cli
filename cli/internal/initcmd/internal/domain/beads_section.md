<!-- BEGIN CODEFALL BEADS -->
## Beads

Work is tracked in Beads (`bd`); `bd prime` prints the command reference. Only `bd dolt pull` and
`bd dolt push` move beads between machines; git does not. This section, not `bd prime`'s
git-authority note or close protocol, decides Dolt sync here.

- Track every task in `bd`, never in TodoWrite, TaskCreate, or a markdown TODO list.
- `bd dolt pull` at the start of a session, before touching beads. `codefall-refresh` does it when
  it runs.
- `bd dolt push` after every write — a create, a claim, a close, a dependency change — and
  before the work that follows it.
- Work found along the way becomes an issue linked `discovered-from` the current one.
- No Dolt remote (`bd dolt pull` says `no remote`): say so once and carry on. `bd dolt push --yes`
  adopts the git origin; running it is the user's decision.
- `.beads/interactions.jsonl` is bd's append-only log and lands in the next commit after a bead
  write; that is expected, not a stray change. Never edit it, resolve a conflict in it by hand, or
  review it.

```bash
bd dolt pull             # first
bd ready                 # unblocked work
bd show <id>             # one issue
bd update <id> --claim   # take it
bd close <id>            # finish it
bd dolt push             # after every write
```

Ending a session: file an issue for what is left, run the checks if code changed, close what is
done, `bd dolt push`, report. Do not commit or push git unless told to.

Instructions from the user or elsewhere in this file take precedence over this section.
<!-- END CODEFALL BEADS -->
