<!-- BEGIN CODEFALL BEADS -->
## Beads

This project tracks work in Beads (`bd`). `bd prime` prints the full workflow and command
reference; Claude Code gets it at session start through the hook in `.claude/settings.json`.

- Track every task in `bd`, never in TodoWrite, TaskCreate, or a markdown TODO list.
- `bd ready` before asking what to work on; `bd update <id> --claim` before starting;
  `bd close <id>` when it is done.
- Work discovered along the way becomes a linked issue:
  `bd create "<title>" --description "<context>" --deps discovered-from:<id>`.

### Quick reference

```bash
bd ready                 # unblocked work
bd show <id>             # one issue, with its dependencies
bd update <id> --claim   # take it
bd close <id>            # finish it
```

### Ending a session

1. File an issue for anything left undone or found along the way.
2. Run the repository's checks if code changed.
3. Close what is finished; update what is not.
4. Report what changed, what was verified, and which issues moved. Do not commit, push, or run
   `bd dolt push` unless the current instructions say to.

Instructions from the user or elsewhere in this file take precedence over this section.
<!-- END CODEFALL BEADS -->
