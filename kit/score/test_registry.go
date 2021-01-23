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
func AddSubtest(testName string, max, weight int) {
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

func GMax(testName string) *Score {
	if sc, ok := scores[testName]; ok {
		sc.Score = sc.GetMaxScore()
		return sc
	}
	panic(errMsg(testName, "unknown score test"))
}

// GetMax returns a score object initialized with Score equal to MaxScore.
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
	panic(fmt.Errorf("%s:%d: unknown score test: %s", filepath.Base(frame.File), frame.Line, testName))
}

// TODO(meling) rename test_registry.go to registry.go
