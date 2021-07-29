package ci

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

const (
	gh = "github.com"
)

// To run this test, please see instructions in the developer guide (dev.md).

// This test uses a test course for experimenting with go.sh behavior.
// The test below will run locally on the test machine, not on the QuickFeed machine.

func TestRunTests(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	// Only used to fetch the user's GitHub login (user name)
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	userName, err := s.GetUserName(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	randomness := make([]byte, 10)
	if _, err := rand.Read(randomness); err != nil {
		t.Fatal(err)
	}
	randomString := fmt.Sprintf("%x", sha1.Sum(randomness))

	repo := pb.RepoURL{ProviderURL: gh, Organization: qfTestOrg}
	info := &AssignmentInfo{
		AssignmentName:     "lab1",
		Script:             "go.sh",
		CreatorAccessToken: accessToken,
		GetURL:             repo.StudentRepoURL(userName),
		TestURL:            repo.TestsRepoURL(),
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
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	t.Logf("\n%s\nExecTime: %v\nSecret: %v\n", ed.out, ed.execTime, info.RandomSecret)
}
