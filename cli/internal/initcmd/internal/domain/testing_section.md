<!-- BEGIN CODEFALL TESTING -->
## Testing

This project's test cases live under `{{TESTING_ROOT}}/`, which is what `test.dir` in
`.codefall/settings.json` declares. `{{TESTING_ROOT}}/AGENTS.md` holds the facts only this project
can state: the runners, the command that runs every spec and the one that runs a single case, how a
run cleans up what it created, and what the environment needs.

- A case is written through `/codefall-implement`, from the criteria the task's plan approved, and
  before the code it covers.
- Cases and suites are run through `/codefall-test`.
- The test harness is set up through `/codefall-equip`, in a pull request of its own. Until it is,
  a task that needs a case stops and says so.

Instructions from the user or elsewhere in this file take precedence over this section.
<!-- END CODEFALL TESTING -->
