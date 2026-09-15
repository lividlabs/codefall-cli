# Plan

The following items are outstanding in order to complete the first verion of the tool.

## Skills
- [ ] `report` — report/file a bug; requires a slightly different process as compared to specifiying a requirement
  - **Open Questions**
    - Does the bug report would take the place of a spec? Likely yes.
    - Does the bug report have a github issue and a markdown report? Probably.
- [ ] `fix` — (under consideration) fix could potentially be a combination of design+implement but specifically for bugs reports
- [ ] `test` — run tests
  - unit, integration, and e2e
  - testing scope: full suite, a subset based on local/branch changes, a specific subset

## Command Line
- [ ] Set up `mise` config for the cli tool inside of init
- [ ] `create` command to start a new project

## Testing
- Test all of the skills against an actual project.
  - [x] `scaffold`
  - [x] `graft`
  - [ ] `conceptualize`
  - [ ] `mock-up`
  - [ ] `specify`
  - [ ] `design`
  - [ ] `implement`
  - [ ] `review`
  - [ ] `report`
  - [ ] `test`

## Completed

### Skills
- [x] `conceptualize` — capture an idea as concept documents
- [x] `mock-up` — create or import mockups
- [x] `specify` — write a spec with acceptance criteria
- [x] `design` — write a design doc and create the task graph
- [x] `implement` — build ready tasks and open pull requests
- [x] `review` — review work, triage findings, apply fixes
- [x] `scaffold` — start a new project with ADRs and AGENTS.md files
- [x] `graft` — update a project's docs to the current templates
- [x] Shorten every SKILL.md to the ADR-004 guidelines

### Command Line
- [x] `init` — survey settings, install skills and hooks for each harness; `upgrade` alias
- [x] `doctor` — check the directory and tools are ready

### Release
- [x] Cross-platform binaries via GoReleaser (macOS and Linux)
- [x] release-please changelog and versioning
- [x] Signed packslip with each release
