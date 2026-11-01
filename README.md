# cfx

A command-line helper for Codeforces problems. It creates local workspaces,
downloads sample tests, runs C++ solutions, and recommends problems for a user.

## Build

From the repository root:

```sh
go build -o cfx ./cmd/cfx
```

## Configuration

The workspace root is set in `.env`:

```sh
CFX_DIR=../ok/codeforces
```

Relative paths are resolved from the current directory. Run `cfx` from the
repository root.

## Usage

```sh
./cfx 71A
./cfx test 71A
./cfx submit 71A
./cfx recommend tourist
```

`./cfx <ID>` saves available samples, creates `solution.cpp` if needed, and
opens the problem page. `./cfx setup <ID>` is an equivalent form. Repeating
setup may replace sample files; an existing `solution.cpp` is kept.

`./cfx test <ID>` compiles `solution.cpp` with `g++` and runs saved cases.
Failed tests show an expected-versus-actual diff.

`./cfx submit <ID>` copies that workspace's `solution.cpp` to the clipboard.
Run `./cfx submit` from a problem directory to copy its local `solution.cpp`.
Paste the code into the open Codeforces page.

`./cfx recommend <HANDLE>` suggests unsolved problems and includes problem
links. Run `./cfx help` for command syntax.

The Codeforces problem cache is stored under `cfx/problems.json` in the
operating system's user cache directory.
