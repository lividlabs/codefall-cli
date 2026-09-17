# codefall-equip — where it came from

Why `SKILL.md` looks the way it does: what was taken from elsewhere, what was rejected, and what was
tried and dropped. None of this is instruction — the skill is the instruction. This exists so nobody
re-adds something that was removed on purpose. The reasoning behind the contract itself is
ADR-005 in the repository's `docs/adrs/`.

## Taken

**A per-project sync skill from an earlier toolkit** — the observation that pulling `main` has
consequences the pull does not perform, and that a teammate who does not read diffs needs a single
thing to run. That skill knew one project's package manager, migration tool, and containers. What
this skill keeps is the job; what it drops is knowing the project, which moved into scripts the
project owns.

**`codefall-scaffold`'s stack question** — name what the evidence already implies and confirm it,
rather than asking a question the repository could have answered. Here that is "search first, then
ask with evidence".

**`codefall-graft`'s report-then-take** — a candidate is shown with what is wrong with it before
anything is declared, and the user takes it or not.

**`codefall-implement`'s definition of done** — a change that makes the scripts stale changes the
scripts in the same pull request, the same way tests and docs are part of done rather than a
follow-up.

## Rejected

**Doing the environment work in the skill.** Re-deriving what to run from the repository on every
invocation varies between runs, spends a model turn on what a lockfile already says, and cannot be
run by a person, by CI, or by a harness without the extension.

**Generating the scripts at run time.** Every cost of the above, plus a file that changes under the
project with no diff anyone reviewed.

**A one-time onboarding verb.** First draft and revision are the same procedure — read the
repository, read the scripts if they exist, propose, confirm, write, declare — so a verb named for
running once would have run every time a tool was introduced. This skill is named for what it does
to the repository rather than for when.

**`codefall-refresh` drafting on its first run.** A verb that runs daily and also does one-time
inference is two verbs. It also put the drafting interview in front of the one person least placed
to answer it: a teammate running refresh because the section in `AGENTS.md` said to.

**A template per stack.** A signals table covers more ground with less to maintain: a project is a
set of signals, not a stack, and a Go API with a Prisma-managed database is two rows, not a profile
nobody wrote.

**Routing the script through `codefall-graft`.** Graft moves documents against templates the
extension ships, with provenance. The scripts have no template behind them once drafted — they are
the project's — and this skill is already the verb that revises them, so graft would be a second
owner.

## Why the rules are shaped this way

**Never destructive.** A script that resets to reach a known state is not idempotent; it is
starting over. "Bring the project up to date?" has to be a question a teammate can always answer
yes to, and that holds only when the answer never costs them data.

**Proving is offered, not done.** `start` brings real services up on the user's machine. That is
the one thing in this skill that reaches outside the repository, so it is a yes-or-no.

**Two calls of `update` are the proof.** The second call exiting `0` quickly is the contract made
observable: cheap when nothing changed, and safe to run again.
