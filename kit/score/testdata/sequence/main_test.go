package sequence

import (
	"fmt"
	"os"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

func TestMain(m *testing.M) {
	score.PrintTestInfo()
	exitCode := m.Run()
	if err := score.Validate(); err != nil {
		fmt.Println(err)
	}
	os.Exit(exitCode)
}
