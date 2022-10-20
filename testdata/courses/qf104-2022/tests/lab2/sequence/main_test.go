package sequence

import (
	"fmt"
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

var scores = score.NewRegistry()

func TestMain(m *testing.M) {
	scores.PrintTestInfo()
	if os.Getenv("SCORE_INIT") != "" {
		os.Exit(0)
	}
	exitCode := m.Run()
	if err := scores.Validate(); err != nil {
		fmt.Println(err)
	}
	os.Exit(exitCode)
}
