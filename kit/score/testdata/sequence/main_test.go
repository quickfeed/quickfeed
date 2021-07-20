package sequence

import (
	"fmt"
	"os"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

var scores = score.NewRegistry()

func TestMain(m *testing.M) {
	scores.PrintTestInfo()
	exitCode := m.Run()
	if err := scores.Validate(); err != nil {
		fmt.Println(err)
	}
	os.Exit(exitCode)
}
