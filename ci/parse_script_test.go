package ci

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
)

func testRunData(qfTestOrg string) *RunData {
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	runData := &RunData{
		Course: &qf.Course{
			ID:                  1,
			Code:                "DAT320",
			ScmOrganizationName: qfTestOrg,
		},
		Assignment: &qf.Assignment{
			Name: "lab3",
		},
		Repo: &qf.Repository{
			HTMLURL:  repo.StudentRepoURL("user"),
			RepoType: qf.Repository_USER,
		},
		JobOwner: "muggles",
		CommitID: "deadbeef",
	}
	return runData
}

func TestLoadRunScript(t *testing.T) {
	t.Setenv("QUICKFEED_REPOSITORY_PATH", env.TestdataPath())
	runData := &RunData{
		Course: &qf.Course{
			ID:                  1,
			Code:                "qf104",
			ScmOrganizationName: "qf104-2022",
		},
		Assignment: &qf.Assignment{
			Name: "lab1",
		},
	}
	runSh, err := runData.loadRunScript(defaultRunScript)
	if err != nil {
		t.Error(err)
	}
	if len(runSh) == 0 {
		t.Error("run script is empty")
	}
	runData.Assignment = &qf.Assignment{Name: "lab2"}
	runSh, err = runData.loadRunScript(defaultRunScript)
	if err != nil {
		t.Error(err)
	}
	if len(runSh) == 0 {
		t.Error("run script is empty")
	}
}

func TestParseTestRunnerScript(t *testing.T) {
	t.Setenv("QUICKFEED_REPOSITORY_PATH", env.TestdataPath())

	const (
		qfTestOrg        = "qf104-2022"
		image            = "quickfeed:go"
		runScriptContent = `#image/quickfeed:go
echo "$TESTS"
echo "$ASSIGNMENTS"
echo "$SUBMITTED"
echo "$CURRENT"
echo "$QUICKFEED_SESSION_SECRET"
`
	)
	randomSecret := rand.String()

	runData := testRunData(qfTestOrg)
	job, err := runData.parseTestRunnerScript(randomSecret, "", nil)
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
		"ASSIGNMENTS=" + filepath.Join(QuickFeedPath, qf.AssignmentsRepo),
		"SUBMITTED=" + filepath.Join(QuickFeedPath, qf.StudentRepoName("user")),
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
	if job.Name != "dat320-lab3-muggles-"+runData.CommitID[:6] {
		t.Errorf("job.Name = %s, want %s", job.Name, "dat320-lab3-muggles-"+runData.CommitID[:6])
	}
}

func TestParseBadTestRunnerScript(t *testing.T) {
	t.Setenv("QUICKFEED_REPOSITORY_PATH", env.TestdataPath())

	const qfTestOrg = "qf104-2022"
	randomSecret := rand.String()

	runData := testRunData(qfTestOrg)
	runData.Assignment = &qf.Assignment{Name: "lab4-bad-run-script"}
	job, err := runData.parseTestRunnerScript(randomSecret, "", nil)
	if err == nil {
		t.Fatalf("expected error, got nil: %+v", job)
	}
	const wantMsg = "failed to parse run script for assignment lab4-bad-run-script in https://github.com/qf104-2022/tests: empty run script"
	if err.Error() != wantMsg {
		t.Errorf("err = '%s', want '%s'", err, wantMsg)
	}

	runData.Assignment = &qf.Assignment{Name: "lab5-bad-run-script"}
	job, err = runData.parseTestRunnerScript(randomSecret, "", nil)
	if err == nil {
		t.Fatalf("expected error, got nil: %+v", job)
	}
	const wantMsg2 = "failed to parse run script for assignment lab5-bad-run-script in https://github.com/qf104-2022/tests: no docker image specified in run script"
	if err.Error() != wantMsg2 {
		t.Errorf("err = '%s', want '%s'", err, wantMsg2)
	}
}
