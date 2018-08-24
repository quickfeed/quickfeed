package models

type CIResult struct {
	Scores    []*ScoreObject `json:"scores"`
	BuildInfo *BuildInfo     `json:"buildinfo"`
}

type BuildInfo struct {
	BuildDate string `json:"builddate"`
	BuildID   int    `json:"buildid"`
	BuildLog  string `json:"buildlog"`
	ExecTime  int    `json:"execTime"`
}

type ScoreObject struct {
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Points int    `json:"points"`
	Weight int    `json:"weight"`
}

type CIOutput struct {
	Secret   string
	TestName string
	Score    int
	MaxScore int
	Weight   int
}

type AssignmentCIInfo struct {
	AccessToken        string
	CreatorAccessToken string
	GetURL             string
	TestURL            string
	AssignmentName     string
	RawGetURL          string
	RawTestURL         string
}
