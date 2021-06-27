package ci

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/log"
)

const (
	// qf101 is a test course for experimenting with go.sh behavior.
	// The test below will run locally on the test machine, not on the QuickFeed machine.
	getURL  = "https://github.com/qf101/meling-labs.git"
	testURL = "https://github.com/qf101/tests.git"
)

func TestRunTests(t *testing.T) {
	// The access token is a 'personal access token' for the user that has access to the repos below.
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' and access to the 'autograder-test' GitHub organization")
	}

	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		t.Fatal(err)
	}
	randomString := fmt.Sprintf("%x", sha1.Sum(randomness))
	info := &AssignmentInfo{
		AssignmentName:     "lab1",
		Script:             "go.sh",
		CreatorAccessToken: accessToken,
		GetURL:             getURL,
		TestURL:            testURL,
		RandomSecret:       randomString,
	}
	runData := &RunData{
		Course: &pb.Course{Code: "DAT320"},
		Assignment: &pb.Assignment{
			Name:             info.AssignmentName,
			ContainerTimeout: 1,
		},
		Repo:     &pb.Repository{},
		JobOwner: "muggles",
	}

	runner, err := NewDockerCI(log.Zap(true))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	ed, err := runTests("scripts", runner, info, runData)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%s\nExecTime: %v\nSecret: %v\n", ed.out, ed.execTime, info.RandomSecret)
}
