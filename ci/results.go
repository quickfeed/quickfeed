package ci

import (
	"time"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
	"go.uber.org/zap"
)

// TODO(meling) Delete file after migrating tests to kit/score

// ExtractResult returns a result struct for the given log.
func ExtractResult(logger *zap.SugaredLogger, out, secret string, execTime time.Duration) (*score.Results, error) {
	results, err := score.ExtractResults(out, secret, execTime)
	if err != nil {
		return nil, err
	}
	logger.Debug("ci.ExtractResults",
		zap.Any("results", log.IndentJson(results)),
	)
	return results, nil
}
