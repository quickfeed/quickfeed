package score

import (
	"strings"
	"time"
)

const (
	layout = "2006-01-02T15:04:05"
)

// TODO(meling) make most methods herein private; only ExtractResults is really needed, I think?

type results struct {
	testNames []string
	scores    map[string]*Score
}

func NewResults() *results {
	return &results{
		testNames: make([]string, 0),
		scores:    make(map[string]*Score),
	}
}

// ToScoreSlice returns a slice of score objects for the proto file.
func (r *results) ToScoreSlice() []*Score {
	scores := make([]*Score, len(r.testNames))
	for i, name := range r.testNames {
		scores[i] = r.scores[name]
	}
	return scores
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
	return &Results{
		BuildInfo: &BuildInfo{
			BuildDate: time.Now().Format(layout),
			BuildLog:  strings.Join(filteredLog, "\n"),
			ExecTime:  execTime.Milliseconds(),
		},
		Scores: results.ToScoreSlice(),
	}, nil
}

// AddScore adds the given score to the set of scores.
// This method assumes that the provided score object is valid.
func (r *results) AddScore(sc *Score) {
	testName := sc.GetTestName()
	if current, found := r.scores[testName]; found {
		if current.GetScore() != 0 {
			// We reach here only if a second non-zero score is found
			// Mark it as faulty with -1.
			sc.Score = -1
		}
	} else {
		// New test: record in TestNames
		r.testNames = append(r.testNames, testName)
	}

	// Record score object if:
	// - current score is nil or zero, or
	// - the first score was zero.
	r.scores[testName] = sc
}

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func (r *Results) Validate(secret string) error {
	for _, sc := range r.GetScores() {
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
	for _, ts := range r.GetScores() {
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
