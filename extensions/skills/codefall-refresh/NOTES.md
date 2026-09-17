# codefall-refresh — where it came from

Why `SKILL.md` looks the way it does: what was taken from elsewhere, what was rejected, and what was
tried and dropped. None of this is instruction — the skill is the instruction. This exists so nobody
re-adds something that was removed on purpose. The reasoning behind the contract it runs is
ADR-005 in the repository's `docs/adrs/`.

## Taken

**`git pull --ff-only`** — the one move the checkout is ever made with. A fast-forward cannot lose
anything and cannot produce a conflict, which is what makes it safe to run on a teammate's behalf.

**`codefall-implement`'s preflight reading** — the checkout lines come from the same shared script
every verb runs; this skill reads them rather than asking git the same questions a second way.

**Doctor's remedies** — a failure carries what to do next, in the same sentence, or nothing. The
failures table follows that shape: what failed, what it means, what to do.

## Rejected

**Drafting the scripts on the first run.** A verb that runs daily and also does one-time inference
is two verbs, and it puts the drafting interview in front of the teammate least placed to answer
it. `codefall-equip` drafts; this skill sends an unequipped project there.

**Rebasing a feature branch.** A rebase can conflict, and a conflict in a skill that a teammate
runs to get going is the opposite of the point. The branch is reported as behind and the rebase is
the user's.

**Stashing a dirty tree to pull.** A stash is state the user did not ask for, in a stack other
sessions may be using. A dirty tree on the default branch is reported and left alone, and the
environment is brought level anyway.

**Adding the `.gitignore` entry when the stamp is not ignored.** That is `codefall init`'s line to
write, and a verb that writes it too is a second owner of the same entry. Refresh says so and
names init.

**Editing a script that fails.** The scripts are the project's, revised by `codefall-equip` at the
point a change makes them stale. A refresh that patched a script would put a change nobody
reviewed into a file the whole team runs.

**A hook that pulls at session start.** A hook cannot ask, and a pull moves the working tree under
whatever the session is doing. The `SessionStart` slot can carry a notice; the action stays a
verb someone invokes.

## Why the rules are shaped this way

**`start` always, `update` by the stamp.** `start` is cheap when everything is up and `update`
assumes it ran, so skipping `start` saves nothing and risks `update` reaching a service that is
down. `update` is skipped only on the stamp's say-so, because the stamp is the one record of the
environment having matched this commit.

**The stamp is written after `update` and never otherwise.** Written before, a failed `update`
would leave a record claiming an environment that does not exist, and the next refresh would skip
the work that was needed.

**Always bring the environment level, even when the checkout is left alone.** A feature branch
that is behind still has a checkout with its own migrations and lockfile, and the environment
should match that checkout whether or not the branch moves. The half a pull leaves undone is the
whole reason the verb exists.
