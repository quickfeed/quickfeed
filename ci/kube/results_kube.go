package kube

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	"github.com/autograde/aguis/kit/score"
	"go.uber.org/zap"
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

var globalBuildID = new(int64)

// ExtractKubeResult returns a result struct for the given log.
func ExtractKubeResult(logger *zap.SugaredLogger, out, secret string, execTime time.Duration) (*Result, error) {
	var filteredLog []string
	scores := make([]*score.Score, 0)
	for _, line := range strings.Split(out, "\n") {
		// check if line has expected JSON score string
		if score.HasPrefix(line) {
			sc, err := score.Parse(line, secret)
			if err != nil {
				logger.Error("ci.ExtractResults",
					zap.Error(err),
					zap.String("line", line),
				)
				continue
			}
			scores = append(scores, sc)
		} else {
			// the filtered log without JSON score strings
			filteredLog = append(filteredLog, line)
		}
	}
	logger.Debug("ci.ExtractResults",
		zap.Any("scores", scores),
		zap.Any("filteredLog", filteredLog),
	)
	return &Result{
		Scores: scores,
		BuildInfo: &BuildInfo{
			BuildID:   atomic.AddInt64(globalBuildID, 1),
			BuildDate: time.Now().Format("2006-01-02T15:04:05"),
			BuildLog:  strings.Join(filteredLog, "\n"),
			ExecTime:  execTime.Milliseconds(),
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
func (r Result) TotalScore() uint32 {
	return score.Total(r.Scores)
}
