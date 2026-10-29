package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"cfx/internal/codeforces"
	"cfx/internal/problem"
)

func setupProblemCommand(workspaceDir, rawID string, openPage bool, stdout, stderr io.Writer) int {
	contestID, index, err := problem.ParseID(rawID)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	problemURL := fmt.Sprintf("https://codeforces.com/contest/%s/problem/%s", contestID, index)
	problemName := findProblemName(contestID, index, codeforces.LoadOrFetchProblems())
	fmt.Fprintf(stdout, "setting up workspace for [%s%s] %s\n", contestID, index, problemName)
	fmt.Fprintf(stdout, "problem page: %s\n", problemURL)

	problemDir := filepath.Join(workspaceDir, contestID, index)
	if err := os.MkdirAll(problemDir, 0755); err != nil {
		fmt.Fprintf(stderr, "error: failed to create directory: %v\n", err)
		return 1
	}

	if hasSampleFiles(problemDir) {
		fmt.Fprintln(stdout, "existing sample files may be overwritten by refreshed samples; solution.cpp will be kept.")
	}
	fmt.Fprintln(stdout, "scraping sample test cases...")
	samples, err := codeforces.ScrapeSampleCases(problemURL)
	if err != nil {
		fmt.Fprintf(stderr, "warning: failed to scrape sample test cases: %v\n", err)
	} else if len(samples) == 0 {
		fmt.Fprintln(stdout, "no sample cases found; existing sample files were left unchanged.")
	} else {
		fmt.Fprintf(stdout, "found %d sample test case(s).\n", len(samples))
		for i, sample := range samples {
			files := []struct{ name, contents string }{
				{fmt.Sprintf("in%d.txt", i+1), sample.Input},
				{fmt.Sprintf("out%d.txt", i+1), sample.Output},
			}
			for _, file := range files {
				if err := os.WriteFile(filepath.Join(problemDir, file.name), []byte(file.contents), 0644); err != nil {
					fmt.Fprintf(stderr, "error: failed to save %s: %v\n", file.name, err)
					return 1
				}
			}
		}
		fmt.Fprintln(stdout, "saved sample test cases to disk.")
	}

	solutionPath := filepath.Join(problemDir, "solution.cpp")
	if _, err := os.Stat(solutionPath); os.IsNotExist(err) {
		if err := os.WriteFile(solutionPath, []byte(solutionTemplate(problemName, contestID, index, problemURL)), 0644); err != nil {
			fmt.Fprintf(stderr, "error: failed to create solution file: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "created solution file: %s\n", solutionPath)
	} else if err != nil {
		fmt.Fprintf(stderr, "error: failed to check solution file: %v\n", err)
		return 1
	} else {
		fmt.Fprintf(stdout, "solution file already exists at: %s\n", solutionPath)
	}

	if openPage {
		fmt.Fprintln(stdout, "opening problem page in browser...")
		if err := openBrowser(problemURL); err != nil {
			fmt.Fprintf(stderr, "warning: could not open browser: %v\n", err)
		}
	}
	return 0
}

func hasSampleFiles(problemDir string) bool {
	for _, pattern := range []string{"in*.txt", "out*.txt"} {
		matches, _ := filepath.Glob(filepath.Join(problemDir, pattern))
		if len(matches) > 0 {
			return true
		}
	}
	return false
}

func findProblemName(contestID, index string, problems []codeforces.Problem) string {
	contestNumber, _ := strconv.Atoi(contestID)
	for _, p := range problems {
		if p.ContestID == contestNumber && strings.EqualFold(p.Index, index) {
			return p.Name
		}
	}
	return "Unknown Problem"
}

func solutionTemplate(name, contestID, index, url string) string {
	return fmt.Sprintf(`#include <iostream>

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
`, name, contestID, index, url)
}

func openBrowser(url string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		command, args = "open", []string{url}
	default:
		command, args = "xdg-open", []string{url}
	}
	return exec.Command(command, args...).Start()
}
