package sh

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/kit/internal/test"
)

func TestRunRaceTest(t *testing.T) {
	tests := []struct {
		testFn           func(*testing.T)
		expectedRace     bool
		expectedOutput   string
		unexpectedOutput string
	}{
		{
			testFn:           TestWithDataRace,
			expectedRace:     true,
			expectedOutput:   "WARNING: DATA RACE",
			unexpectedOutput: "PASS",
		},
		{
			testFn:           TestWithoutDataRace,
			expectedRace:     false,
			expectedOutput:   "PASS",
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
		testName := test.Name(tt.testFn)
		t.Run(testName, func(t *testing.T) {
			// The tags argument is the empty string; we are currently not testing it.
			output, race := RunRaceTest(tt.testFn, "")
			if race != tt.expectedRace {
				t.Errorf("%s data race warning from %s", unexpected(race), testName)
			}
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output with '%s' from %s", tt.expectedOutput, testName)
				t.Log(output)
			}
			if strings.Contains(output, tt.unexpectedOutput) {
				t.Errorf("Unexpected output with '%s' from %s", tt.unexpectedOutput, testName)
				t.Log(output)
			}
		})
	}
}

// The following tests are not meant to test the kit package for data races.
// They are meant to test the RunRaceTest function, triggered by the TestRunRaceTest above.

// TestWithDataRace is intended to fail with a data race if the data race detector is enabled.
func TestWithDataRace(t *testing.T) {
	if !RaceEnabled {
		t.Skip(SkipMessage)
	}
	c := make(chan bool)
	m := make(map[string]string)
	go func() {
		m["1"] = "a" // First conflicting access.
		c <- true
	}()
	m["2"] = "b" // Second conflicting access.
	<-c
	for k, v := range m {
		fmt.Println(k, v)
	}
}

// TestWithoutDataRace should not have a data race.
func TestWithoutDataRace(t *testing.T) {
	if !RaceEnabled {
		t.Skip(SkipMessage)
	}
	c := make(chan bool)
	m := make(map[string]string)
	go func() {
		m["1"] = "a" // First access.
		c <- true
	}()
	<-c          // Wait for goroutine to finish.
	m["2"] = "b" // Second access.
	for k, v := range m {
		fmt.Println(k, v)
	}
}
