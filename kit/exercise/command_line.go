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

// CommandLineError contains the stdout, stderr and error from a command line
// execution returned by the CommandLine function.
type CommandLineError struct {
	stdOut string
	stdErr string
	err    error
}

func (e CommandLineError) StdOut() string {
	return e.stdOut
}

func (e CommandLineError) StdErr() string {
	return e.stdErr
}

func (e *CommandLineError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

// CommandLine computes the score for a set of command line exercises
// that students provided. The function requires the list of commands
// and their expected answers, and a Score object. The function
// will produce both string output and JSON output.
//
// In addition to displaying error output through the t argument, the
// output written to stdout and stderr, along with the error message
// of each command, are returned in the CommandLineError slice,
// where indices match those of the commands.
func CommandLine(t *testing.T, sc *score.Score, answers Commands) []CommandLineError {
	defer sc.WriteString(os.Stdout)
	defer sc.WriteJSON(os.Stdout)

	cmdLineErrors := make([]CommandLineError, len(answers))
	for i := range answers {
		cmdArgs := strings.Split(answers[i].Command, " ")
		cmd := exec.Command(cmdArgs[0])
		cmd.Args = cmdArgs
		var stdOut, stdErr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdOut, &stdErr
		err := cmd.Run()
		// Store the outputs and error to be returned
		cmdLineErrors[i].stdOut = stdOut.String()
		cmdLineErrors[i].stdErr = stdErr.String()
		cmdLineErrors[i].err = err
		if err != nil {
			t.Errorf("%v\n%v: %v.\n", sc.TestName, err, stdErr.String())

			// If length of stdout > 0, then the application probably puts its error output in stdout
			// instead of stderr. In that case we want to check the contents of stdout in the switch
			// statement below to determine whether to decrement the score.
			if stdOut.Len() == 0 {
				sc.Dec()
				continue
			}
		}

		outStr := stdOut.String()
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

	return cmdLineErrors
}
