package score

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/kit/score/testlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Results contains the score objects, build info, and errors.
type Results struct {
	BuildInfo    *BuildInfo // build info for tests
	Scores       []*Score   // list of scores for different tests
	ParsedScores int        // number of valid score lines parsed from the test output
	testNames    []string   // defines the order
	scoreMap     map[string]*Score
}

func newResults(scores ...*Score) *Results {
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

// addScore adds the given score to the set of scores.
// This method assumes that the provided score object is valid.
func (r *Results) addScore(sc *Score) {
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

// ExtractResults returns the results from a test execution extracted from the given out string.
// The provided zeroScoreTests must contain a zero score value for all tests that are expected
// to be present in the results.
//
// The output is attributed to the tests that produced it, so that each score
// carries what its own test printed, the outcome the test framework reported
// for it, and how long it took. The build log keeps what belonged to no test:
// the run script's own output, the compilation phase, and anything printed
// after the tests finished.
func ExtractResults(out, secret string, execTime time.Duration, zeroScoreTests []*Score) (*Results, error) {
	errs := make(parseErrors, 0)
	results := newResults()
	parsedScores := 0

	// first, add all expected tests (assumed to already have zero scores)
	for _, expectedTest := range zeroScoreTests {
		results.addScore(expectedTest)
	}

	log := testlog.Scan(out, HasPrefix)
	// A test's name in the score object need not be the name of the Go test
	// that emitted it, so the emitting test is recorded as the score's source.
	source := make(map[string]*testlog.TestRun)
	addScore := func(line string, run *testlog.TestRun) {
		sc, err := parse(line, secret)
		if err != nil {
			errs = append(errs, fmt.Errorf("parsing line '%s': %w", line, err))
			return
		}
		// only add the score if it's in the expected tests
		if !slices.ContainsFunc(zeroScoreTests, func(expected *Score) bool {
			return expected.GetTestName() == sc.GetTestName()
		}) {
			return
		}
		parsedScores++
		results.addScore(sc)
		if run != nil {
			source[sc.GetTestName()] = run
		}
	}
	for _, line := range log.Scores {
		addScore(line, nil)
	}
	for _, run := range log.Tests {
		for _, line := range run.Scores {
			addScore(line, run)
		}
	}

	scores := results.toScoreSlice()
	for _, sc := range scores {
		run, ok := source[sc.GetTestName()]
		if !ok {
			// The score line was not attributed to a test, either because it
			// was printed outside one or because no score line was parsed for
			// this expected test at all. Fall back to the test of that name.
			run = log.Test(sc.GetTestName())
		}
		sc.attach(run, secret)
	}

	res := &Results{
		BuildInfo: &BuildInfo{
			BuildDate:      timestamppb.Now(),
			SubmissionDate: timestamppb.Now(),
			BuildLog:       Redact(strings.Join(log.Unattributed, "\n"), secret),
			ExecTime:       execTime.Milliseconds(),
		},
		Scores:       scores,
		ParsedScores: parsedScores,
	}
	if len(errs) > 0 {
		return res, errs
	}
	return res, nil
}

// attach records on the score what the run of its test produced. A test that
// reported its own details keeps them: they are the teacher's own words, and
// the diagnostics scraped from the output would only repeat them.
func (s *Score) attach(run *testlog.TestRun, secret string) {
	if run == nil {
		return
	}
	s.Status = statusOf(run.Status)
	s.Elapsed = run.Elapsed
	s.TestOutput = Redact(strings.Join(run.Output, "\n"), secret)
	if s.GetTestDetails() == "" {
		s.TestDetails = Redact(strings.Join(run.Failures, "\n"), secret)
	}
}

func statusOf(status testlog.Status) TestStatus {
	switch status {
	case testlog.StatusPassed:
		return TestStatus_PASSED
	case testlog.StatusFailed:
		return TestStatus_FAILED
	case testlog.StatusSkipped:
		return TestStatus_SKIPPED
	default:
		return TestStatus_NOT_RUN
	}
}

// GetBuildInfo returns the build info for the results object after nil check.
func (r *Results) GetBuildInfo() *BuildInfo {
	if r != nil && r.BuildInfo != nil {
		return r.BuildInfo
	}
	return &BuildInfo{}
}
