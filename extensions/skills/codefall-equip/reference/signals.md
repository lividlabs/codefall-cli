# Signals, entry points, and what they map to

What a repository says about the tools its local environment needs, where a project may already
keep the commands, and what each signal becomes in `start` and `update`. Read at step 2, and again
at step 3 while drafting.

## Contents

- [Entry points a project may already have](#entry-points-a-project-may-already-have)
- [Signals and what they map to](#signals-and-what-they-map-to)
- [What fails the contract](#what-fails-the-contract)

## Entry points a project may already have

Search these first. A hit is a candidate; read its body before offering it.

| Where | What to look for |
| --- | --- |
| `Makefile`, `GNUmakefile` | targets named `up`, `dev`, `dev-up`, `start`, `setup`, `bootstrap`, `sync`, `migrate`, `install` |
| `justfile`, `Justfile` | recipes with the same names |
| `package.json` `scripts` | `dev:up`, `db:up`, `setup`, `bootstrap`, `db:migrate`, `migrate`, `postinstall` that runs codegen |
| `Taskfile.yml`, `.mise.toml` `[tasks]` | tasks with the same names |
| `scripts/`, `bin/`, `script/` | `setup`, `bootstrap`, `dev`, `up`, `start`, `update`, `sync`, `migrate`, with or without `.sh` |
| `README.md`, `CONTRIBUTING.md`, `docs/` | a "getting started" or "local development" section that lists commands nobody scripted |
| `.codefall/settings.json` | a `local` block already declared — the revision case |

A README section that lists commands is not an entry point; it is the evidence for a draft.

## Signals and what they map to

One row per signal. A draft's `update` runs the row's update step for every signal present, in the
order the rows appear: dependencies before codegen, codegen before migrations. `start` runs every
row's start step.

| Signal | Tool | `start` | `update` |
| --- | --- | --- | --- |
| `compose.yaml`, `docker-compose.yml` | Docker Compose | `docker compose up -d --wait` (`up -d` when the file has no healthchecks) | — |
| `package-lock.json` | npm | — | `npm ci` |
| `pnpm-lock.yaml` | pnpm | — | `pnpm install --frozen-lockfile` |
| `yarn.lock` | Yarn | — | `yarn install --immutable` (Berry) or `yarn install --frozen-lockfile` (1.x) |
| `bun.lock`, `bun.lockb` | Bun | — | `bun install --frozen-lockfile` |
| `go.mod` | Go | — | `go mod download` |
| `uv.lock` | uv | — | `uv sync` |
| `poetry.lock` | Poetry | — | `poetry install --sync` |
| `requirements.txt` with a venv | pip | — | `pip install -r requirements.txt` inside the venv |
| `Gemfile.lock` | Bundler | — | `bundle install` |
| `Cargo.lock` | Cargo | — | `cargo fetch` |
| `mix.lock` | Mix | — | `mix deps.get` |
| `prisma/schema.prisma` | Prisma | — | `prisma generate`, then `prisma migrate deploy` |
| `drizzle.config.*` | Drizzle | — | `drizzle-kit migrate` |
| `knexfile.*` | Knex | — | `knex migrate:latest` |
| `alembic.ini` | Alembic | — | `alembic upgrade head` |
| `db/migrate/` with a `Rakefile` | Rails | — | `bin/rails db:migrate` |
| `manage.py` | Django | — | `python manage.py migrate` |
| `migrations/` with `goose` or `migrate` in `go.mod` or a Makefile | goose / golang-migrate | — | `goose up` / `migrate up` |
| `sqlc.yaml` | sqlc | — | `sqlc generate` |
| `codegen.ts`, `codegen.yml` | GraphQL Codegen | — | `graphql-codegen` |
| `openapi*.yaml` with a generator config | OpenAPI generator | — | the project's generate command |
| `buf.yaml`, `*.proto` with a build rule | protobuf | — | `buf generate` / the project's `protoc` rule |
| `.env.example`, `.env.sample` | dotenv | — | `cp .env.example .env` only when `.env` is missing |
| `mise.toml`, `.tool-versions` | mise / asdf | — | `mise install` — first, before anything that needs a runtime |

A signal with no row is still a signal: read the tool's documentation for its idempotent
install-or-sync form and its migrate-forward form, and use those.

## What fails the contract

A candidate or a draft with any of these is reported with the line and never declared:

- `db:reset`, `migrate reset`, `migrate fresh`, `drop`, `DROP DATABASE`, `flush`
- `rm -rf` on a data directory, a volume, or `node_modules` as a matter of course
- `docker compose down -v`, `--force-recreate`, `--renew-anon-volumes`
- `git clean`, `git reset --hard`, `git checkout .`, or any pull: the checkout is
  `codefall-refresh`'s, and the tree is the user's
- `npm install` where a lockfile exists: it may rewrite the lockfile, and `npm ci` is the form that
  installs exactly what is committed
- an interactive prompt: `refresh` runs these with nobody at the keyboard

A candidate that does one of these on a flag — `make reset` beside `make up` — is fine; only the
command being declared is judged.
