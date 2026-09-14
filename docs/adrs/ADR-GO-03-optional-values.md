# ADR-GO-03: Optional Values

## Status

Accepted — 2026-08-16

## Context

Go has no way to say "this value may be absent" without either overloading the zero value or reaching
for a pointer. `""` might be an empty middle name or no middle name; `0` might be a real quantity or
an unset one. The usual workaround, `*string`, encodes absence honestly but pays for it with nil
dereferences and a heap allocation per optional field.

`samber/mo` supplies the functional container types — `Option`, `Result`, `Either`, and friends. They
are one library but not one decision: `Option` fills a gap the language leaves open, while `Result`
competes with an idiom Go already has and a syntax it lacks.

## Decision

### Adopt `mo.Option` for absence

- Model an optional value as **`mo.Option[T]`**, not `*T` and not a sentinel zero value.

  ```go
  type Customer struct {
      ID         string            `json:"id"`
      MiddleName mo.Option[string] `json:"middle_name"`
  }
  ```

- Read it with the comma-ok form, which matches how a map read already looks:

  ```go
  if name, ok := c.MiddleName.Get(); ok { ... }
  c.MiddleName.OrElse("(none)")
  ```

- `mo.Some("")` is **present with an empty value**, distinct from `mo.None[string]()`. That
  distinction is the whole point; do not collapse it back into a zero-value check.
- It marshals correctly in both directions — `None` becomes `null` and round-trips back to absent — so
  it is usable on DTOs and not only on in-memory types.
- Pointers remain correct for genuinely optional *references* and for large structs you do not want
  copied. `Option` is for optional **values**.

### Reject `mo.Result` for errors

- Errors stay `(T, error)`, wrapped with `%w`, inspected with `errors.Is` and `errors.As`, and
  defined as package-level sentinels or typed errors in `domain/`.

Three reasons, in order of weight.

Go has no `?` operator, so propagating a `Result` is more verbose than `if err != nil`, not less.
The payoff that makes `Result` worth its ceremony in Rust does not exist here.

Chaining is not expressible. A `FlatMap` turning `Result[T]` into `Result[U]` cannot be a method,
because Go rejects it outright:

```
syntax error: method must have no type parameters
```

So transformations become nested package-level calls rather than a fluent chain — the shape that
makes the type attractive is exactly the shape the language forbids.

And it leaves the ecosystem behind. `errors.Is`, `errors.As`, `%w` wrapping, and every library that
returns `(T, error)` assume the standard idiom; a `Result` boundary means translating at every edge.

### Reactive streams are out of scope

`samber/ro` (ReactiveX for Go) is **not adopted**. It is pre-1.0, it layers a second concurrency
model over channels and goroutines, and its operators are free functions returning transformers —
`func All[T any](func(T) bool) func(Observable[T]) Observable[bool]` — for the same reason `Result`
cannot chain. A project doing genuine stream processing with multiplexed sources and time-window
operators may adopt it in its own ADR; nothing in this profile assumes it.

## Consequences

- One import (`samber/mo`) enters the domain layer. It is a small, dependency-light library, and the
  types it adds are values with no framework attached.
- Optional fields stop allocating and stop being nil-checkable, which removes a class of panic from
  `domain/` entirely.
- Error handling stays boring and idiomatic, so every Go developer and every linter already
  understands it.
- Mixing the two styles is the failure to watch for: an `Option` returned where an `error` belongs
  discards the reason for absence. If a caller needs to know *why* something is missing, it is an
  error, not a `None`.

## Related

- ADR-BASE-01 — Clean Architecture (`domain/` is where these types are modelled)
- ADR-GO-01 — Dependency injection (the other stance this profile takes on modelling)
