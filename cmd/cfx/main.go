package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"cfx/internal/recommend"
	"cfx/internal/workspace"
)

var version = "devel"

const usage = `usage: cfx [options] COMMAND [ARG ...]

commands:
  recommend   recommend problems based on your rating and solved history
  test        build and run C++ tests for a problem
  help        print this help
  version     print the program version

options:
  -dir string   base directory for problem workspaces (default "workspace")

examples:
  cfx 71A             setup workspace & open browser for problem 71A
  cfx test 71A        build and run tests for 71A
  cfx recommend tourist
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("cfx", flag.ContinueOnError)
	fs.SetOutput(stderr)
	workspaceFlag := fs.String("dir", "workspace", "base directory for problem workspaces")

	fs.Usage = func() {
		fmt.Fprint(stdout, usage)
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	switch subArgs[0] {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	case "version", "-v", "--version":
		fmt.Fprintf(stdout, "cfx %s\n", version)
		return 0
	case "test":
		if len(subArgs) < 2 {
			fmt.Fprintln(stderr, "error: missing problem ID for test command.")
			return 2
		}
		workspace.RunTests(*workspaceFlag, subArgs[1])
		return 0
	case "recommend":
		if len(subArgs) < 2 {
			fmt.Fprintln(stderr, "error: missing Codeforces handle. usage: cfx recommend <HANDLE> [tag]")
			return 2
		}
		handle := subArgs[1]
		tag := ""
		if len(subArgs) >= 3 {
			tag = subArgs[2]
		}
		recommend.RecommendProblems(handle, tag)
		return 0
	default:
		// Default setup workflow for a problem ID (e.g. "71A")
		return setupProblemCommand(*workspaceFlag, subArgs[0], stdout, stderr)
	}
}
