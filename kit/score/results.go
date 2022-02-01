package score

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	layout = "2006-01-02T15:04:05"
)

func NewResults(scores ...*Score) *Results {
	r := &Results{
		testNames: make([]string, 0),
		scoreMap:  make(map[string]*Score),
	}
	for _, sc := range scores {
		r.addScore(sc)
	}
	r.Scores = r.toScoreSlice()
	return r
}

// toScoreSlice returns a slice of score objects for the proto file.
func (r *Results) toScoreSlice() []*Score {
	scores := make([]*Score, len(r.testNames))
	for i, name := range r.testNames {
		scores[i] = r.scoreMap[name]
	}
	return scores
}

// Results contains the score objects, build info, and errors.
type Results struct {
	BuildInfo *BuildInfo // build info for tests
	Scores    []*Score   // list of scores for different tests
	testNames []string   // defines the order
	scoreMap  map[string]*Score
}

// parseErrors encountered during test execution.
type parseErrors []error

// Error prints a newline separated list of errors that occurred during parsing.
func (pe parseErrors) Error() string {
	if len(pe) == 0 {
		return ""
	}
	sErr := make([]string, 0, len(pe)+1)
	sErr = append(sErr, fmt.Sprintf("failed to parse score; %d occurrences", len(pe)))
	for _, err := range pe {
		sErr = append(sErr, err.Error())
	}
	return strings.Join(sErr, "\n")
}

// ExtractResults returns the results from a test execution extracted from the given out string.
func ExtractResults(out, secret string, execTime time.Duration) (*Results, error) {
	var filteredLog []string
	errs := make(parseErrors, 0)
	results := NewResults()
	for _, line := range strings.Split(out, "\n") {
		// check if line has expected JSON score string
		if HasPrefix(line) {
			sc, err := Parse(line, secret)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed on line '%s': %v", line, err))
				continue
			}
			results.addScore(sc)
		} else if line != "" { // include only non-empty lines
			// the filtered log without JSON score strings
			filteredLog = append(filteredLog, line)
		}
	}
	res := &Results{
		BuildInfo: &BuildInfo{
			BuildDate: time.Now().Format(layout),
			BuildLog:  strings.Join(filteredLog, "\n"),
			ExecTime:  execTime.Milliseconds(),
		},
		Scores: results.toScoreSlice(),
	}
	if len(errs) > 0 {
		return res, errs
	}
	return res, nil
}

// addScore adds the given score to the set of scores.
// This method assumes that the provided score object is valid.
func (r *Results) addScore(sc *Score) {
	testName := sc.GetTestName()
	if current, found := r.scoreMap[testName]; found {
		if current.GetScore() != 0 {
			// We reach here only if a second non-zero score is found
			// Mark it as faulty with -1.
			sc.Score = -1
		}
	} else {
		// New test: record in r.testNames
		r.testNames = append(r.testNames, testName)
	}

	// Record score object if:
	// - current score is nil or zero, or
	// - the first score was zero.
	r.scoreMap[testName] = sc
}

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func (r *Results) Validate(secret string) error {
	for _, sc := range r.Scores {
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
	totalWeight := float64(0)
	var max, score, weight []float64
	for _, ts := range r.Scores {
		totalWeight += float64(ts.Weight)
		weight = append(weight, float64(ts.Weight))
		score = append(score, float64(ts.Score))
		max = append(max, float64(ts.MaxScore))
	}
	total := float64(0)
	for i := 0; i < len(score); i++ {
		if score[i] > max[i] {
			score[i] = max[i]
		}
		total += (score[i] / max[i]) * (weight[i] / totalWeight)
	}
	return uint32(math.Round(total * 100))
}
