package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"cfx/internal/recommend"
	"cfx/internal/workspace"
)

var version = "devel"

const usage = `usage: cfx [options] COMMAND [ARG ...]

commands:
  setup <ID>                 create a problem workspace (also: cfx <ID>)
  test <ID>                  build and run the problem's sample tests
  recommend <HANDLE>         recommend problems for a Codeforces user
  help [COMMAND]             show general or command-specific help
  version                    print the program version

options:
  -dir, --dir <PATH>         workspace directory (default "workspace")
                             place it before or after the command

setup options:
  --no-open                  do not open the problem page in a browser

recommend options:
  --tag <TAG>                limit recommendations to a tag

Examples:
  cfx setup 71A
  cfx test 71A
  cfx recommend tourist --tag dp
  cfx test 71A --dir ~/codeforces
`

const (
	setupUsage     = "usage: cfx setup <ID> [--no-open]\n       cfx <ID> [--no-open]\n"
	testUsage      = "usage: cfx test <ID>\n"
	recommendUsage = "usage: cfx recommend <HANDLE> [--tag <TAG>]\n"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	workspaceDir, args, err := extractWorkspaceDir(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		if len(args) == 1 {
			fmt.Fprint(stdout, usage)
			return 0
		}
		return printCommandHelp(args[1], stdout, stderr)
	case "version", "-v", "--version":
		fmt.Fprintf(stdout, "cfx %s\n", version)
		return 0
	case "setup":
		if hasHelp(args[1:]) {
			fmt.Fprint(stdout, setupUsage)
			return 0
		}
		id, openPage, err := parseSetupArgs(args[1:])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, setupUsage)
			return 2
		}
		return setupProblemCommand(workspaceDir, id, openPage, stdout, stderr)
	case "test":
		if hasHelp(args[1:]) {
			fmt.Fprint(stdout, testUsage)
			return 0
		}
		id, err := parseSingleID(args[1:])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, testUsage)
			return 2
		}
		workspace.RunTests(workspaceDir, id)
		return 0
	case "recommend":
		if hasHelp(args[1:]) {
			fmt.Fprint(stdout, recommendUsage)
			return 0
		}
		handle, tag, err := parseRecommendArgs(args[1:])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, recommendUsage)
			return 2
		}
		recommend.RecommendProblems(handle, tag)
		return 0
	default:
		// Keep the bare problem ID as a shortcut for setup.
		if hasHelp(args[1:]) {
			fmt.Fprint(stdout, setupUsage)
			return 0
		}
		id, openPage, err := parseSetupArgs(args)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, setupUsage)
			return 2
		}
		return setupProblemCommand(workspaceDir, id, openPage, stdout, stderr)
	}
}

func extractWorkspaceDir(args []string) (string, []string, error) {
	workspaceDir := "workspace"
	commandArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-dir" || arg == "--dir":
			if i+1 == len(args) {
				return "", nil, fmt.Errorf("%s requires a path", arg)
			}
			i++
			workspaceDir = args[i]
			if workspaceDir == "" {
				return "", nil, fmt.Errorf("%s requires a path", arg)
			}
		case strings.HasPrefix(arg, "-dir="):
			workspaceDir = strings.TrimPrefix(arg, "-dir=")
			if workspaceDir == "" {
				return "", nil, fmt.Errorf("-dir requires a path")
			}
		case strings.HasPrefix(arg, "--dir="):
			workspaceDir = strings.TrimPrefix(arg, "--dir=")
			if workspaceDir == "" {
				return "", nil, fmt.Errorf("--dir requires a path")
			}
		default:
			commandArgs = append(commandArgs, arg)
		}
	}
	return workspaceDir, commandArgs, nil
}

func parseSetupArgs(args []string) (string, bool, error) {
	openPage := true
	id := ""
	for _, arg := range args {
		switch arg {
		case "--no-open":
			openPage = false
		default:
			if strings.HasPrefix(arg, "-") {
				return "", false, fmt.Errorf("unknown setup option %q", arg)
			}
			if id != "" {
				return "", false, fmt.Errorf("unexpected argument %q", arg)
			}
			id = arg
		}
	}
	if id == "" {
		return "", false, fmt.Errorf("missing problem ID")
	}
	return id, openPage, nil
}

func parseSingleID(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing problem ID")
	}
	if len(args) > 1 {
		return "", fmt.Errorf("unexpected argument %q", args[1])
	}
	if strings.HasPrefix(args[0], "-") {
		return "", fmt.Errorf("unknown test option %q", args[0])
	}
	return args[0], nil
}

func parseRecommendArgs(args []string) (string, string, error) {
	var tag string
	tagSet := false
	positionals := make([]string, 0, 2)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--tag":
			if tagSet {
				return "", "", fmt.Errorf("tag provided more than once")
			}
			if i+1 == len(args) {
				return "", "", fmt.Errorf("--tag requires a value")
			}
			i++
			tag = args[i]
			if tag == "" {
				return "", "", fmt.Errorf("--tag requires a value")
			}
			tagSet = true
		case strings.HasPrefix(arg, "--tag="):
			if tagSet {
				return "", "", fmt.Errorf("tag provided more than once")
			}
			tag = strings.TrimPrefix(arg, "--tag=")
			if tag == "" {
				return "", "", fmt.Errorf("--tag requires a value")
			}
			tagSet = true
		default:
			if strings.HasPrefix(arg, "-") {
				return "", "", fmt.Errorf("unknown recommend option %q", arg)
			}
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) == 0 {
		return "", "", fmt.Errorf("missing Codeforces handle")
	}
	if len(positionals) > 2 {
		return "", "", fmt.Errorf("unexpected argument %q", positionals[2])
	}
	handle := positionals[0]
	if len(positionals) == 2 {
		if tagSet {
			return "", "", fmt.Errorf("tag provided more than once")
		}
		tag = positionals[1] // Keep the old positional tag syntax working.
	}
	return handle, tag, nil
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func printCommandHelp(command string, stdout, stderr io.Writer) int {
	switch command {
	case "setup":
		fmt.Fprint(stdout, setupUsage)
	case "test":
		fmt.Fprint(stdout, testUsage)
	case "recommend":
		fmt.Fprint(stdout, recommendUsage)
	case "help", "version":
		fmt.Fprint(stdout, usage)
	default:
		fmt.Fprintf(stderr, "error: unknown command %q\n", command)
		return 2
	}
	return 0
}
