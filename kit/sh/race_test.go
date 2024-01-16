//go:build race

package sh

import (
	"fmt"
	"testing"
)

// TestWithDataRace is intended to fail with a data race if the data race detector is enabled.
func TestWithDataRace(t *testing.T) {
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
