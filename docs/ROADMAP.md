# Roadmap

The objective of this plugin is to create skills to streamline the operations of
a software factory with an opinionated approach. Future versions will allow for
customization but the first version will be a fixed approach.

## Process via Skills

The process resembles the waterfall process of yore but the ability it iterate on functionality end-toend within hours instead of months.

 - **conceptualize** an idea into a short document that says what the problem is and why it matters, before anyone specifies or scaffolds it; concepts live in the repo at `docs/concepts/CONCEPT-NNN-slug.md`. `scaffold` requires one — a requirement the user can break, recorded in the decision log when they do — because its worst failure is picking architecture defaults from a one-sentence description. `specify` may draw on one and never requires it.
 - **scaffold** the architecture and this process
 - **specify** requirements for a feature and their acceptance criteria; the specification is a document in the repo at `docs/specs/SPEC-NNN-slug.md`, holding one or more requirements, each with a user story and criteria written in EARS notation. The document is canonical and is mirrored to an issue tracker — a parent issue for the spec, a child issue per requirement — so people can see what is ready, in progress, and done. The first version mirrors only to GitHub Issues; Jira and Linear are profiles that come later.
 - **mock-up** the visual surface of a feature; runs before or after specify. It imports mockups exported from a design tool, or helps create one when the user doesn't have or want a design tool. Mockups land in `docs/mockups/<slug>/` and the specification references them; an issue labelled `requires-mockup` blocks design until the mockup exists.
 - **design** a system to implement a feature; the output is a design document in the repo at `docs/designs/DESIGN-NNN-slug.md` and a work breakdown whose tickets live inside of beads. The document is scaled to the change — a contained fix skips it entirely and gets beads only. It determines dependencies on the tickets it creates as well as any potential dependencies on existing tickets, and inserts the work into the graph. Choices that are hard to reverse become ADRs in the project's own sequence.
 - **implement** a feature following the dependency graph and executing work in parallel waves where possible; every test the current work needs — unit, integration, and end-to-end, planned by design or discovered mid-task — is part of the definition of done. It also owns the concept transition to `Active` — a concept is Active once work has started against it, and no earlier skill can observe that moment.
 - **review** code and tests with or without another model or harness
 - **test** the completed system after work lands: regression passes, retesting in a fresh context, coverage campaigns, and agentic driven tests. The boundary with `implement` is time and scope — implement's tests prove the change being built; `test` re-proves the system.

 <insert a diagram?>

 ## Roadmap

 - [x] scaffold skill for Typescript projects: Node.js, React, and/or Next.js
 - [x] conceptualize skill: concept documents in `docs/concepts/`, required by scaffold, optional for specify
 - [ ] scaffold skill for Golang projects
 - [x] specify skill: spec documents in `docs/specs/` with EARS acceptance criteria, mirrored to Github Issues
 - [x] mock-up skill: import from a design tool, or author one when there isn't one
 - [ ] specify tracker profiles for Jira and Linear
 - [ ] tracker resolution for specify: read the destination from a project-level record instead of asking each run
 - [x] design skill integrated with Beads, this requires a Beads configuration in the application
 - [x] implement skill integrated with Beads, including the concept transition to `Active`
 - [ ] test skill
 - [ ] scaffold skill for Flutter projects
