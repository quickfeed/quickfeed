// Command qcm is QuickFeed's course manager: a command line tool for teachers
// to clone their course repositories and run and check their course's tests
// locally, with the same test runner the QuickFeed server uses.
//
// The tests are always executed in Docker, exactly as on the server, so a
// working Docker installation is required for the run and check subcommands.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	// localCommitID names the containers of the test runs this command starts.
	// Two concurrent runs of the same assignment for the same owner therefore
	// conflict, which is what we want: they would test the same code twice.
	localCommitID = "local"
	// localOwner is the job owner, and thereby the repository name, of a run
	// against a local submission directory. The name must not collide with the
	// tests and assignments mounts; see ci.RunData.SubmissionDir.
	localOwner = "local"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "qcm: %v\n", err)
		os.Exit(1)
	}
}

// run dispatches to the subcommand named by the first argument. It returns an
// error for anything that should make the tool exit non-zero; a request for
// help is not such an error.
func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stderr)
		return errors.New("no subcommand given")
	}
	var err error
	switch args[0] {
	case "clone":
		err = cloneCmd(args[1:], stdout, stderr)
	case "run":
		err = runCmd(args[1:], stdout, stderr)
	case "check":
		err = checkCmd(args[1:], stdout, stderr)
	case "help", "-h", "-help", "--help":
		usage(stdout)
		return nil
	default:
		usage(stderr)
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
	if errors.Is(err, flag.ErrHelp) {
		// The subcommand's flag set has already printed its usage text.
		return nil
	}
	return err
}

func usage(w io.Writer) {
	fmt.Fprint(w, `qcm is QuickFeed's course manager.

Usage:

	qcm <command> [flags]

The commands are:

	clone   clone the course's tests and assignments repositories
	run     run a course's tests for one assignment
	check   check that the course's test environment is working

Every command takes -course ORG, naming the course's GitHub organization,
e.g., dat320-2025. Run 'qcm <command> -help' for the command's own flags.

The run and check commands execute the tests in Docker, just like the
QuickFeed server does, and therefore require a working Docker installation.
`)
}

// newFlagSet returns a flag set for the given subcommand that reports parse
// errors and prints its usage text to stderr, as the flag package does by
// default for the command line.
func newFlagSet(name string, stderr io.Writer, synopsis string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "%s\n\nFlags:\n", synopsis)
		fs.PrintDefaults()
	}
	return fs
}
