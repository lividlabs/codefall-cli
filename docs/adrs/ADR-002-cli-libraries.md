# ADR-002: CLI Libraries — Cobra, Fang, and the Charm Ecosystem

## Status

Accepted — 2026-08-27

## Context

The decision log has carried "CLI framework" as Open since scaffold. ADR-GO-02 rule 5 already fixes
where such a library may live — `presentation/` and the composition root, never `domain/` or
`application/` — but not which one it is. The first component needs three things, distinct enough
that no single library provides all of them:

1. **Command structure** — a root command with subcommands, flags, help, `--version`, shell
   completions, and man pages. Every Go CLI library covers this part.
2. **Styled output** — colour, weight, and layout in human-facing output, falling back to plain text
   when `stdout` is not a terminal or `NO_COLOR` is set.
3. **Forms** — interactive prompts (select, input, confirm) for the cases where a command should ask
   rather than fail on a missing flag.

For command structure the candidates were the standard library's `flag`, `urfave/cli`,
`alecthomas/kong`, and `spf13/cobra`. `flag` has no subcommand model, so using it means writing the
framework by hand. `kong` is the strongest alternative — its struct-tag declarations are terser than
Cobra's — but nothing else pairs with it. `cobra` is the most widely used, has the largest body of
examples and tooling (completions, docs generation, POSIX flags via `pflag`), and is the only one the
Charm libraries target.

For output and forms the field is narrower. Charm's libraries — Lip Gloss for styling, Huh for
forms, Bubble Tea for full-screen terminal UIs, Fang as a Cobra wrapper — are designed to be used
together and share one colour-profile and terminal-detection layer. A second styling or prompt
library alongside them would mean two answers to "what does an error look like", which is an
inconsistency users notice.

**One distinction that is easy to get wrong:** the Charm libraries are mid-way through a major
version transition, and the two generations do not mix. Lip Gloss, Bubble Tea, Bubbles, and Huh have
`v2` majors whose module paths are `charm.land/<name>/v2`, not `github.com/charmbracelet/<name>`. Huh
v2 requires Bubble Tea v2 and Lip Gloss v2 directly. A v1 Huh form cannot be embedded in a v2 Bubble
Tea program, and a module that pulls in both generations carries two copies of each library. Fang is
neutral — it imports neither Lip Gloss major, rendering through `colorprofile` and `x/ansi` directly,
and its canonical path is still `github.com/charmbracelet/fang` — so it does not force the choice.
Everything else does.

## Decision

Stay in the Charm ecosystem, on its v2 generation, with Cobra underneath.

### Command structure — Cobra, run through Fang

- **`github.com/spf13/cobra`** defines the command tree. Flags are `pflag`, Cobra's own; the standard
  library's `flag` package is not used alongside it.
- **`github.com/charmbracelet/fang`** runs the root command: `fang.Execute(ctx, root, opts...)`
  replaces `root.Execute()`. Fang supplies the styled help and usage pages, styled errors, the
  `--version` flag (from build info by default), the `completion` subcommand, and a hidden `man`
  subcommand. Cobra's help and version templates are left alone: Fang owns that rendering, and its
  `fang.With*` options are the way to change it.
- The command tree is a `presentation/` artefact. Each component builds its commands in its own
  `presentation/` package and exposes them through its facade; the composition root
  (`cmd/codefall/main.go`) mounts them on the root command and calls Fang. `main` is the only
  importer of Fang. The exact shape of the facade function is settled with the first component.
- A `*cobra.Command` crossing a facade is a delivery type, not a domain entity, so ADR-001 does not
  apply to it.

### Styled output — Lip Gloss, as needed

- **`charm.land/lipgloss/v2`** for any styling of human-facing output. No hand-written ANSI escape
  sequences.
- Plain text is the default. A command's output is written unstyled first; style is added where it
  makes the output easier to read — a heading, an error, a status word — not because it is available.
- Output meant to be parsed — anything a `--json`-style flag produces, anything designed to be piped
  — is never styled.
- Styled output must degrade to plain text on a non-terminal `stdout` and under `NO_COLOR`. Lip
  Gloss v2 does not do this at render time as v1 did: `Style.Render` emits full ANSI, and the
  downsampling happens in the writer. Human-facing output therefore goes through a
  `colorprofile`-aware writer — `lipgloss.Fprint*` builds one per call, or `colorprofile.NewWriter`
  builds one to hold — and never by writing a rendered style straight to `os.Stdout`.

### Forms — Huh

