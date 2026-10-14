package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func runTests(workspaceFlag, rawProblemID string) {
	contestID, index, err := parseProblemID(rawProblemID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	problemDir := filepath.Join(workspaceFlag, contestID, index)
	solutionPath := filepath.Join(problemDir, "solution.cpp")
	if _, err := os.Stat(solutionPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Solution file not found at %s. Run setup first.\n", solutionPath)
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

	matches, err := filepath.Glob(filepath.Join(problemDir, "in*.txt"))
	if err != nil || len(matches) == 0 {
		fmt.Println("no test cases found in workspace.")
		return
	}

	fmt.Printf("running %d test case(s)...\n\n", len(matches))

	passedCount := 0
	for i := 1; i <= len(matches); i++ {
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
			fmt.Println("--- expected ---")
			fmt.Println(expectedOutput)
			fmt.Println("--- got ---")
			fmt.Println(actualOutput)
			fmt.Println("--------------")
		}
	}

	fmt.Printf("\nresult: %d/%d tests passed.\n", passedCount, len(matches))
	if passedCount < len(matches) {
		os.Exit(1)
	}
}
