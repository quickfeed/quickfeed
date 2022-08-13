package ci

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
)

// Testdata copied from run_tests_test.go (since they are in different packages)
func testRunData(qfTestOrg, userName, accessToken, runScriptContent string) *RunData {
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	courseID := uint64(1)
	qf.SetAccessToken(courseID, accessToken)
	runData := &RunData{
		Course: &qf.Course{
			ID:   courseID,
			Code: "DAT320",
		},
		Assignment: &qf.Assignment{
			Name:             "lab1",
			RunScriptContent: runScriptContent,
			ContainerTimeout: 1,
		},
		Repo: &qf.Repository{
			HTMLURL:  repo.StudentRepoURL(userName),
			RepoType: qf.Repository_USER,
		},
		JobOwner: "muggles",
		CommitID: "deadbeef",
	}
	return runData
}

func TestParseTestRunnerScript(t *testing.T) {
	const (
		// these are only used in text; no access to qf101 organization or user is needed
		qfTestOrg        = "qf101"
		image            = "quickfeed:go"
		runScriptContent = `#image/quickfeed:go
echo $TESTS
echo $ASSIGNMENTS
echo $CURRENT
echo $QUICKFEED_SESSION_SECRET
`
		githubUserName = "user"
		accessToken    = "open sesame"
	)
	randomSecret := rand.String()

	runData := testRunData(qfTestOrg, githubUserName, "access_token", runScriptContent)
	job, err := runData.parseTestRunnerScript(randomSecret)
	if err != nil {
		t.Fatal(err)
	}
	if job.Image != image {
		t.Errorf("job.Image = %s, want %s", job.Image, image)
	}
	gotVars := job.Env
	wantVars := []string{
		"HOME=" + QuickFeedPath,
		"TESTS=" + filepath.Join(QuickFeedPath, qf.TestsRepo),
		"ASSIGNMENTS=" + filepath.Join(QuickFeedPath, qf.AssignmentRepo),
		"CURRENT=" + runData.Assignment.GetName(),
		"QUICKFEED_SESSION_SECRET=" + randomSecret,
	}
	trans := cmp.Transformer("Sort", func(in []string) []string {
		out := append([]string(nil), in...)
		sort.Strings(out)
		return out
	})
	if diff := cmp.Diff(wantVars, gotVars, trans); diff != "" {
		t.Errorf("parseTestRunnerScript() mismatch (-want +got):\n%s", diff)
	}
	_, after, found := strings.Cut(runScriptContent, image+"\n")
	if !found {
		t.Errorf("No script content found for image: %s", image)
	}
	for i, line := range strings.Split(after, "\n") {
		if line != job.Commands[i] {
			t.Errorf("job.Commands[%d] = %s, want %s", i, job.Commands[i], line)
		}
	}
	if job.Name != "dat320-lab1-muggles-"+runData.CommitID[:6] {
		t.Errorf("job.Name = %s, want %s", job.Name, "dat320-lab1-muggles-"+runData.CommitID[:6])
	}
}

func TestParseBadTestRunnerScript(t *testing.T) {
	const (
		// these are only used in text; no access to qf101 organization or user is needed
		qfTestOrg      = "qf101"
		image          = "quickfeed:go"
		githubUserName = "user"
		accessToken    = "open sesame"
	)
	randomSecret := rand.String()

	const runScriptContent = `#image/quickfeed:go`
	runData := testRunData(qfTestOrg, githubUserName, "access_token", runScriptContent)
	_, err := runData.parseTestRunnerScript(randomSecret)
	const wantMsg = "no run script for assignment lab1 in https://github.com/qf101/tests"
	if err.Error() != wantMsg {
		t.Errorf("err = '%s', want '%s'", err, wantMsg)
	}

	const runScriptContent2 = `
start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

`
	runData = testRunData(qfTestOrg, githubUserName, "access_token", runScriptContent2)
	_, err = runData.parseTestRunnerScript(randomSecret)
	const wantMsg2 = "no docker image specified in run script for assignment lab1 in https://github.com/qf101/tests"
	if err.Error() != wantMsg2 {
		t.Errorf("err = '%s', want '%s'", err, wantMsg2)
	}
}
