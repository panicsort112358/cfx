package main

type Problem struct {
	ContestID int      `json:"contestId"`
	Index     string   `json:"index"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Points    float64  `json:"points"`
	Rating    int      `json:"rating"`
	Tags      []string `json:"tags"`
}

type ProblemStatistics struct {
	ContestID   int    `json:"contestId"`
	Index       string `json:"index"`
	SolvedCount int    `json:"solvedCount"`
}

type ProblemSetResponse struct {
	Status string `json:"status"`
	Result struct {
		Problems          []Problem           `json:"problems"`
		ProblemStatistics []ProblemStatistics `json:"problemStatistics"`
	} `json:"result"`
}

type SampleCase struct {
	Input  string
	Output string
}
