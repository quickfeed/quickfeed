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

// TestSubstringMatchPrevention tests that RunRaceTest correctly anchors the test name
// pattern to prevent substring matching that could cause infinite recursion.
// This test verifies the fix for the issue where TestConcurrentWork and
// TestConcurrentWorkRace would both run when only TestConcurrentWork was intended.
func TestSubstringMatchPrevention(t *testing.T) {
	// Run the target test that should be executed
	output, race := RunRaceTest(TestConcurrentWork, "")

	// Verify that the test ran
	if !strings.Contains(output, "TestConcurrentWork") {
		t.Error("Expected TestConcurrentWork to run")
		t.Log(output)
	}

	// Verify that the wrapper test did NOT run (no infinite recursion)
	// The wrapper test name is TestConcurrentWorkRace which contains TestConcurrentWork as substring
	if strings.Contains(output, "TestConcurrentWorkRace") {
		t.Error("TestConcurrentWorkRace should not have been matched by the anchored regex")
		t.Log(output)
	}

	// The test should not detect a race (it doesn't have one)
	if race {
		t.Errorf("TestConcurrentWork should not have a data race:\n%s", output)
	}

	// Verify the test passed
	if !strings.Contains(output, "PASS") {
		t.Error("Expected TestConcurrentWork to pass")
		t.Log(output)
	}
}

// TestConcurrentWork is a target test for testing substring match prevention.
// This test simulates a concurrent workload but without any race conditions.
func TestConcurrentWork(t *testing.T) {
	if !RaceEnabled {
		t.Skip(SkipMessage)
	}
	// Simple test that should pass without races
	c := make(chan int)
	go func() {
		c <- 42
	}()
	result := <-c
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestConcurrentWorkRace is a wrapper test that would cause infinite recursion
// if the regex pattern in RunRaceTest is not anchored properly, because its name
// contains TestConcurrentWork as a substring.
func TestConcurrentWorkRace(t *testing.T) {
	// This test intentionally has a name that contains "TestConcurrentWork" as a substring
	// to verify that the anchored regex in RunRaceTest prevents it from being matched
	// when we only want to run TestConcurrentWork.
	output, race := RunRaceTest(TestConcurrentWork, "")
	if race {
		t.Errorf("Race detected:\n%s", output)
	}
}
