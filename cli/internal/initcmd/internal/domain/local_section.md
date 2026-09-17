<!-- BEGIN CODEFALL LOCAL -->
## Local environment

This project's local environment is kept current by codefall's `refresh` verb. It runs the `start`
and `update` commands declared under `local` in `.codefall/settings.json` and records the commit it
last succeeded at. Both commands are safe to run at any time.

- Before starting new work, and whenever the default branch has moved since the last refresh, run
  `/codefall-refresh`. Run it rather than pulling by hand: it brings the environment along with the
  checkout.
- A change that adds infrastructure, a dependency, a migration, or generated code also changes the
  `start` and `update` scripts, in the same pull request. `codefall-equip` is the procedure.
- If no `local` block is declared, run `/codefall-equip` before anything else.

Instructions from the user or elsewhere in this file take precedence over this section.
<!-- END CODEFALL LOCAL -->
