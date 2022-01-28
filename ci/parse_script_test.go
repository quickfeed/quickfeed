package ci

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/rand"
)

// Testdata copied from run_tests_test.go (since they are in different packages)
func testRunData(qfTestOrg, userName, accessToken, scriptTemplate string) *RunData {
	repo := pb.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	courseID := uint64(1)
	pb.SetAccessToken(courseID, accessToken)
	runData := &RunData{
		Course: &pb.Course{
			ID:   courseID,
			Code: "DAT320",
		},
		Assignment: &pb.Assignment{
			Name:             "lab1",
			ScriptFile:       scriptTemplate,
			ContainerTimeout: 1,
		},
		Repo: &pb.Repository{
			HTMLURL:  repo.StudentRepoURL(userName),
			RepoType: pb.Repository_USER,
		},
		JobOwner: "muggles",
		CommitID: "deadbeef",
	}
	return runData
}

func TestParseScriptTemplate(t *testing.T) {
	const (
		// these are only used in text; no access to qf101 organization or user is needed
		qfTestOrg      = "qf101"
		image          = "quickfeed:go"
		scriptTemplate = `#image/quickfeed:go
AssignmentName: {{ .AssignmentName }}
RandomSecret: {{ .RandomSecret }}
`
		githubUserName = "user"
		accessToken    = "open sesame"
	)
	randomSecret := rand.String()

	runData := testRunData(qfTestOrg, githubUserName, "access_token", scriptTemplate)
	job, err := runData.parseScriptTemplate(randomSecret)
	if err != nil {
		t.Fatal(err)
	}
	if job.Image != image {
		t.Errorf("job.Image = %s, want %s", job.Image, image)
	}
	if job.Commands[0] != "AssignmentName: lab1" {
		t.Errorf("job.Commands[0] = %s, want %s", job.Commands[0], "AssignmentName: lab1")
	}
	if job.Commands[1] != "RandomSecret: "+randomSecret {
		t.Errorf("job.Commands[1] = %s, want %s", job.Commands[1], "RandomSecret: "+randomSecret)
	}
	if job.Name != "dat320-lab1-muggles-"+runData.CommitID[:6] {
		t.Errorf("job.Name = %s, want %s", job.Name, "dat320-lab1-muggles-"+runData.CommitID[:6])
	}
}

func TestParseBadScriptTemplate(t *testing.T) {
	const (
		// these are only used in text; no access to qf101 organization or user is needed
		qfTestOrg      = "qf101"
		image          = "quickfeed:go"
		githubUserName = "user"
		accessToken    = "open sesame"
	)
	randomSecret := rand.String()

	const scriptTemplate = `#image/quickfeed:go`
	runData := testRunData(qfTestOrg, githubUserName, "access_token", scriptTemplate)
	_, err := runData.parseScriptTemplate(randomSecret)
	const wantMsg = "no script template for assignment lab1 in https://github.com/qf101/tests"
	if err.Error() != wantMsg {
		t.Errorf("err = '%s', want '%s'", err, wantMsg)
	}

	const scriptTemplate2 = `
start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

`
	runData = testRunData(qfTestOrg, githubUserName, "access_token", scriptTemplate2)
	_, err = runData.parseScriptTemplate(randomSecret)
	const wantMsg2 = "no docker image specified in script template for assignment lab1 in https://github.com/qf101/tests"
	if err.Error() != wantMsg2 {
		t.Errorf("err = '%s', want '%s'", err, wantMsg2)
	}
}
