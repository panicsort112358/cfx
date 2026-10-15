package codeforces

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func LoadOrFetchProblems() []Problem {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to locate user cache directory: %v\n", err)
		os.Exit(1)
	}
	cacheDir = filepath.Join(cacheDir, "cfx")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cache directory: %v\n", err)
		os.Exit(1)
	}
	cacheFile := filepath.Join(cacheDir, "problems.json")
	var data ProblemSetResponse
	useCache := false

	if info, err := os.Stat(cacheFile); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			fileBytes, err := os.ReadFile(cacheFile)
			if err == nil {
				if err := json.Unmarshal(fileBytes, &data); err == nil {
					useCache = true
				}
			}
		}
	}

	// Inside api.go
	if !useCache {
		fmt.Println("Fetching problems from Codeforces API...")
		resp, err := http.Get("https://codeforces.com/api/problemset.problems")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to Codeforce: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			fmt.Fprint(os.Stderr, "Failed to parse JSON response: %v\n", err)
			os.Exit(1)
		}

		prettyJSON, _ := json.MarshalIndent(data, "", " ")
		_ = os.WriteFile(cacheFile, prettyJSON, 0644)
	}

	return data.Result.Problems
}
