package score

import (
	"fmt"
	"math"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Results contains the score objects, build info, and errors.
type Results struct {
	BuildInfo    *BuildInfo // build info for tests
	Scores       []*Score   // list of scores for different tests
	ParsedScores int        // number of valid score lines parsed from the test output
	testNames    []string   // defines the order
	scoreMap     map[string]*Score
	source       map[string]*testRun // score name to the test that emitted it; see attach
}

func newResults(scores ...*Score) *Results {
	r := &Results{
		testNames: make([]string, 0),
		scoreMap:  make(map[string]*Score),
		source:    make(map[string]*testRun),
	}
	for _, sc := range scores {
		r.addScore(sc, nil)
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

// addScore adds the given score to the set of scores.
// This method assumes that the provided score object is valid.
// run identifies the emitting test, or is nil for a score without test markers.
func (r *Results) addScore(sc *Score, run *testRun) {
	testName := sc.GetTestName()
	if current, found := r.scoreMap[testName]; found {
		if current.GetScore() != 0 {
			// We reach here only if a second non-zero score is found for the same test.
			// Mark it as faulty with -1.
			sc.Score = -1
			sc.TestDetails = "(duplicate)"
		}
	} else {
		// New test: record in r.testNames
		r.testNames = append(r.testNames, testName)
	}

	// Record score object if:
	// - current score is nil or zero, or
	// - the first score was zero.
	r.scoreMap[testName] = sc
	if run != nil {
		r.source[testName] = run
	}
}

// validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
//
// This method is only used for testing. The actual validation is done in the
// ExtractResults method when parsing the output of a test execution.
func (r *Results) validate(secret string) error {
	for _, sc := range r.Scores {
		if err := sc.isValid(secret); err != nil {
			return err
		}
	}
	return nil
}

// Sum returns the total score the of recorded scores.
// The total is a grade in the range 0-100.
// This method must only be called after Validate has returned nil.
func (r *Results) Sum() uint32 {
	totalWeight := float64(0)
	var maxScore, score, weight []float64
	for _, ts := range r.Scores {
		// If the score is negative, it means that the test is faulty (e.g. duplicate).
		// We need to set the score to zero to avoid certain edge cases where
		// the total score would end up being -1 or lower. If not, the total score
		// would end up being uint32(-1) = 4294967295. See issue #975
		testScore := max(ts.GetScore(), 0)
		totalWeight += float64(ts.GetWeight())
		weight = append(weight, float64(ts.GetWeight()))
		score = append(score, float64(testScore))
		maxScore = append(maxScore, float64(ts.GetMaxScore()))
	}
	total := float64(0)
	for i := 0; i < len(score); i++ {
		if score[i] > maxScore[i] {
			score[i] = maxScore[i]
		}
		total += weightedScore(score[i], maxScore[i], weight[i], totalWeight)
	}
	return uint32(math.Round(total * 100))
}

// weightedScore returns the weighted score of a given test.
func weightedScore(score, maxScore, weight, totalWeight float64) float64 {
	return (score / maxScore) * (weight / totalWeight)
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

// ExtractResults returns the results from a test execution extracted from
// the given out string. The provided zeroScoreTests must contain a zero
// score value for all tests that are expected to be present in the results.
//
// The output is attributed to the tests that produced it, so that each score
// carries what its own test printed, the outcome the test framework reported
// for it, and how long it took. The build log keeps what belonged to no test:
// the run script's own output, the compilation phase, and anything printed
// after the tests finished.
func ExtractResults(out, secret string, execTime time.Duration, zeroScoreTests []*Score) (*Results, error) {
	// add all expected tests (assumed to already have zero scores)
	results := newResults(zeroScoreTests...)
	log := scanTestLog(out)
	errs := results.parseScores(log.Scores, nil, secret)
	for _, run := range log.Tests {
		errs = append(errs, results.parseScores(run.Scores, run, secret)...)
	}

	results.attach(log, secret)
	results.BuildInfo = &BuildInfo{
		BuildDate:      timestamppb.Now(),
		SubmissionDate: timestamppb.Now(),
		BuildLog:       Redact(strings.Join(log.Unattributed, "\n"), secret),
		ExecTime:       execTime.Milliseconds(),
	}
	if len(errs) > 0 {
		return results, errs
	}
	return results, nil
}

func (r *Results) parseScores(lines []string, run *testRun, secret string) parseErrors {
	var errs parseErrors
	for _, line := range lines {
		sc, err := parse(line, secret)
		if err != nil {
			errs = append(errs, fmt.Errorf("parsing line '%s': %w", line, err))
			continue
		}
		if _, expected := r.scoreMap[sc.GetTestName()]; !expected {
			continue
		}
		r.ParsedScores++
		r.addScore(sc, run)
	}
	return errs
}

// attach records on each score the run of the test that emitted it, and passes
// the output of the remaining tests to retainOutput.
//
// The emitting test need not be the test the score is named after. Go rewrites
// a subtest's name, so a score registered as "TestX/Six jobs" is printed by
// "TestX/Six_jobs", and a score looked up in one test may be printed from a
// nested subtest. A score printed outside any test, as in a run without -v,
// has no emitting test, so the test of the same name is used if the log has one.
func (r *Results) attach(log *testLog, secret string) {
	r.Scores = r.toScoreSlice()
	scoredRuns := make(map[string]*testRun)
	for _, sc := range r.Scores {
		testName := sc.GetTestName()
		run := r.source[testName]
		if run == nil {
			// Without an emitting test, match the registered test name.
			run = log.Test(testName)
			r.source[testName] = run
		}
		if run != nil {
			// Mark the run as scored, by Go test name, so that retainOutput
			// leaves its output on its own score.
			scoredRuns[run.Name] = run
		}
	}
	// Retain the output of tests that emitted no score, so that it is not lost.
	log.retainOutput(scoredRuns)
	for _, sc := range r.Scores {
		sc.attach(r.source[sc.GetTestName()], secret)
	}
}

// attach records on the score what the run of its test produced. A test that
// reported its own details keeps them: they are the teacher's own words, and
// the diagnostics scraped from the output would only repeat them. Those details
// are redacted too, since the score object is free to put anything in them.
func (s *Score) attach(run *testRun, secret string) {
	s.TestDetails = Redact(s.GetTestDetails(), secret)
	s.TestOutput = Redact(s.GetTestOutput(), secret)
	if run == nil {
		return
	}
	s.Status = run.Status
	s.Elapsed = run.Elapsed
	s.TestOutput = Redact(strings.Join(run.Output, "\n"), secret)
	if s.GetTestDetails() == "" {
		s.TestDetails = Redact(strings.Join(run.Failures, "\n"), secret)
	}
}

// GetBuildInfo returns the build info for the results object after nil check.
func (r *Results) GetBuildInfo() *BuildInfo {
	if r != nil && r.BuildInfo != nil {
		return r.BuildInfo
	}
	return &BuildInfo{}
}
