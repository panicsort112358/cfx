package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"cfx/internal/problem"
)

func submitProblemCommand(workspaceDir, rawID string, stdout, stderr io.Writer) int {
	contestID, index, err := problem.ParseID(rawID)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}

	solutionPath := filepath.Join(workspaceDir, contestID, index, "solution.cpp")
	return copySolutionFile(solutionPath, "", stdout, stderr)
}

func submitCurrentCommand(stdout, stderr io.Writer) int {
	return copySolutionFile("solution.cpp", "run from a problem directory or pass a problem ID", stdout, stderr)
}

func copySolutionFile(solutionPath, missingHint string, stdout, stderr io.Writer) int {
	solution, err := os.ReadFile(solutionPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(stderr, "error: solution not found at %s", solutionPath)
			if missingHint != "" {
				fmt.Fprintf(stderr, "; %s", missingHint)
			}
			fmt.Fprintln(stderr)
		} else {
			fmt.Fprintf(stderr, "error: could not read %s: %v\n", solutionPath, err)
		}
		return 1
	}

	if err := copyToClipboard(solution); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "copied solution.cpp to clipboard; paste it into the open browser.")
	return 0
}

func copyToClipboard(contents []byte) error {
	command, args, err := clipboardCommand()
	if err != nil {
		return err
	}

	cmd := exec.Command(command, args...)
	cmd.Stdin = bytes.NewReader(contents)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard command %s failed: %w", command, err)
	}
	return nil
}

func clipboardCommand() (string, []string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "pbcopy", nil, nil
	case "windows":
		return "clip", nil, nil
	case "linux":
		candidates := []struct {
			name string
			args []string
		}{
			{name: "xclip", args: []string{"-selection", "clipboard"}},
			{name: "xsel", args: []string{"--clipboard", "--input"}},
			{name: "wl-copy"},
		}
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			candidates[0], candidates[2] = candidates[2], candidates[0]
		}
		for _, candidate := range candidates {
			if path, err := exec.LookPath(candidate.name); err == nil {
				return path, candidate.args, nil
			}
		}
		return "", nil, fmt.Errorf("no clipboard command found; install wl-clipboard, xclip, or xsel")
	default:
		return "", nil, fmt.Errorf("clipboard is not supported on %s", runtime.GOOS)
	}
}
