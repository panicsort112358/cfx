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

const usage = `usage: cfx <ID>
       cfx setup <ID>
       cfx test <ID>
       cfx submit [ID]
       cfx recommend <HANDLE>
       cfx help [COMMAND]
       cfx version
`

const (
	setupUsage = `usage: cfx <ID>
       cfx setup <ID>
`
	testUsage = `usage: cfx test <ID>
`
	submitUsage = `usage: cfx submit [ID]
`
	recommendUsage = `usage: cfx recommend <HANDLE>
`
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	switch args[0] {
	case "help":
		if len(args) > 2 {
			fmt.Fprintf(stderr, "error: unexpected argument %q\n", args[2])
			return 2
		}
		if len(args) == 1 {
			fmt.Fprint(stdout, usage)
			return 0
		}
		return printCommandHelp(args[1], stdout, stderr)
	case "version":
		if len(args) > 1 {
			fmt.Fprintf(stderr, "error: unexpected argument %q\n", args[1])
			return 2
		}
		fmt.Fprintf(stdout, "cfx %s\n", version)
		return 0
	case "setup":
		id, err := parseSetupArgs(args[1:])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, setupUsage)
			return 2
		}
		workspaceDir, err := loadWorkspaceDir()
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
		return setupProblemCommand(workspaceDir, id, stdout, stderr)
	case "test":
		id, err := parseSingleID(args[1:])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, testUsage)
			return 2
		}
		workspaceDir, err := loadWorkspaceDir()
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
		workspace.RunTests(workspaceDir, id)
		return 0
	case "submit":
		if len(args) == 1 {
			return submitCurrentCommand(stdout, stderr)
		}
		id, err := parseSingleID(args[1:])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, submitUsage)
			return 2
		}
		workspaceDir, err := loadWorkspaceDir()
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
		return submitProblemCommand(workspaceDir, id, stdout, stderr)
	case "recommend":
		handle, err := parseRecommendArgs(args[1:])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, recommendUsage)
			return 2
		}
		recommend.RecommendProblems(handle)
		return 0
	default:
		// Keep the bare problem ID as a shortcut for setup.
		id, err := parseSetupArgs(args)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n%s", err, setupUsage)
			return 2
		}
		workspaceDir, err := loadWorkspaceDir()
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
		return setupProblemCommand(workspaceDir, id, stdout, stderr)
	}
}

func parseSetupArgs(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing problem ID")
	}
	if len(args) > 1 {
		return "", fmt.Errorf("unexpected argument %q", args[1])
	}
	if strings.HasPrefix(args[0], "-") {
		return "", fmt.Errorf("invalid problem ID %q", args[0])
	}
	return args[0], nil
}

func parseSingleID(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing problem ID")
	}
	if len(args) > 1 {
		return "", fmt.Errorf("unexpected argument %q", args[1])
	}
	if strings.HasPrefix(args[0], "-") {
		return "", fmt.Errorf("invalid problem ID %q", args[0])
	}
	return args[0], nil
}

func parseRecommendArgs(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing Codeforces handle")
	}
	if len(args) > 1 {
		return "", fmt.Errorf("unexpected argument %q", args[1])
	}
	if strings.HasPrefix(args[0], "-") {
		return "", fmt.Errorf("invalid Codeforces handle %q", args[0])
	}
	return args[0], nil
}

func printCommandHelp(command string, stdout, stderr io.Writer) int {
	switch command {
	case "setup":
		fmt.Fprint(stdout, setupUsage)
	case "test":
		fmt.Fprint(stdout, testUsage)
	case "submit":
		fmt.Fprint(stdout, submitUsage)
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
