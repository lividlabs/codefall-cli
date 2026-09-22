<!-- BEGIN CODEFALL TESTING -->
## Testing

Test cases live under `{{TESTING_ROOT}}/`. `{{TESTING_ROOT}}/AGENTS.md` names the runners, their
commands, and what the environment needs.

- A case is written through `/codefall-implement`, from the task's approved criteria, before the
  code.
- Cases run through `/codefall-test`.
- The harness is set up through `/codefall-equip`, in its own pull request; until then a task that
  needs a case stops and says so.

Instructions from the user or elsewhere in this file take precedence over this section.
<!-- END CODEFALL TESTING -->
