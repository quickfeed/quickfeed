package score

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
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

// TODO(meling) prefer to keep this private
// TODO(meling) should also check that 'name' is a proper Test function (ref the deleted checkTest func in earlier commit)
func TestName(x interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name()
	return lastElem(name)
}

func lastElem(name string) string {
	return name[strings.LastIndex(name, ".")+1:]
}

// Add test with given max score and weight to the registry.
func Add(test interface{}, max, weight int) {
	testName := TestName(test)
	if _, found := scores[testName]; found {
		frame := getFrame(3)
		panic(fmt.Errorf("%s:%d: duplicate score test: %s", filepath.Base(frame.File), frame.Line, testName))
	}
	sc := &Score{
		Secret:   sessionSecret,
		TestName: testName,
		MaxScore: int32(max),
		Weight:   int32(weight),
	}
	scores[testName] = sc
}

// Add dynamically test: can be lost
func A(testName string, max, weight int) {
	if _, found := scores[testName]; found {
		frame := getFrame(3)
		panic(fmt.Errorf("%s:%d: duplicate score test: %s", filepath.Base(frame.File), frame.Line, testName))
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
	frame := getFrame(3)
	panic(fmt.Errorf("%s:%d: unknown score test: %s", filepath.Base(frame.File), frame.Line, testName))
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
	frame := getFrame(4)
	testName := lastElem(frame.Function)
	if sc, ok := scores[testName]; ok {
		return sc
	}
	panic(fmt.Errorf("%s:%d: unknown score test: %s", filepath.Base(frame.File), frame.Line, testName))
}

func getFrame(skip int) runtime.Frame {
	pc := make([]uintptr, 15)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame
}
