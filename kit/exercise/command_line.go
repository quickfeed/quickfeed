package exercise

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

type SearchType int

const (
	ResultEquals SearchType = iota
	ResultContains
	ResultDoesNotContain
)

type Commands []struct {
	Command string
	Result  string
	Search  SearchType
}

// CommandLine computes the score for a set of command line exercises
// that students provided. The function requires the list of commands
// and their expected answers, and a Score object. The function
// will produce both string output and JSON output.
func CommandLine(t *testing.T, sc *score.Score, answers Commands) {
	defer sc.WriteString(os.Stdout)
	defer sc.WriteJSON(os.Stdout)

	for i := range answers {
		cmdArgs := strings.Split(answers[i].Command, " ")
		cmd := exec.Command(cmdArgs[0])
		cmd.Args = cmdArgs
		var sout, serr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &sout, &serr
		err := cmd.Run()
		if err != nil {
			t.Errorf("%v\n%v: %v.\n", sc.TestName, err, serr.String())

			// If length of stdout > 0, then the application probably puts its error output in stdout
			// instead of stderr. In that case we want to check the contents of stdout in the switch
			// statement below to determine whether to decrement the score.
			if sout.Len() == 0 {
				sc.Dec()
				continue
			}
		}

		outStr := sout.String()
		// Compare output with expected output
		switch answers[i].Search {
		case ResultEquals:
			if outStr != answers[i].Result {
				t.Errorf("%v: \ngot: %v \nwant: %v \nfor command: %v\n",
					sc.TestName, outStr, answers[i].Result, answers[i].Command)
				sc.Dec()
			}
		case ResultContains:
			if !strings.Contains(outStr, answers[i].Result) {
				t.Errorf("%v: \nResult does not contain: %v \nfor command: %v\n",
					sc.TestName, answers[i].Result, answers[i].Command)
				sc.Dec()
			}
		case ResultDoesNotContain:
			if strings.Contains(outStr, answers[i].Result) {
				t.Errorf("%v: \nResult contains: %v \nfor command: %v\n",
					sc.TestName, answers[i].Result, answers[i].Command)
				sc.Dec()
			}
		}
	}
}
