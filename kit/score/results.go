package score

import (
	"strings"
	"sync/atomic"
	"time"
)

const (
	layout = "2006-01-02T15:04:05"
)

var globalBuildID = new(int64)

// TODO(meling) make most methods herein private; only ExtractResults is really needed, I think?

func NewResults() *Results {
	return &Results{
		TestNames: make([]string, 0),
		ScoreMap:  make(map[string]*Score),
	}
}

// ExtractResults returns the results from a test execution extracted from the given out string.
func ExtractResults(out, secret string, execTime time.Duration) (*Results, error) {
	var filteredLog []string
	results := NewResults()
	for _, line := range strings.Split(out, "\n") {
		// check if line has expected JSON score string
		if HasPrefix(line) {
			sc, err := Parse(line, secret)
			if err != nil {
				return nil, err
			}
			results.AddScore(sc)
		} else if line != "" { // include only non-empty lines
			// the filtered log without JSON score strings
			filteredLog = append(filteredLog, line)
		}
	}
	results.BuildInfo = &BuildInfo{
		BuildID:   atomic.AddInt64(globalBuildID, 1),
		BuildDate: time.Now().Format(layout),
		BuildLog:  strings.Join(filteredLog, "\n"),
		ExecTime:  execTime.Milliseconds(),
	}
	return results, nil
}

// AddScore adds the given score to the set of scores.
// This method assumes that the provided score object is valid.
func (r *Results) AddScore(sc *Score) {
	testName := sc.GetTestName()
	if current, found := r.ScoreMap[testName]; found {
		if current.GetScore() != 0 {
			// We reach here only if a second non-zero score is found
			// Mark it as faulty with -1.
			sc.Score = -1
		}
	} else {
		// New test: record in TestNames
		r.TestNames = append(r.TestNames, testName)
	}

	// Record score object if:
	// - current score is nil or zero, or
	// - the first score was zero.
	r.ScoreMap[testName] = sc
}

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func (r *Results) Validate(secret string) error {
	for _, sc := range r.GetScoreMap() {
		if err := sc.IsValid(secret); err != nil {
			return err
		}
	}
	return nil
}

// Sum returns the total score computed over the set of recorded scores.
// The total is a grade in the range 0-100.
// This method must only be called after Validate has returned nil.
func (r *Results) Sum() uint32 {
	totalWeight := float32(0)
	var max, score, weight []float32
	for _, ts := range r.GetScoreMap() {
		totalWeight += float32(ts.Weight)
		weight = append(weight, float32(ts.Weight))
		score = append(score, float32(ts.Score))
		max = append(max, float32(ts.MaxScore))
	}
	total := float32(0)
	for i := 0; i < len(score); i++ {
		if score[i] > max[i] {
			score[i] = max[i]
		}
		total += ((score[i] / max[i]) * (weight[i] / totalWeight))
	}
	return uint32(total * 100)
}
