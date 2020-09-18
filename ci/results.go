package ci

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
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

// ExtractResult returns a result struct for the given log.
func ExtractResult(logger *zap.SugaredLogger, out, secret string, execTime time.Duration) (*Result, error) {
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
		} else if line != "" { // include only non-empty lines
			// the filtered log without JSON score strings
			filteredLog = append(filteredLog, line)
		}
	}
	scores = filter(scores)
	logger.Debug("ci.ExtractResults",
		zap.Any("scores", log.IndentJson(scores)),
		zap.Any("filteredLog", log.IndentJson(filteredLog)),
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

// filter returns a slice of scores, exactly one per TestName.
// The input score slice may contain one or two entries per TestName.
// If more than two entries are found for a given TestName, we return a 0 score for that test.
//
// If there is only one entry for a test it is likely that the test panicked,
// in which case we want to report a 0 score for that test.
func filter(scores []*score.Score) []*score.Score {
	// map: test name -> scores for that test (one or more)
	tests := make(map[string][]*score.Score, len(scores))
	// map: test name -> test number (to keep same test order as in the input)
	testOrder := make(map[string]int, len(scores))
	numTests := 0
	for _, score := range scores {
		tests[score.TestName] = append(tests[score.TestName], score)
		if _, found := testOrder[score.TestName]; !found {
			testOrder[score.TestName] = numTests
			numTests++
		}
	}

	newScores := make([]*score.Score, len(testOrder))
	for name, pos := range testOrder {
		scoresForTest := tests[name]
		// get the score for the test
		// (the last element should hold the actual score)
		newScores[pos] = scoresForTest[len(scoresForTest)-1]
		if len(scoresForTest) > 2 {
			// if more than two scores were found, always return 0 score
			// This is probably be a bug in the teacher's test code.
			newScores[pos].Score = 0
		}
	}
	return newScores
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
