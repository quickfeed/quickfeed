package score

import (
	"errors"
	"fmt"
	"os"
	"sort"

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

// PrintTestInfo prints a JSON representation of all registered tests
// in the order they were registered, or if sorted is true the test names
// will be sorted before printing.
//
// This should be called after test registration has been completed,
// but before test execution. This can be done in TestMain.
//
// If the environment variable SCORE_INFO is set to a non-empty value,
// the test info will be printed and the program will exit.
// This can be used to ensure that the test info is always printed;
// otherwise, a test failure may prevent the test info from being printed.
//
// Will panic if called from a non-test function.
func (s *registry) PrintTestInfo(sorted ...bool) {
	test.CallFrame()
	if len(sorted) == 1 && sorted[0] {
		sort.Strings(s.testNames)
	}
	// iterate over the test names in registration or sorted order
	for _, name := range s.testNames {
		if sc, ok := s.scores[name]; ok {
			fmt.Println(sc.Json())
		}
	}
	// force exit after printing test info if SCORE_INIT is set
	if os.Getenv("SCORE_INIT") != "" {
		os.Exit(0) // skipcq: RVV-A0003
	}
}

// Add test with given max score and weight to the registry.
//
// Will panic if the test has already been registered or if max or weight is non-positive.
func (s *registry) Add(testFn any, max, weight int) {
	s.internalAdd(test.Name(testFn), "", max, weight)
}

// AddWithTask test with given taskName, max score and weight to the registry.
// This function is identical to Add, with the addition of assigning a task name.
//
// Will panic if the test has already been registered or if max or weight is non-positive.
func (s *registry) AddWithTask(testFn any, taskName string, max, weight int) {
	s.internalAdd(test.Name(testFn), taskName, max, weight)
}

// AddSub test with given max score and weight to the registry.
// This function should be used to register subtests, and should be used in
// conjunction with MaxByName and MinByName called from within a subtest.
//
// Will panic if the test has already been registered or if max or weight is non-positive.
func (s *registry) AddSub(testFn any, subTestName string, max, weight int) {
	tstName := fmt.Sprintf("%s/%s", test.Name(testFn), subTestName)
	s.internalAdd(tstName, "", max, weight)
}

// AddSubWithTask test with given taskName, max score and weight to the registry.
// This function should be used to register subtests, and should be used in
// conjunction with MaxByName and MinByName called from within a subtest.
// This function is identical to AddSub, with the addition of assigning a task name.
//
// Will panic if the test has already been registered or if max or weight is non-positive.
func (s *registry) AddSubWithTask(testFn any, subTestName, taskName string, max, weight int) {
	tstName := fmt.Sprintf("%s/%s", test.Name(testFn), subTestName)
	s.internalAdd(tstName, taskName, max, weight)
}

// Max returns a score object with Score equal to MaxScore.
// The returned score object should be used with score.Dec() and score.DecBy().
//
// Will panic with unknown score test, if the test hasn't been added.
func (s *registry) Max() *Score {
	testName := test.CallerName()
	sc := s.get(testName)
	sc.Score = sc.GetMaxScore()
	return sc
}

// Min returns a score object with Score equal to zero.
// The returned score object should be used with score.Inc() and score.IncBy().
//
// Will panic with unknown score test, if the test hasn't been added.
func (s *registry) Min() *Score {
	testName := test.CallerName()
	return s.get(testName)
}

// MaxByName returns score object for the given test name with Score equal to MaxScore.
// This function is meant to be used from within subtests, and in conjunction with AddSub.
// The returned score object should be used with score.Dec() and score.DecBy().
//
// Will panic with unknown score test, if the test hasn't been added.
func (s *registry) MaxByName(testName string) *Score {
	sc := s.get(testName)
	sc.Score = sc.GetMaxScore()
	return sc
}

// MinByName returns a score object for the given test name with Score equal to zero.
// This function is meant to be used from within subtests, and in conjunction with AddSub.
// The returned score object should be used with score.Inc() and score.IncBy().
//
// Will panic with unknown score test, if the test hasn't been added.
func (s *registry) MinByName(testName string) *Score {
	return s.get(testName)
}

var (
	ErrDuplicateScoreTest = errors.New("duplicate score test")
	ErrUnauthorizedLookup = errors.New("unauthorized lookup")
	ErrUnknownScoreTest   = errors.New("unknown score test")
)

func (s *registry) internalAdd(testName, taskName string, max, weight int) {
	if _, found := s.scores[testName]; found {
		panic(test.ErrMsg(testName, ErrDuplicateScoreTest.Error()))
	}
	if max < 1 {
		panic(test.ErrMsg(testName, ErrMaxScore.Error()))
	}
	if weight < 1 {
		panic(test.ErrMsg(testName, ErrWeight.Error()))
	}
	sc := &Score{
		Secret:   sessionSecret,
		TestName: testName,
		TaskName: taskName,
		MaxScore: int32(max),
		Weight:   int32(weight),
	}
	// record the TestName in separate slice to preserve registration order
	s.testNames = append(s.testNames, testName)
	s.scores[testName] = sc
}

func (s *registry) get(testName string) *Score {
	if !test.IsCaller(testName) {
		// Only the registered Test function can call the lookup functions
		panic(test.ErrMsg(testName, ErrUnauthorizedLookup.Error()))
	}
	if sc, ok := s.scores[testName]; ok {
		return sc
	}
	panic(test.ErrMsg(testName, ErrUnknownScoreTest.Error()))
}
