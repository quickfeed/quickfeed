package ci

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// AssignmentInfo holds metadata needed to fetch student code
// and the test repository for an assignment.
type AssignmentInfo struct {
	AssignmentName     string
	BranchName         string
	CreatorAccessToken string
	GetURL             string
	TestURL            string
	RandomSecret       string
}

// parseScriptTemplate returns a job specifying the docker image and commands
// to be executed by the docker image. The job's commands are extracted from
// the script template file (run.sh) associated with the RunData's assignment.
// The script template may use the variables of the AssignmentInfo struct, e.g.,
// {{ .AssignmentName }}, {{ .RandomSecret }}, etc.
func (r RunData) parseScriptTemplate(secret string) (*Job, error) {
	info := &AssignmentInfo{
		AssignmentName: r.Assignment.GetName(),
		// TODO(Meling): I have added BranchName here, since we need to checkout the corresponding branch when necessary.
		// To accommodate this, script files will now have to include the following:
		//
		// if [ {{ .BranchName }} != "main" ]; then
		//	git checkout {{ .BranchName }}
		// fi
		//
		// I do not know if this is sufficient. I have tested pushing from both the default branch and a feature branch,
		// and both worked. I also tried checking out a non-existant branch in the script, which also seemed to work, as it just stayed on the default branch.
		// We could force it to check against the default branch, but that would mean adding another field to AssignmentInfo.
		BranchName:         r.BranchName,
		CreatorAccessToken: r.Course.GetAccessToken(),
		GetURL:             r.Repo.GetHTMLURL(),
		TestURL:            r.Repo.GetTestURL(),
		RandomSecret:       secret,
	}
	// TODO(meling) rename ScriptFile field to ScriptTemplate
	// ScriptTemplate contains the script itself with variables in double curly braces
	t, err := template.New("script_template").Parse(r.Assignment.ScriptFile)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	if err := t.Execute(buffer, info); err != nil {
		return nil, err
	}
	s := strings.Split(buffer.String(), "\n")
	if len(s) < 2 {
		return nil, fmt.Errorf("no script template for assignment %s in %s", info.AssignmentName, info.TestURL)
	}
	parts := strings.Split(s[0], "#image/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("no docker image specified in script template for assignment %s in %s", info.AssignmentName, info.TestURL)
	}
	return &Job{Name: r.String(), Image: parts[1], Commands: s[1:]}, nil
}
