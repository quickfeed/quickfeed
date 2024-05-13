package sh_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/kit/sh"
)

func TestRun(t *testing.T) {
	if err := sh.Run("cat doesnotexist.txt"); err == nil {
		t.Error("expected: exit status 1")
	}
}

func TestOutput(t *testing.T) {
	s, err := sh.Output("ls -la")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(s)
}

func TestLintAG(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		// since we don't have golangci-lint installed in the GitHub Actions runner
		t.Skip("Skipping test since it is running in GitHub Actions")
	}

	// check formatting using goimports
	s, err := sh.Output("golangci-lint run --tests=false --disable-all --enable goimports")
	if err != nil {
		t.Error(err)
	}
	if s != "" {
		t.Error("Formatting check failed: goimports")
	}

	// check for TODO/BUG/FIXME comments using godox
	s, err = sh.Output("golangci-lint run --tests=false --disable-all --enable godox")
	if err == nil {
		t.Error("Expected golangci-lint to return error with 'exit status 1'")
	}
	if s == "" {
		t.Error("Expected golangci-lint to return message 'TODO(meling) test golangci-lint godox check for TODO items'")
	}

	// check many things: probably too aggressive for DAT320
	s, err = sh.Output("golangci-lint run --tests=true --disable structcheck --disable unused --disable deadcode --disable varcheck")
	if err != nil {
		t.Error(err)
	}
	if s != "" {
		fmt.Println(s)
	}
}

func TestRunRaceTest(t *testing.T) {
	tests := []struct {
		testName         string
		expectedRace     bool
		expectedOutput   string
		unexpectedOutput string
	}{
		{
			testName:         "TestWithDataRace",
			expectedRace:     true,
			expectedOutput:   "WARNING: DATA RACE",
			unexpectedOutput: "PASS",
		},
		{
			testName:         "TestWithoutDataRace",
			expectedRace:     false,
			expectedOutput:   "PASS",
			unexpectedOutput: "WARNING: DATA RACE",
		},
		{
			testName:         "TestThatDoesNotExist",
			expectedRace:     false,
			expectedOutput:   "warning: no tests to run",
			unexpectedOutput: "WARNING: DATA RACE",
		},
	}

	// unexpected returns Unexpected if race is true, otherwise Expected.
	// This is used to make the test output more readable.
	// It should only be used together with race != expectedRace.
	unexpected := func(race bool) string {
		if race {
			return "Unexpected"
		}
		return "Expected"
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// The tags argument is the empty string; we are currently not testing it.
			output, race := sh.RunRaceTest(tt.testName, "")
			if race != tt.expectedRace {
				t.Errorf("%s data race warning from %s", unexpected(race), tt.testName)
			}
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output with '%s' from %s", tt.expectedOutput, tt.testName)
				t.Log(output)
			}
			if strings.Contains(output, tt.unexpectedOutput) {
				t.Errorf("Unexpected output with '%s' from %s", tt.unexpectedOutput, tt.testName)
				t.Log(output)
			}
		})
	}
}
