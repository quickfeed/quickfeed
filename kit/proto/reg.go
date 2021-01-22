package proto

import (
	"os"
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

func Add(testName string, max, weight int32) {
	sc := &Score{
		Secret:   sessionSecret,
		TestName: testName,
		MaxScore: max,
		Weight:   weight,
	}
	scores[testName] = sc
}

func Get(testName string) *Score {
	return scores[testName]
}

// func TestX(t *testing.T) {
// 	sc := Get(t.Name())
// }
