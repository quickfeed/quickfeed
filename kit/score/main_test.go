package score_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

var scores = score.NewRegistry()

func TestMain(m *testing.M) {
	fmt.Println("Registration order:")
	scores.PrintTestInfo()
	// print also in sorted order for the benefit of TestPrintTestInfoOrder
	fmt.Println("Sorted order:")
	scores.PrintTestInfo(true)
	exitCode := m.Run()
	if err := scores.Validate(); err != nil {
		fmt.Println(err)
	}
	os.Exit(exitCode)
}
