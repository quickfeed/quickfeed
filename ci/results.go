package ci

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/autograde/kit/score"
)

// Result holds scores and build information for one test execution
// for an assignment.
type Result struct {
	Scores    []*score.Score `json:"scores"`
	BuildInfo *BuildInfo     `json:"buildinfo"`
}

// BuildInfo holds build data for one test execution for an assignment.
type BuildInfo struct {
	BuildID   int64  `json:"buildid"`
	BuildDate string `json:"builddate"`
	BuildLog  string `json:"buildlog"`
	ExecTime  int64  `json:"execTime"`
}

// ExtractResult returns a result struct for the given log.
func ExtractResult(out, secret string, execTime time.Duration) (*Result, error) {
	var filteredLog []string
	var scores []*score.Score
	for _, line := range strings.Split(out, "\n") {
		// check if line has expected JSON score string
		if score.HasPrefix(line) {
			sc, err := score.Parse(line, secret)
			if err != nil {
				//TODO(meling) we should probably log parse errors?
				continue
			}
			scores = append(scores, sc)
		} else {
			// the filtered log without JSON score strings
			filteredLog = append(filteredLog, line)
		}
	}

	return &Result{
		Scores: scores,
		BuildInfo: &BuildInfo{
			BuildID:   1, //TODO(meling) this should be changed
			BuildDate: time.Now().Format("2006-01-02"),
			BuildLog:  strings.Join(filteredLog, "\n"),
			ExecTime:  int64(execTime),
		},
	}, nil
}

// Marshal returns marshalled information from the result struct.
func (r Result) Marshal() (buildInfo string, scores string, err error) {
	bi, e := json.Marshal(r.BuildInfo)
	if e == nil {
		scs, e := json.Marshal(r.Scores)
		if e == nil {
			buildInfo = string(bi)
			scores = string(scs)
		}
	}
	err = e
	return
}

// TotalScore returns the total score for this execution result.
func (r Result) TotalScore() uint8 {
	return score.Total(r.Scores)
}
