#!/usr/bin/env bash
#
# The local environment for this project: `start` brings up what it needs running locally, and
# `update` makes the local environment match the checkout. codefall-refresh runs them in that
# order; anyone can run either from a terminal.
#
# Both are safe to run at any time and cheap when nothing changed. Neither drops, resets, or
# deletes anything. A step that fails says what to do about it on stderr.
#
# Usage: scripts/local.sh start | update

set -euo pipefail

cd "$(dirname "$0")/.."

start() {
  # <One line per service: what it brings up, and the wait that makes exit 0 mean "up".>
  # docker compose up -d --wait
  :
}

update() {
  # <One line per step, in this order: runtimes, the env file, dependencies, generated code,
  # migrations. The env file comes before anything that reads it: the migration step below takes
  # its database URL from there.>
  # mise install
  # [ -f <env file> ] || cp <sample file> <env file>
  #   <sample file> is what the project commits — .env.example, .env.sample — and <env file> is
  #   what its tooling loads: .env for most dotenv loaders, .env.local for Next.js.
  # lock=$(git hash-object package-lock.json)
  # [ "$lock" = "$(cat node_modules/.codefall-lock 2>/dev/null)" ] || {
  #   npm ci   # reinstalls from scratch every run, so it runs only when the lockfile changed
  #   echo "$lock" > node_modules/.codefall-lock
  # }
  # npx prisma generate
  # npx prisma migrate deploy
  :
}

case ${1:-} in
  start) start ;;
  update) update ;;
  *)
    echo "usage: $0 start | update" >&2
    exit 2
    ;;
esac
