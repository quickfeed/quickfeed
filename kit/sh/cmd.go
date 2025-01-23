package sh

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/kit/internal/test"
)

// MustRun runs the given command with wd as the working directory,
// and prints the output to stdout/stderr if output is true.
// If the command fails, MustRun panics.
func MustRun(output bool, wd, cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Dir = wd
	if output {
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		log.Println("running:", cmd, strings.Join(args, " "))
	}
	if err := c.Run(); err != nil {
		panic(fmt.Sprintf("failed to run %s %v: %v", cmd, args, err))
	}
}

// Run runs the given command, directing stderr to the command's stderr and
// printing stdout to stdout. Run returns an error if any.
func Run(cmd string) error {
	s := strings.Split(cmd, " ")
	return RunA(s[0], s[1:]...)
}

// RunA runs the given command, directing stderr to the command's stderr and
// printing stdout to stdout. RunA returns an error if any.
func RunA(cmd string, args ...string) error {
	_, err := internalRun(os.Stdout, os.Stderr, cmd, args...)
	return err
}

// Output runs the command and returns the text from stdout, or an error.
// The command's stderr is sent to stderr.
func Output(cmd string) (string, error) {
	s := strings.Split(cmd, " ")
	return OutputA(s[0], s[1:]...)
}

// OutputA runs the command and returns the text from stdout, or an error.
// The command's stderr is sent to stderr.
func OutputA(cmd string, args ...string) (string, error) {
	stdout := &bytes.Buffer{}
	_, err := internalRun(stdout, os.Stderr, cmd, args...)
	return strings.TrimSuffix(stdout.String(), "\n"), err
}

// OutputErr runs the command and returns the text from stdout and stderr, or an error.
func OutputErr(cmd string) (string, string, error) {
	s := strings.Split(cmd, " ")
	return OutputErrA(s[0], s[1:]...)
}

// OutputErrA runs the command and returns the text from stdout and stderr, or an error.
func OutputErrA(cmd string, args ...string) (string, string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	_, err := internalRun(stdout, stderr, cmd, args...)
	return strings.TrimSuffix(stdout.String(), "\n"), strings.TrimSuffix(stderr.String(), "\n"), err
}

// RunCountTest runs the given test count times.
// It returns the test output and false if the test passed.
func RunCountTest(tst func(*testing.T), count, tags string) (s string, fail bool) {
	testName := test.Name(tst)
	if tags != "" {
		s, _ = OutputA("go", "test", "-run", testName, "-count", count, "-tags", tags)
	} else {
		s, _ = OutputA("go", "test", "-run", testName, "-count", count)
	}
	return s, strings.Contains(s, "FAIL")
}

// RunRaceTest runs the given test with the race detector enabled.
// It returns the test output and false if there weren't any data races.
// Otherwise, it returns the stack trace and true if there was a data race.
//
// The test to be run with the race detector should be in a separate file with
// the race build tag. See the race_test.go file in this package for an example.
//
// If the tags argument is non-zero, it is passed to the go test command.
func RunRaceTest(tst func(*testing.T), tags string) (s string, race bool) {
	testName := test.Name(tst)
	if tags != "" {
		s, _ = OutputA("go", "test", "-v", "-race", "-run", testName, "-tags", tags)
	} else {
		s, _ = OutputA("go", "test", "-v", "-race", "-run", testName)
	}
	return s, strings.Contains(s, "WARNING: DATA RACE")
}

func internalRun(stdout, stderr io.Writer, cmd string, args ...string) (ran bool, err error) {
	c := exec.Command(cmd, args...)
	c.Stderr = stderr
	c.Stdout = stdout
	log.Println("running:", cmd, strings.Join(args, " "))
	err = c.Run()
	return cmdRan(err), err
}

func cmdRan(err error) bool {
	if err == nil {
		return true
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.Exited()
	}
	return false
}
