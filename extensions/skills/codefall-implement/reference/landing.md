# Landing strategies

How work reaches `main`. Read at step 4, when classifying the run.

## The three strategies

The graph's shape picks one; the user can overrule it at the go gate. The project's `AGENTS.md`
constrains the choice before either speaks, `CUSTOMIZE.md` tunes it within those constraints, and a
conflict between the two is drift — show the difference and ask, never silently pick.

| Strategy | When | Shape |
| --- | --- | --- |
| **Parallel stacks** (default) | Independent chains with disjoint predicted file scopes | Each chain stacks toward `main`; chains run in parallel |
| **Single stack** | Overlapping file scopes, uncertainty, or fan-in without a need for parallelism | The whole graph in topological order, one branch atop another |
| **Epic branch** | Fan-in across chains, or work that must not land on `main` in increments — *and* parallelism matters | Workers branch off `epic/<id>-<slug>`, PRs target it, one aggregate PR to `main` |

- **There is no depth cap.** Depth alone never forces the epic branch; a deep stack's cost is
  disclosed at the go gate, not capped against.
- **Any DAG serializes into a single stack.** Topological order makes one branch-atop-branch chain
  out of any graph. That is the fallback whenever parallel stacks cannot be shown safe.
- **Fan-in is the epic branch's real trigger.** A bead whose parents sit on different branches has
  no single stack base. Run it serially as one topological stack, or give the chains an integration
  branch — never pause mid-run waiting for a human to merge the parents.

## Predicting whether stacks can run in parallel

Parallel stacks require **predicted-disjoint file scopes**, derived before the go gate: each chain's
Design refs name components, components own files, and disjoint components usually mean disjoint
files.

**Hotspot files count as overlap by default.** Lockfiles, generated artifacts, migrations, shared
type barrels, route and DI registries — cross-cutting files are where parallel work collides. Two
chains that both touch one are serialized unless the overlap can be shown harmless. The burden runs
toward serializing: a single stack is always correct.

**Prediction is verified at integration, never trusted.** After building completes, restack the
stack bottoms onto current `main` and re-run verification before reporting the merge order. A
conflict found there is resolved deliberately, in the open — not by an automated fix-up.

## The diagrams

The go gate renders the plan as a branch diagram built from the actual graph. Stacked:

```
main ──┬── bd-a1 ── bd-a2 ── bd-a3 ── bd-a4    stack A · parser chain    [src/parse/**]
       └── bd-b1 ── bd-b2                      stack B · CLI chain       [src/cli/**]
             2 stacks in parallel · merges drain bottom-up · A and B independent
```

Epic branch:

```
main ── epic/bd-e7-stage-context
          ├── bd-t1 ─┐
          ├── bd-t2 ─┼── bd-t4    fan-in: t4 needs t1 + t2
          └── bd-t3 ─┘            wave 1: t1 t2 t3 · wave 2: t4
```
