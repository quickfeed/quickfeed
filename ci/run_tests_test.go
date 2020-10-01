package ci

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
)

const (
	//TODO(meling) these repos should be replaced with a test course's repos, preferably public repos; so we don't need a token.
	getURL  = "https://github.com/dat320-2020/meling-stud-labs.git"
	testURL = "https://github.com/dat320-2020/tests.git"
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
		AssignmentName:     "lab4",
		Script:             "go.sh",
		CreatorAccessToken: accessToken,
		GetURL:             getURL,
		TestURL:            testURL,
		RawGetURL:          rawURL(getURL),
		RawTestURL:         rawURL(testURL),
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

	runner, err := NewDockerCI()
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
