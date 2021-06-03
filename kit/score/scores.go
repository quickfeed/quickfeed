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

func newResults() *Results {
	return &Results{
		TestNames: make([]string, 0),
		ScoreMap:  make(map[string]*Score),
	}
}

// func (r *Results) 
// ExtractResults returns the results from a test execution extracted from the given out string.
func ExtractResults(out, secret string, execTime time.Duration) (*Results, error) {
	var filteredLog []string
	results := newResults()
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

// TODO(meling) Remove Scores message type from proto and methods below
// TODO(meling) need to update tests to use newResults() and ExtractResults instead.

func NewScores() *Scores {
	return &Scores{
		TestNames: make([]string, 0),
		ScoreMap:  make(map[string]*Score),
	}
}

// ToScoreSlice returns a slice of score objects for use with the current frontend and database.
// This is experimental API and should not be used. It may be removed in the future.
func (s *Scores) ToScoreSlice() []*Score {
	scores := make([]*Score, 0)
	for _, name := range s.TestNames {
		scores = append(scores, s.ScoreMap[name])
	}
	return scores
}

// AddScore adds the given score to the set of scores.
// This method assumes that the provided score object is valid.
func (s *Scores) AddScore(sc *Score) {
	testName := sc.GetTestName()
	if current, found := s.ScoreMap[testName]; found {
		if current.GetScore() != 0 {
			// We reach here only if a second non-zero score is found
			// Mark it as faulty with -1.
			sc.Score = -1
		}
	} else {
		// New test: record in TestNames
		s.TestNames = append(s.TestNames, testName)
	}

	// Record score object if:
	// - current score is nil or zero, or
	// - the first score was zero.
	s.ScoreMap[testName] = sc
}

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func (s *Scores) Validate(secret string) error {
	for _, sc := range s.GetScoreMap() {
		if err := sc.IsValid(secret); err != nil {
			return err
		}
	}
	return nil
}

// Sum returns the total score computed over the set of recorded scores.
// The total is a grade in the range 0-100.
// This method must only be called after Validate has returned nil.
func (s *Scores) Sum() uint32 {
	totalWeight := float32(0)
	var max, score, weight []float32
	for _, ts := range s.GetScoreMap() {
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
