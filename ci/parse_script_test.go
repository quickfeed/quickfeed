package ci

import (
	"fmt"
	"os"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
)

// To run this test, please see instructions in the developer guide (dev.md).

// This test is meant for debugging template parsing of script files, such as go.sh and python.sh.
// We don't actually test anything here since the output is expected to be unstable; must be inspected manually.

func TestParseScript(t *testing.T) {
	const (
		// these are only used in text; no access to qf101 organization or user is needed
		qfTestOrg      = "qf101"
		githubUserName = "user"
	)
	randomString := qtest.RandomString(t)

	repo := pb.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	info := &AssignmentInfo{
		AssignmentName:     "lab2",
		Script:             "#image/qf101\n A script",
		CreatorAccessToken: "secret",
		GetURL:             repo.StudentRepoURL(githubUserName),
		TestURL:            repo.TestsRepoURL(),
		RandomSecret:       randomString,
	}
	j, err := parseScriptTemplate(info)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("TEST_TMPL") != "" {
		for _, cmd := range j.Commands {
			fmt.Println(cmd)
		}
	}
	if os.Getenv("TEST_IMAGE") != "" {
		fmt.Println(j.Image)
	}
	if os.Getenv("TEST_TMPL") != "" {
		for _, cmd := range j.Commands {
			fmt.Println(cmd)
		}
	}
	if os.Getenv("TEST_IMAGE") != "" {
		fmt.Println(j.Image)
	}
}
