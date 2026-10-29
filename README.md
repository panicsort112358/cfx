# cfx

A command-line helper for Codeforces problems. The CLI lives in `cmd/cfx`,
with Codeforces API and scraping code, recommendation logic, problem reference parsing,
and workspace test running separated under `internal/`.

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
