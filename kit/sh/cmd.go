package sh

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

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

// RunRaceTest runs the given test with the race detector enabled.
// It returns the test output and false if there weren't any data races.
// Otherwise, it returns the stack trace and true if there was a data race.
//
// The test to be run with the race detector should be in a separate file with
// the race build tag. See the race_test.go file in this package for an example.
func RunRaceTest(testName string) (string, bool) {
	s, _ := OutputA("go", "test", "-v", "-race", "-run", testName)
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
