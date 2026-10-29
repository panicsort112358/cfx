package workspace

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cfx/internal/problem"
)

func RunTests(workspaceDir, rawID string) {
	contestID, index, err := problem.ParseID(rawID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	problemDir := filepath.Join(workspaceDir, contestID, index)
	solutionPath := filepath.Join(problemDir, "solution.cpp")
	if _, err := os.Stat(solutionPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Solution file not found at %s. Run setup first.\n", solutionPath)
		os.Exit(1)
	}

	matches, err := filepath.Glob(filepath.Join(problemDir, "in*.txt"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error finding test cases: %v\n", err)
		os.Exit(1)
	}

	var testNumbers []int
	for _, path := range matches {
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "in"), ".txt")
		number, err := strconv.Atoi(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: ignoring unexpected test input file %s\n", path)
			continue
		}
		testNumbers = append(testNumbers, number)
	}
	sort.Ints(testNumbers)
	if len(testNumbers) == 0 {
		fmt.Fprintln(os.Stderr, "error: no test cases found; run setup to fetch samples or add inN.txt/outN.txt files.")
		os.Exit(1)
	}

	binPath := filepath.Join(problemDir, "solution")
	if _, err := os.Stat(binPath); err == nil {
		_ = os.Remove(binPath)
	}

	fmt.Println("compiling solution.cpp...")
	compileCmd := exec.Command("g++", "-O3", "-std=c++17", solutionPath, "-o", binPath)
	var compileStderr bytes.Buffer
	compileCmd.Stderr = &compileStderr
	if err := compileCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "compilation failed:\n%s\n", compileStderr.String())
		os.Exit(1)
	}

	fmt.Printf("running %d test case(s)...\n\n", len(testNumbers))

	passedCount := 0
	for _, i := range testNumbers {
		inFile := filepath.Join(problemDir, fmt.Sprintf("in%d.txt", i))
		outFile := filepath.Join(problemDir, fmt.Sprintf("out%d.txt", i))

		if _, err := os.Stat(inFile); os.IsNotExist(err) {
			continue
		}

		expectedOutputBytes, err := os.ReadFile(outFile)
		if err != nil {
			fmt.Printf("test %d ... error reading expected output: %v\n", i, err)
			continue
		}
		expectedOutput := strings.TrimSpace(string(expectedOutputBytes))

		inputBytes, err := os.ReadFile(inFile)
		if err != nil {
			fmt.Printf("test %d ... error reading input: %v\n", i, err)
			continue
		}

		cmd := exec.Command(binPath)
		cmd.Stdin = bytes.NewReader(inputBytes)
		var outBytes bytes.Buffer
		var errBytes bytes.Buffer
		cmd.Stdout = &outBytes
		cmd.Stderr = &errBytes

		start := time.Now()
		runErr := cmd.Run()
		duration := time.Since(start)

		if runErr != nil {
			fmt.Printf("test %d ... FAILED (runtime error: %v) [%.2fms]\n", i, runErr, float64(duration.Microseconds())/1000.0)
			if errBytes.Len() > 0 {
				fmt.Printf("stderr:\n%s\n", errBytes.String())
			}
			continue
		}

		actualOutput := strings.TrimSpace(outBytes.String())

		if actualOutput == expectedOutput {
			fmt.Printf("test %d ... ok [%.2fms]\n", i, float64(duration.Microseconds())/1000.0)
			passedCount++
		} else {
			fmt.Printf("test %d ... FAILED [%.2fms]\n", i, float64(duration.Microseconds())/1000.0)
			fmt.Println("--- input ---")
			fmt.Println(strings.TrimSpace(string(inputBytes)))
			printOutputDiff(expectedOutput, actualOutput)
		}
	}

	fmt.Printf("\nresult: %d/%d tests passed.\n", passedCount, len(testNumbers))
	if passedCount < len(testNumbers) {
		os.Exit(1)
	}
}

func printOutputDiff(expected, actual string) {
	expectedLines := outputLines(expected)
	actualLines := outputLines(actual)
	fmt.Println("--- expected")
	fmt.Println("+++ actual")

	if len(expectedLines) > 100_000/max(1, len(actualLines)) {
		limit := max(len(expectedLines), len(actualLines))
		for i := 0; i < limit; i++ {
			if i < len(expectedLines) && i < len(actualLines) && expectedLines[i] == actualLines[i] {
				continue
			}
			if i < len(expectedLines) {
				fmt.Printf("-%s\n", expectedLines[i])
			}
			if i < len(actualLines) {
				fmt.Printf("+%s\n", actualLines[i])
			}
		}
		return
	}

	lcs := make([][]int, len(expectedLines)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(actualLines)+1)
	}
	for i := len(expectedLines) - 1; i >= 0; i-- {
		for j := len(actualLines) - 1; j >= 0; j-- {
			if expectedLines[i] == actualLines[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}

	for i, j := 0, 0; i < len(expectedLines) || j < len(actualLines); {
		switch {
		case i < len(expectedLines) && j < len(actualLines) && expectedLines[i] == actualLines[j]:
			fmt.Printf(" %s\n", expectedLines[i])
			i++
			j++
		case i < len(expectedLines) && (j == len(actualLines) || lcs[i+1][j] >= lcs[i][j+1]):
			fmt.Printf("-%s\n", expectedLines[i])
			i++
		default:
			fmt.Printf("+%s\n", actualLines[j])
			j++
		}
	}
}

func outputLines(output string) []string {
	if output == "" {
		return nil
	}
	return strings.Split(output, "\n")
}