- **`charm.land/huh/v2`** for interactive prompts: select, multi-select, text input, confirm, and
  grouped forms of those.
- A form is a `presentation/` concern. It collects values, validates them at the field level, and
  hands a contract to a use case. It holds no other logic.
- **A command that prompts by default also runs without prompting.** Each value its form asks for
  is also a flag. When the flags supply every required value the form does not run. When `stdin` is
  not a terminal and a value is missing, the command fails with an error naming the flag — it never
  blocks on input that cannot arrive. This is what keeps the tool usable from scripts, CI, and
  agents. The rule is scoped to commands that are interactive by default; a command that is
  non-interactive by default already satisfies it.
- Forms run with `RunWithContext` so a cancelled command cancels its prompt. `WithAccessible` follows
  the `ACCESSIBLE` environment variable, as Huh's own examples do.

### One generation, and no dependency before its code

- The v2 majors, on their `charm.land/…/v2` paths, are the ones this project uses. Nothing that
  imports the v1 paths (`github.com/charmbracelet/{lipgloss,bubbletea,bubbles,huh}`) is added — a
  library that does brings the second generation in as a transitive dependency.
- Consistent with the scaffold's "`go.mod` requires nothing": each library arrives with the change
  that first calls it. Cobra and Fang with the first command, Lip Gloss with the first styled line,
  Huh with the first prompt.

## Consequences

- **Bubble Tea is already in the graph.** Huh v2 requires `charm.land/bubbletea/v2` and runs every
  form as a Bubble Tea program. If this tool ever needs a full-screen or persistently interactive
  view, Bubble Tea is the library, and that is not a new decision — it is a consequence of this one.
  A `huh.Form` is itself a Bubble Tea model, so form work carries over rather than being rewritten.
  What would warrant a new ADR is the architectural question a TUI raises — where a long-lived view's
  state lives relative to `application/` — not the choice of library.
- **Fang owns the help, error, and version surface.** The trade is a consistent look with no
  template code in exchange for Fang's opinions: usage is not printed after a user error, errors are
  styled, `man` and `completion` are present. Overriding one is a `fang.With*` option
  (`WithoutManpage`, `WithoutCompletions`, `WithoutVersion`, `WithErrorHandler`, `WithTheme`), not
  a bypass of `fang.Execute`.
- **Two flag packages exist and only one is used.** Cobra uses `pflag`; the standard library's
  `flag` looks the same and is not compatible. A `flag.String` in a command file is a bug.
- **Forms are not unit-tested; the use cases behind them are.** Huh needs a terminal. Keeping forms
  to value collection means the logic they feed lives in `application/`, where hand-written fakes
  (ADR-GO-01) test it. Cobra commands themselves test without a terminal, via `SetArgs` and `SetOut`
  on the command.
- **Dependency weight.** The Charm graph is not small — Bubble Tea, Bubbles, `x/ansi`,
  `colorprofile`, `ultraviolet`, and their dependencies arrive with Huh. For a tool that ships one
  static binary this is acceptable; it would not be for a library.
- **Enforcement is already in place, and already unproven.** The strict allow-lists in
  `.golangci.yml` forbid every one of these packages in `domain/` and `application/` without naming
  them, because `list-mode: strict` denies whatever is not allowed. The existing gotcha applies:
  those rules match no files until the first component exists and must be proven against a
  deliberate violation then — `github.com/spf13/cobra` imported from `application/` is the natural
  one.
- **The v1/v2 split is a review item, not a build failure.** The day a v1-only dependency is added,
  `go mod graph` gains a `github.com/charmbracelet/lipgloss` or `bubbletea` line without `/v2` and
  nothing stops the build. Checking for it is part of reviewing any change to `go.mod`.

## Related

- ADR-GO-02 — Boundary enforcement (rule 5: no CLI framework in the inner rings)
- ADR-BASE-01 — Clean Architecture (`presentation/` is the delivery layer these libraries live in)
- ADR-GO-01 — Dependency injection (how commands reach their use cases; the composition root that
  calls Fang)
- ADR-001 — Facade contracts (what a form hands to a use case)
- Cobra — <https://github.com/spf13/cobra>
- Fang — <https://github.com/charmbracelet/fang>
- Lip Gloss v2 — <https://github.com/charmbracelet/lipgloss> (module `charm.land/lipgloss/v2`)
- Huh v2 — <https://github.com/charmbracelet/huh> (module `charm.land/huh/v2`)
- Bubble Tea v2 — <https://github.com/charmbracelet/bubbletea> (module `charm.land/bubbletea/v2`)
