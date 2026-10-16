package score

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/kit/internal/test"
)

// registry keeps a map of score objects and a slice of test names,
// in registration order, which is used to preserve deterministic iteration order.
type registry struct {
	testNames []string          // testNames in registration order
	scores    map[string]*Score // map from TestName to score object
}

func NewRegistry() *registry { // skipcq: RVV-B0011
	return &registry{
		testNames: make([]string, 0),
		scores:    make(map[string]*Score),
	}
}

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func (s *registry) Validate() error {
	test.CallFrame()
	for _, sc := range s.scores {
		if err := sc.isValid(sessionSecret); err != nil {
			return err
		}
	}
	return nil
}

// PrintTestInfo prints a JSON line for each registered test, in registration
// order, and exits, but only if the environment variable SCORE_INIT is set
// to a non-empty value. Otherwise, it does nothing.
//
// This should be called after test registration has been completed,
// but before test execution. This can be done in TestMain.
//
// Will panic if called from a non-test function.
func (s *registry) PrintTestInfo() {
	test.CallFrame()
	if os.Getenv("SCORE_INIT") == "" {
		return
	}
	for _, name := range s.testNames {
		if sc, ok := s.scores[name]; ok {
			fmt.Println(sc.json())
		}
	}
	os.Exit(0) // skipcq: RVV-A0003
}

// Add test with given max score and weight to the registry.
//
// Will panic if the test has already been registered or if max or weight is non-positive.
func (s *registry) Add(testFn any, maxScore, weight int) {
	s.internalAdd(test.Name(testFn), maxScore, weight)
}

// Max wraps a registered test with its score initially at MaxScore.
// Subtests started with T.Run share this score.
func (s *registry) Max(t *testing.T) *T {
	return s.wrap(t, true)
}

// Min wraps a registered test with its score initially at zero.
// Subtests started with T.Run share this score.
func (s *registry) Min(t *testing.T) *T {
	return s.wrap(t, false)
}

var (
	ErrDuplicateScoreTest = errors.New("duplicate score test")
	ErrUnauthorizedLookup = errors.New("unauthorized lookup")
	ErrUnknownScoreTest   = errors.New("unknown score test")
)

func (s *registry) internalAdd(testName string, maxScore, weight int) {
	if _, found := s.scores[testName]; found {
		panic(test.ErrMsg(testName, ErrDuplicateScoreTest.Error()))
	}
	if maxScore < 1 {
		panic(test.ErrMsg(testName, ErrMaxScore.Error()))
	}
	if weight < 1 {
		panic(test.ErrMsg(testName, ErrWeight.Error()))
	}
	sc := &Score{
		Secret:   sessionSecret,
		TestName: testName,
		MaxScore: int32(maxScore),
		Weight:   int32(weight),
	}
	// record the TestName in separate slice to preserve registration order
	s.testNames = append(s.testNames, testName)
	s.scores[testName] = sc
}

func (s *registry) wrap(t *testing.T, maxScore bool) *T {
	t.Helper()
	sc, ok := s.scores[t.Name()]
	if !ok {
		t.Fatalf("score: test %q is not registered; call Add before running it", t.Name())
	}
	if maxScore {
		sc.Score = sc.GetMaxScore()
	}
	wrapped := &T{T: t, st: &state{sc: sc, max: maxScore}}
	t.Cleanup(wrapped.print)
	return wrapped
}
