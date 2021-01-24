package score

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

const (
	secretEnvName = "QUICKFEED_SESSION_SECRET"
)

var (
	sessionSecret string
	scores        = make(map[string]*Score)
)

func init() {
	sessionSecret = os.Getenv(secretEnvName)
	// remove variable as soon as it has been read
	_ = os.Setenv(secretEnvName, "")
}

func TestName(testFn interface{}) string {
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

func errMsg(testFn interface{}, msg string) error {
	frame := callFrame()
	return fmt.Errorf("%s:%d: %s: %v", filepath.Base(frame.File), frame.Line, msg, testFn)
}

func lastElem(name string) string {
	return name[strings.LastIndex(name, ".")+1:]
}

// Add test with given max score and weight to the registry.
func Add(test interface{}, max, weight int) {
	add(TestName(test), max, weight)
}

// AddSubtest with given max score and weight to the registry.
func AddSubtest(test interface{}, subTestName string, max, weight int) {
	testName := fmt.Sprintf("%s/%s", TestName(test), subTestName)
	add(testName, max, weight)
}

func add(testName string, max, weight int) {
	if _, found := scores[testName]; found {
		panic(errMsg(testName, "duplicate score test"))
	}
	sc := &Score{
		Secret:   sessionSecret,
		TestName: testName,
		MaxScore: int32(max),
		Weight:   int32(weight),
	}
	scores[testName] = sc
}

func stripPkg(name string) string {
	start := strings.LastIndex(name, "/") + 1
	dot := strings.Index(name[start:], ".") + 1
	return name[start+dot:]
}

func GMax(testName string) *Score {
	frame := callFrame()
	fmt.Printf("%s:%d: %s: %v\n", filepath.Base(frame.File), frame.Line, stripPkg(frame.Function), frame.Function)
	if sc, ok := scores[testName]; ok {
		sc.Score = sc.GetMaxScore()
		return sc
	}
	panic(errMsg(testName, "unknown score test"))
}

// GetMax returns a score object with Score equal to MaxScore.
// The returned score object should be used with score.Dec() and score.DecBy().
func GetMax() *Score {
	sc := get()
	sc.Score = sc.GetMaxScore()
	return sc
}

// Get returns a score object with Score equal to zero.
// The returned score object should be used with score.Inc() and score.IncBy().
func Get() *Score {
	return get()
}

func get() *Score {
	frame := callFrame()
	testName := lastElem(frame.Function)
	if sc, ok := scores[testName]; ok {
		return sc
	}
	panic(errMsg(testName, "unknown score test"))
}

// TODO(meling) rename test_registry.go to registry.go
