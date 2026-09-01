// Package server simulates how the QuickFeed server uses the score package,
// that is, from a caller without a test function on the call stack.
//
// Score extraction runs on the server, where the call frame lookup in the
// kit/internal/test package panics if it cannot find a test function.
// Extraction must therefore never depend on that lookup; a student can emit
// arbitrary score lines, and an invalid one must not take down the server.
package server

import (
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
)

// ExtractResults calls [score.ExtractResults] from a non-test function.
func ExtractResults(out, secret string, zeroScoreTests []*score.Score) (*score.Results, error) {
	return score.ExtractResults(out, secret, time.Millisecond, zeroScoreTests)
}
