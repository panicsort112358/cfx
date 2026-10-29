# cfx

A command-line helper for Codeforces problems. It creates local workspaces,
downloads sample tests, runs C++ solutions, and recommends problems for a user.

## Build

From the repository root:

```sh
go build -o cfx ./cmd/cfx
```

## Usage

```sh
cfx setup 71A
cfx test 71A
cfx recommend tourist --tag dp
```

`cfx 71A` is a shorthand for `cfx setup 71A`. Setup creates a workspace under
`workspace/<contest>/<index>`, saves available samples, creates `solution.cpp`
if needed, and opens the problem page. Pass `--no-open` to skip opening the
browser. Repeating setup may replace sample files; an existing `solution.cpp`
is kept.

Use `test` to compile `solution.cpp` with `g++` and run the saved sample cases.
Failed tests show an expected-versus-actual diff. A workspace without sample
tests is reported as an error.

`recommend` suggests up to ten unsolved problems around the user's rating.
Use `--tag` to filter by a Codeforces tag. Recommendations include problem
links.

## Options and help

Use `-dir` or `--dir` to choose a workspace directory. The option can appear
before or after the command:

```sh
cfx --dir ~/codeforces setup 71A --no-open
cfx test 71A --dir ~/codeforces
```

Run `cfx help` for the command list or `cfx help setup`, `cfx help test`, or
`cfx help recommend` for command-specific usage. The Codeforces problem cache
is stored under `cfx/problems.json` in the operating system's user cache
directory.
