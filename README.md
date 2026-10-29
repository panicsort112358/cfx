# cfx

A command-line helper for Codeforces problems. The CLI lives in `cmd/cfx`,
with Codeforces API and scraping code, recommendation logic, problem reference parsing,
and workspace test running separated under `internal/`.

Use `cfx setup 71A` to create a workspace, `cfx test 71A` to run its sample
tests, or `cfx recommend tourist --tag dp` to find problems by rating and tag.
The setup shorthand `cfx 71A` remains available. Set the workspace directory
with `--dir PATH` anywhere in the command, and use `--no-open` to skip opening
the problem page during setup.

## Build

From the repository root, run:

```sh
go build ./cmd/cfx
```

This creates a `cfx` executable in the current directory. Run it with, for example:

```sh
./cfx help
```

By default, generated problem workspaces are stored under `workspace/` in the
current directory. The Codeforces problem cache is stored in the operating
system's user cache directory under `cfx/problems.json`.
