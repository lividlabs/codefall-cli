# Testing

Cases and specs under this tree are written through `/codefall-implement`, run through
`/codefall-test`, and the harness they need is set up through `/codefall-equip`.

## Runners

One line per runner: its name, the command that runs every spec, and the command that runs one case.

## Setup and state-forcing commands

The commands that prepare a run, and the ones that put the product into a state a case needs.

## Real side effects

How a run records what it creates against the real services, and how it cleans that up afterwards.

## Environment notes

What a run needs from this machine or this account, and anything that is only true here.
