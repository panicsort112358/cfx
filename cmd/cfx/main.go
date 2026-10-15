package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"cfx/internal/codeforces"
	"cfx/internal/problemid"
	"cfx/internal/recommend"
	problemworkspace "cfx/internal/workspace"
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
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
		fmt.Fprint(stdout, "cfx %s\n", version)
		return 0
	case "test":
		if len(subArgs) < 2 {
			fmt.Fprintln(stderr, "error: missing problem ID for test command.")
			return 2
		}
		return runTestsCommand(*workspaceFlag, subArgs[1], stdout, stderr)
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
		return recommendProblemsCommand(handle, tag, stdout, stderr)
	default:
		// Default setup workflow for a problem ID (e.g. "71A")
		return setupProblemCommand(*workspaceFlag, subArgs[0], stdout, stderr)
	}
}

func setupProblemCommand(workspaceDir, rawProblemID string, stdout, stderr io.Writer) int {
	contestID, index, err := problemid.Parse(rawProblemID)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	problemURL := fmt.Sprintf("https://codeforces.com/contest/%s/problem/%s", contestID, index)

	problems := codeforces.LoadOrFetchProblems()
	problemName := "Unknown Problem"
	var contestIDInt int
	fmt.Sscanf(contestID, "%d", &contestIDInt)

	for _, p := range problems {
		if p.ContestID == contestIDInt && strings.EqualFold(p.Index, index) {
			problemName = p.Name
			break
		}
	}

	fmt.Fprintf(stdout, "setting up workspace for [%s%s] %s\n", contestID, index, problemName)

	problemDir := filepath.Join(workspaceDir, contestID, index)
	if err := os.MkdirAll(problemDir, 0755); err != nil {
		fmt.Fprintf(stderr, "error: failed to created directory: %v\n", err)
		return 1
	}

	fmt.Println(stdout, "scraping sample test cases...")
	samples, err := codeforces.ScrapeSampleCases(problemURL)
	if err != nil {
		fmt.Fprintf(stderr, "warning: failed to scrape sample test cases: %v\n", err)
	} else {
		fmt.Fprintf(stdout, "found %d sample test case(s).\n", len(samples))
		for i, sample := range samples {
			inFile := filepath.Join(problemDir, fmt.Sprintf("in%d.txt", i+1))
			outFile := filepath.Join(problemDir, fmt.Sprintf("out%d.txt", i+1))
			_ = os.WriteFile(inFile, []byte(sample.Input), 0644)
			_ = os.WriteFile(outFile, []byte(sample.Output), 0644)
		}
		fmt.Println(stdout, "saved sample test cases to disk.")
	}

	solutionPath := filepath.Join(problemDir, "solution.cpp")
	if _, err := os.Stat(solutionPath); os.IsNotExist(err) {
		boilerplate := fmt.Sprintf(`#include <iostream>

using namespace std;

/*
- Problem: %s [%s%s]
- URL: %s
 */

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    
    // Write your solution here
    
    return 0;
}
`, problemName, contestID, index, problemURL)
		_ = os.WriteFile(solutionPath, []byte(boilerplate), 0644)
		fmt.Fprintf(stdout, "created solution file: %s\n", solutionPath)
	} else {
		fmt.Fprintf(stdout, "solution file already exists at: %s\n", solutionPath)
	}

	fmt.Fprintf(stdout, "opening problem URL in browser...")
	if err := openBrowser(problemURL); err != nil {
		fmt.Fprintf(stderr, "warning: could not open browser: %v\n", err)
	}

	return 0
}

func runTestsCommand(workspaceFlag, problemID string, stdout, stderr io.Writer) int {
	problemworkspace.RunTests(workspaceFlag, problemID)
	return 0
}

func recommendProblemsCommand(handle, tag string, stdout, stderr io.Writer) int {
	recommend.RecommendProblems(handle, tag)
	return 0
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}
