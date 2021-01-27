package score

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

var scores = make(map[string]*Score)

// Add test with given max score and weight to the registry.
func Add(test interface{}, max, weight int) {
	add(testName(test), max, weight)
}

// AddSub test with given max score and weight to the registry.
// This function should be used to register subtests, and should be used in
// conjunction with MaxByName and MinByName called from within a subtest.
func AddSub(test interface{}, subTestName string, max, weight int) {
	tstName := fmt.Sprintf("%s/%s", testName(test), subTestName)
	add(tstName, max, weight)
}

// Max returns a score object with Score equal to MaxScore.
// The returned score object should be used with score.Dec() and score.DecBy().
//
// May panic with unknown score test, if the test hasn't been added.
func Max() *Score {
	testName := callerTestName()
	sc := get(testName)
	sc.Score = sc.GetMaxScore()
	return sc
}

// Min returns a score object with Score equal to zero.
// The returned score object should be used with score.Inc() and score.IncBy().
//
// May panic with unknown score test, if the test hasn't been added.
func Min() *Score {
	testName := callerTestName()
	return get(testName)
}

// MaxByName returns score object for the given test name with Score equal to MaxScore.
// This function is meant to be used from within subtests, and in conjunction with AddSub.
// The returned score object should be used with score.Dec() and score.DecBy().
//
// May panic with unknown score test, if the test hasn't been added.
func MaxByName(testName string) *Score {
	sc := get(testName)
	sc.Score = sc.GetMaxScore()
	return sc
}

// MinByName returns a score object for the given test name with Score equal to zero.
// This function is meant to be used from within subtests, and in conjunction with AddSub.
// The returned score object should be used with score.Inc() and score.IncBy().
//
// May panic with unknown score test, if the test hasn't been added.
func MinByName(testName string) *Score {
	return get(testName)
}

func testName(testFn interface{}) string {
	typ := reflect.TypeOf(testFn)
	if typ.Kind() != reflect.Func {
		panic(errMsg(reflect.ValueOf(testFn), "not a function"))
	}
	name := runtime.FuncForPC(reflect.ValueOf(testFn).Pointer()).Name()
	name = lastElem(name)
	if typ.NumIn() != 1 || typ.NumOut() > 0 || !strings.HasPrefix(name, "Test") {
		panic(errMsg(name, "not a test function"))
	}
	if !typ.In(0).AssignableTo(reflect.TypeOf(&testing.T{})) {
		panic(errMsg(name, "test function missing *testing.T argument"))
	}
	return name
}

func callerTestName() string {
	frame := callFrame()
	return lastElem(frame.Function)
}

func errMsg(testFn interface{}, msg string) error {
	frame := callFrame()
	return fmt.Errorf("%s:%d: %s: %v", filepath.Base(frame.File), frame.Line, msg, testFn)
}

func lastElem(name string) string {
	return name[strings.LastIndex(name, ".")+1:]
}

func add(testName string, max, weight int) {
	if _, found := scores[testName]; found {
		panic(errMsg(testName, "duplicate score test"))
	}
	if max < 1 {
		panic(errMsg(testName, "max must be greater than 0"))
	}
	if weight < 1 {
		panic(errMsg(testName, "weight must be greater than 0"))
	}
	sc := &Score{
		Secret:   sessionSecret,
		TestName: testName,
		MaxScore: int32(max),
		Weight:   int32(weight),
	}
	scores[testName] = sc
}

func get(testName string) *Score {
	if sc, ok := scores[testName]; ok {
		return sc
	}
	panic(errMsg(testName, "unknown score test"))
}
