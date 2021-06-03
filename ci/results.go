package ci

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
	"go.uber.org/zap"
)

var globalBuildID = new(int64)

// ExtractResult returns a result struct for the given log.
func ExtractResult(logger *zap.SugaredLogger, out, secret string, execTime time.Duration) (*score.Result, error) {
	var filteredLog []string
	scores := score.NewScores()
	for _, line := range strings.Split(out, "\n") {
		// check if line has expected JSON score string
		if score.HasPrefix(line) {
			// Parse returns score object with hidden secret
			sc, err := score.Parse(line, secret)
			if err != nil {
				logger.Error("ci.ExtractResults",
					zap.Error(err),
					zap.String("line", line),
				)
				continue
			}
			scores.AddScore(sc)
		} else if line != "" { // include only non-empty lines
			// the filtered log without JSON score strings
			filteredLog = append(filteredLog, line)
		}
	}

	// TODO(meling) Fix scores and Result types to use protobuf??
	// Currently, BuildInfo and Scores are stored as string in database
	// (and transmitted as JSON string to frontend).
	// This should be changed to protobuf as well.
	logger.Debug("ci.ExtractResults",
		zap.Any("scores", log.IndentJson(scores)),
		zap.Any("filteredLog", log.IndentJson(filteredLog)),
	)
	return &score.Result{
		Scores: scores.ToScoreSlice(),
		BuildInfo: &score.BuildInfo{
			BuildID:   atomic.AddInt64(globalBuildID, 1),
			BuildDate: time.Now().Format(layout),
			BuildLog:  strings.Join(filteredLog, "\n"),
			ExecTime:  execTime.Milliseconds(),
		},
	}, nil
}
