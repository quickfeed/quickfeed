package sequence

import (
	"os"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

func TestMain(m *testing.M) {
	score.PrintTestInfo()
	os.Exit(m.Run())
}
