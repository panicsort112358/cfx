package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type UserInfoResponse struct {
	Status string `json:"status"`
	Result []struct {
		Rating int `json:"rating"`
	} `json:"result"`
}

type SubmissionResponse struct {
	Status string `json:"status"`
	Result []struct {
		Problem struct {
			ContestID int    `json:"contestId"`
			Index     string `json:"index"`
		} `json:"problem"`
		Verdict string `json:"verdict"`
	} `json:"result"`
}

func recommendProblems(handle string, tagFilter string) {
	fmt.Printf("fetching user info for %s...\n", handle)

	// 1. Fetch user rating from Codeforces API
	infoURL := fmt.Sprintf("https://codeforces.com/api/user.info?handles=%s", handle)
	resp, err := http.Get(infoURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to Codeforces API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var userInfo UserInfoResponse
	if err := json.Unmarshal(bodyBytes, &userInfo); err != nil || userInfo.Status != "OK" || len(userInfo.Result) == 0 {
		fmt.Fprintf(os.Stderr, "error: could not fetch user info for handle '%s'\n", handle)
		os.Exit(1)
	}

	userRating := userInfo.Result[0].Rating
	if userRating == 0 {
		userRating = 800 // Default baseline for beginners
		fmt.Println("notice: user is unrated, defaulting rating baseline to 800.")
	} else {
		fmt.Printf("current rating: %d\n", userRating)
	}

	// 2. Fetch submission history to track solved problems
	fmt.Println("fetching submission history...")
	statusURL := fmt.Sprintf("https://codeforces.com/api/user.status?handle=%s", handle)
	respStatus, err := http.Get(statusURL)
	solvedMap := make(map[string]bool)
	if err == nil {
		defer respStatus.Body.Close()
		statusBytes, _ := io.ReadAll(respStatus.Body)
		var userStatus SubmissionResponse
		if json.Unmarshal(statusBytes, &userStatus) == nil && userStatus.Status == "OK" {
			for _, sub := range userStatus.Result {
				if sub.Verdict == "OK" {
					key := fmt.Sprintf("%d-%s", sub.Problem.ContestID, strings.ToUpper(sub.Problem.Index))
					solvedMap[key] = true
				}
			}
		}
	}

	// 3. Load problems from cache/API
	problems := loadOrFetchProblems()

	// Codeforces minimum problem rating is 800
	targetMin := userRating
	if targetMin < 800 {
		targetMin = 800
	}
	targetMax := targetMin + 200

	fmt.Printf("\nrecommending problems (rating range %d - %d):\n\n", targetMin, targetMax)

	count := 0
	for _, p := range problems {
		// Filter criteria: must have a rating, be within target range, and not already solved
		if p.Rating >= targetMin && p.Rating <= targetMax {
			key := fmt.Sprintf("%d-%s", p.ContestID, strings.ToUpper(p.Index))
			if solvedMap[key] {
				continue
			}

			// Optional tag filtering
			if tagFilter != "" {
				matchTag := false
				for _, t := range p.Tags {
					if strings.EqualFold(t, tagFilter) {
						matchTag = true
						break
					}
				}
				if !matchTag {
					continue
				}
			}

			fmt.Printf("[%d%s] %s (rating: %d) [%s]\n", p.ContestID, p.Index, p.Name, p.Rating, strings.Join(p.Tags, ", "))
			count++
			if count >= 10 {
				break
			}
		}
	}

	if count == 0 {
		fmt.Println("no recommendations found matching criteria.")
	}
}
